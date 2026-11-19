// Package meta normalizes and hashes customer information parameters for the Meta Conversions API (spec meta v1,
// see SPEC.md). Normalization reproduces Meta's reference implementation (meta/capi-param-builder) so the value
// hashed on the server equals the value Meta's own browser code would hash.
package meta

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/allsignalme/customer-data-normalize/internal/jsstr"
	"github.com/allsignalme/customer-data-normalize/phone"
)

// Field is a user_data key of the Meta Conversions API that must (or should) be hashed.
type Field string

// Hashed user_data keys.
const (
	Email       Field = "em"
	Phone       Field = "ph"
	FirstName   Field = "fn"
	LastName    Field = "ln"
	DateOfBirth Field = "db"
	Gender      Field = "ge"
	City        Field = "ct"
	State       Field = "st"
	Zip         Field = "zp"
	Country     Field = "country"
	ExternalID  Field = "external_id"
)

// Option adds context to a normalization.
type Option func(*config)

type config struct{ country string }

// WithCountry gives the ISO 3166-1 alpha-2 country of the customer (any case). It is used by Phone only: a national
// number with that country's trunk prefix gets the country calling code (see package phone). Other fields ignore it.
func WithCountry(code string) Option { return func(c *config) { c.country = code } }

// ShouldHash reports whether a user_data key is one of the hashed fields. Keys such as client_ip_address,
// client_user_agent, fbc and fbp must be sent in clear.
func ShouldHash(key string) bool {
	switch Field(key) {
	case Email, Phone, FirstName, LastName, DateOfBirth, Gender, City, State, Zip, Country, ExternalID:
		return true
	}
	return false
}

// IsSHA256Hex reports whether s is a SHA-256 hex digest (64 hex characters, either case).
func IsSHA256Hex(s string) bool { return jsstr.IsSHA256Hex(s) }

// Normalize returns the normalized value of raw for field f. A value that is already a SHA-256 hex digest (after
// trimming) is returned lowercased as is. ok=false means the value is empty, invalid or the field is unknown: do not
// send it.
func Normalize(f Field, raw string, opts ...Option) (string, bool) {
	if raw == "" || !ShouldHash(string(f)) {
		return "", false
	}
	if t := jsstr.Trim(raw); jsstr.IsSHA256Hex(t) {
		return strings.ToLower(t), true
	}
	var cfg config
	for _, o := range opts {
		o(&cfg)
	}
	var n string
	switch f {
	case Email:
		n = normalizeEmail(raw)
	case Phone:
		n = normalizePhone(raw, cfg.country)
	case FirstName, LastName:
		n = normalizeName(raw)
	case DateOfBirth:
		n = normalizeDateOfBirth(raw, time.Now())
	case Gender:
		n = normalizeGender(raw)
	case City:
		n = normalizeCity(raw)
	case State:
		n = normalizeAbbreviation(raw, stateNames, stateIndex)
	case Zip:
		n = normalizeZip(raw)
	case Country:
		n = normalizeAbbreviation(raw, countryNames, countryIndex)
	case ExternalID:
		n = jsstr.ToLower(jsstr.RemoveSpace(raw))
	}
	return n, n != ""
}

// Hash returns the lowercase hex SHA-256 of the normalized value. A value that is already a SHA-256 hex digest is
// returned lowercased, not hashed again. ok=false: do not send the field.
func Hash(f Field, raw string, opts ...Option) (string, bool) {
	n, ok := Normalize(f, raw, opts...)
	if !ok {
		return "", false
	}
	if jsstr.IsSHA256Hex(jsstr.Trim(raw)) {
		return n, true
	}
	return jsstr.SHA256Hex(n), true
}

// emailPattern is the reference pattern. It runs on lowercased input without a case-insensitive flag: Go's (?i)
// would also accept U+017F and U+212A, which JavaScript's /i does not.
var emailPattern = regexp.MustCompile("^[\\w!#$%&'*+/=?^`{|}~-]+(?:\\.[\\w!#$%&'*+/=?^`{|}~-]+)*@(?:[a-z0-9](?:[a-z0-9-]*[a-z0-9])?\\.)+[a-z0-9](?:[a-z0-9-]*[a-z0-9])?$")

func normalizeEmail(raw string) string {
	n := jsstr.Trim(jsstr.ToLower(raw))
	if !emailPattern.MatchString(n) {
		return ""
	}
	return n
}

// normalizePhone keeps digits and drops leading zeros. With a country, a national number is first converted to
// international format when that can be done safely (deviation D3).
func normalizePhone(raw, country string) string {
	if country != "" {
		if intl := phone.International(raw, country); intl != "" {
			raw = intl
		}
	}
	digits := strings.Map(func(r rune) rune {
		if r < '0' || r > '9' {
			return -1
		}
		return r
	}, raw)
	return strings.TrimLeft(digits, "0")
}

const asciiPunctuation = "!\"#$%&'()*+,-./:;<=>?@[\\]^_`{|}~"

