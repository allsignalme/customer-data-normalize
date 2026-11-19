// Package openai normalizes and hashes identifiers for the OpenAI Conversions API (spec openai v1, see SPEC.md).
// Normalization matches the OpenAI browser SDK (oaiq) so browser and server produce the same value.
package openai

import (
	"regexp"
	"strings"

	"github.com/allsignalme/customer-data-normalize/internal/jsstr"
	"github.com/allsignalme/customer-data-normalize/phone"
)

// Identifier kinds. Email, phone, names and external id are hashed; geo kinds are sent in clear.
const (
	IdentEmail      = "email"
	IdentPhone      = "phone"
	IdentFirstName  = "first_name"
	IdentLastName   = "last_name"
	IdentExternalID = "external_id"
	IdentCity       = "city"
	IdentRegion     = "region"
	IdentPostalCode = "postal_code"
	IdentCountry    = "country"
)

// maxIdentifiersPerField: the API only uses the first 3 valid values of each array.
const maxIdentifiersPerField = 3

// Length limits of the browser SDK.
const (
	maxEmailLength = 254
	maxNameLength  = 1024
	maxGeoLength   = 128
	maxPostalCode  = 31
)

// emailPattern is the SDK pattern. It runs on lowercased input without a case-insensitive flag: Go's (?i) would also
// accept U+017F and U+212A, which JavaScript's /i does not.
var emailPattern = regexp.MustCompile("^[\\w!#$%&'*+/=?^`{|}~-]+(?:\\.[\\w!#$%&'*+/=?^`{|}~-]+)*@(?:[a-z0-9](?:[a-z0-9-]*[a-z0-9])?\\.)+[a-z0-9](?:[a-z0-9-]*[a-z0-9])?$")

const asciiPunctuation = "!\"#$%&'()*+,-./:;<=>?@[\\]^_`{|}~"

// NormalizeEmail trims and lowercases; longer than 254 characters or not an email => "".
func NormalizeEmail(raw string) string {
	n := jsstr.ToLower(jsstr.Trim(raw))
	if len(n) > maxEmailLength || !emailPattern.MatchString(n) {
		return ""
	}
	return n
}

// NormalizePhone removes spaces, parentheses, dots and dashes; the rest must be digits (optionally a leading +),
// not "+0". Drops the + and all leading zeros. With an international prefix (+ or 00) a number starting with 1 must
// have exactly 11 digits (North America); otherwise 8–15 digits. It does NOT add a country code — use
// InternationalPhone first for address phones. Invalid => "".
func NormalizePhone(raw string) string {
	var b strings.Builder
	for _, r := range jsstr.Trim(raw) {
		if jsstr.IsSpace(r) || r == '(' || r == ')' || r == '.' || r == '-' {
			continue
		}
		b.WriteRune(r)
	}
	n := b.String()
	body := strings.TrimPrefix(n, "+")
	if body == "" || strings.HasPrefix(n, "+0") || strings.ContainsFunc(body, func(r rune) bool { return r < '0' || r > '9' }) {
		return ""
	}
	international := strings.HasPrefix(n, "+") || strings.HasPrefix(n, "00")
	digits := strings.TrimLeft(body, "0")
	if international && strings.HasPrefix(digits, "1") && len(digits) != 11 {
		return ""
	}
	if len(digits) < 8 || len(digits) > 15 {
		return ""
	}
	return digits
}

// InternationalPhone converts an address phone to international format using the address country (see package
// phone). Run it before NormalizePhone.
func InternationalPhone(raw, countryCode string) string { return phone.International(raw, countryCode) }

// NormalizeName: longer than 1024 characters => ""; removes ASCII punctuation and whitespace, lowercases; keeps
// non-ASCII characters without Unicode normalization; must keep at least one non-digit character.
func NormalizeName(raw string) string {
	if jsstr.UTF16Len(raw) > maxNameLength {
		return ""
	}
	stripped := strings.Map(func(r rune) rune {
		if jsstr.IsSpace(r) || (r < 0x80 && strings.ContainsRune(asciiPunctuation, r)) {
			return -1
		}
		return r
	}, raw)
	n := jsstr.ToLower(stripped)
	if !strings.ContainsFunc(n, func(r rune) bool { return r < '0' || r > '9' }) {
		return ""
	}
	return n
}

// NormalizeGeo (city, region): lowercase, collapse whitespace runs to one space, trim, keep accents; 1–128
// characters. Geo values are sent in clear, not hashed.
func NormalizeGeo(raw string) string {
	n := strings.Join(strings.FieldsFunc(jsstr.ToLower(raw), jsstr.IsSpace), " ")
	if n == "" || jsstr.UTF16Len(n) > maxGeoLength {
		return ""
	}
	return n
}

// NormalizeCountry: two-letter code, lowercase; anything else => "".
func NormalizeCountry(raw string) string {
	n := NormalizeGeo(raw)
	if len(n) != 2 || n[0] < 'a' || n[0] > 'z' || n[1] < 'a' || n[1] > 'z' {
		return ""
	}
	return n
}

// NormalizePostalCode: like NormalizeGeo, then only a-z, 0-9, space and dash, at most 31 characters.
func NormalizePostalCode(raw string) string {
	n := NormalizeGeo(raw)
	if len(n) > maxPostalCode || strings.ContainsFunc(n, func(r rune) bool {
		return !(r >= 'a' && r <= 'z' || r >= '0' && r <= '9' || r == ' ' || r == '-')
	}) {
		return ""
	}
	return n
}

// NormalizeExternalID trims only; case is kept.
func NormalizeExternalID(raw string) string { return jsstr.Trim(raw) }

// NormalizeByKind dispatches on kind; unknown kind => "".
func NormalizeByKind(kind, raw string) string {
	switch kind {
	case IdentEmail:
		return NormalizeEmail(raw)
	case IdentPhone:
		return NormalizePhone(raw)
	case IdentFirstName, IdentLastName:
		return NormalizeName(raw)
	case IdentExternalID:
		return NormalizeExternalID(raw)
	case IdentCity, IdentRegion:
		return NormalizeGeo(raw)
	case IdentPostalCode:
		return NormalizePostalCode(raw)
	case IdentCountry:
		return NormalizeCountry(raw)
	}
	return ""
}

// IsSHA256Hex reports whether s is 64 hex characters (either case).
func IsSHA256Hex(s string) bool { return jsstr.IsSHA256Hex(s) }

func isHashedKind(kind string) bool {
	switch kind {
	case IdentEmail, IdentPhone, IdentFirstName, IdentLastName, IdentExternalID:
		return true
	}
	return false
}

// HashIdentifier normalizes by kind, then returns the lowercase hex SHA-256. A value that is already 64 hex characters
// is returned lowercased, not hashed again. ok=false: empty, invalid after normalization, or not a hashed kind (geo
// kinds are sent in clear) — do not send it.
func HashIdentifier(kind, raw string) (string, bool) {
	if !isHashedKind(kind) {
		return "", false
	}
	trimmed := jsstr.Trim(raw)
	if trimmed == "" {
		return "", false
	}
	if jsstr.IsSHA256Hex(trimmed) {
		return strings.ToLower(trimmed), true
	}
	normalized := NormalizeByKind(kind, trimmed)
	if normalized == "" {
		return "", false
	}
	return jsstr.SHA256Hex(normalized), true
}

// AppendIdentifier hashes raw and appends it when valid, not already present, and the list has fewer than 3 values
// (order kept).
func AppendIdentifier(list []string, kind, raw string) []string {
	if len(list) >= maxIdentifiersPerField {
		return list
	}
	hashed, ok := HashIdentifier(kind, raw)
	if !ok {
		return list
	}
	for _, existing := range list {
		if existing == hashed {
			return list
		}
	}
	return append(list, hashed)
}