// normalizeName removes whitespace and ASCII punctuation, then lowercases. Other characters (accents, Hangul,
// non-ASCII punctuation) are kept.
func normalizeName(raw string) string {
	stripped := strings.Map(func(r rune) rune {
		if jsstr.IsSpace(r) || (r < 0x80 && strings.ContainsRune(asciiPunctuation, r)) {
			return -1
		}
		return r
	}, raw)
	return jsstr.ToLower(stripped)
}

// normalizeDateOfBirth reads the first three digit groups as a date (year is the group > 31; month/day order is
// decided by which value is > 12, defaulting to month first), or a single 8-digit group as is. Years from 1800 to
// next year (UTC).
func normalizeDateOfBirth(raw string, now time.Time) string {
	groups := strings.FieldsFunc(raw, func(r rune) bool { return r < '0' || r > '9' })
	if len(groups) >= 3 {
		n1, n2, n3 := parseGroup(groups[0]), parseGroup(groups[1]), parseGroup(groups[2])
		var year, month, day int64
		switch {
		case n1 > 31:
			year = n1
			month, day = n2, n3
			if n2 > 12 {
				month, day = n3, n2
			}
		case n2 > 31:
			year = n2
			month, day = n3, n1
			if n3 > 12 {
				month, day = n1, n3
			}
		default:
			year = n3
			month, day = n1, n2
			if n1 > 12 {
				month, day = n2, n1
			}
		}
		if validDate(year, month, day, now) {
			return fmt.Sprintf("%04d%02d%02d", year, month, day)
		}
		return ""
	}
	if len(groups) == 1 && len(groups[0]) == 8 {
		return groups[0]
	}
	return ""
}

// parseGroup parses a digit group; anything too long to be a date part is mapped to a value that fails validation.
func parseGroup(s string) int64 {
	s = strings.TrimLeft(s, "0")
	if s == "" {
		return 0
	}
	if len(s) > 12 {
		return 1 << 50
	}
	n, _ := strconv.ParseInt(s, 10, 64)
	return n
}

func validDate(year, month, day int64, now time.Time) bool {
	if year < 1800 || year > int64(now.UTC().Year())+1 || month < 1 || month > 12 || day < 1 {
		return false
	}
	daysInMonth := time.Date(int(year), time.Month(month)+1, 0, 0, 0, 0, 0, time.UTC).Day()
	return day <= int64(daysInMonth)
}

var genderWords = func() map[string]string {
	m := make(map[string]string, len(maleWords)+len(femaleWords))
	for _, w := range maleWords {
		m[w] = "m"
	}
	for _, w := range femaleWords {
		m[w] = "f"
	}
	return m
}()

// normalizeGender maps a gender word (male, female, woman, m, f…) to "m" or "f". Unknown words are dropped.
func normalizeGender(raw string) string { return genderWords[jsstr.Trim(jsstr.ToLower(raw))] }

// normalizeCity keeps only ASCII letters and digits, lowercased; the result must start with a letter.
func normalizeCity(raw string) string {
	n := keepASCII(raw, true)
	if n == "" || n[0] < 'a' || n[0] > 'z' {
		return ""
	}
	return n
}

// keepASCII keeps a-z (lowercasing A-Z) and, if digits is true, 0-9.
func keepASCII(s string, digits bool) string {
	var b strings.Builder
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z', digits && r >= '0' && r <= '9':
			b.WriteRune(r)
		case r >= 'A' && r <= 'Z':
			b.WriteRune(r + ('a' - 'A'))
		}
	}
	return b.String()
}

// nameCode is a lookup row: name with letters a-z only, and its two-letter code.
type nameCode struct{ name, code string }

func indexOf(rows []nameCode) map[string]string {
	m := make(map[string]string, len(rows))
	for _, r := range rows {
		if _, dup := m[r.name]; !dup {
			m[r.name] = r.code
		}
	}
	return m
}

var (
	stateIndex   = indexOf(stateNames)
	countryIndex = indexOf(countryNames)
)

// normalizeAbbreviation (state, country): lowercase, keep a-z, then a two-letter value as is, a known name (exact,
// then the first name contained in the value, in table order) as its code, anything else cut to two letters.
func normalizeAbbreviation(raw string, rows []nameCode, index map[string]string) string {
	n := keepASCII(jsstr.Trim(jsstr.ToLower(raw)), false)
	if len(n) != 2 {
		if code, ok := index[n]; ok {
			n = code
		} else {
			for _, r := range rows {
				if strings.Contains(n, r.name) {
					n = r.code
					break
				}
			}
		}
	}
	if len(n) > 2 {
		n = n[:2]
	}
	return n
}

// normalizeZip lowercases, keeps the part before the first dash, trims; at least 2 characters.
func normalizeZip(raw string) string {
	n := jsstr.ToLower(raw)
	if i := strings.IndexByte(n, '-'); i >= 0 {
		n = n[:i]
	}
	n = jsstr.Trim(n)
	if jsstr.UTF16Len(n) < 2 {
		return ""
	}
	return n
}
