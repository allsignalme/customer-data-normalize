// Package phone converts phone numbers taken from an address to international format (spec phone-international v1).
package phone

import (
	"strings"

	"github.com/allsignalme/customer-data-normalize/internal/jsstr"
)

// Calling codes by ISO 3166-1 alpha-2 country, grouped by national trunk prefix.
var (
	trunk0Codes = map[string]string{
		"VN": "84", "TH": "66", "ID": "62", "MY": "60", "PH": "63", "JP": "81", "KR": "82", "CN": "86", "TW": "886",
		"IN": "91", "PK": "92", "BD": "880", "LK": "94", "AU": "61", "NZ": "64", "GB": "44", "IE": "353", "FR": "33",
		"DE": "49", "NL": "31", "BE": "32", "AT": "43", "CH": "41", "SE": "46", "FI": "358", "TR": "90", "IL": "972",
		"AE": "971", "SA": "966", "EG": "20", "ZA": "27", "NG": "234", "KE": "254", "UA": "380",
	}
	trunk8Codes   = map[string]string{"RU": "7", "KZ": "7"}
	nanpCountries = map[string]bool{"US": true, "CA": true, "PR": true}
)

// International converts a phone number from an address to international format using the country of that same
// address (ISO 3166-1 alpha-2, any case). It only converts when certain: the number carries the national trunk
// prefix of that country (0, 8, or North American numbering). A number that already starts with + or 00 is returned
// unchanged (trimmed). Anything else returns "" — better no phone than a hash of a wrong number.
func International(raw, country string) string {
	phone := jsstr.Trim(raw)
	if phone == "" {
		return ""
	}
	if strings.HasPrefix(phone, "+") || strings.HasPrefix(phone, "00") {
		return phone
	}
	digits := strings.Map(func(r rune) rune {
		if jsstr.IsSpace(r) || r == '(' || r == ')' || r == '.' || r == '-' {
			return -1
		}
		return r
	}, phone)
	cc := strings.ToUpper(jsstr.Trim(country))
	if digits == "" || cc == "" || strings.ContainsFunc(digits, func(r rune) bool { return r < '0' || r > '9' }) {
		return ""
	}
	if nanpCountries[cc] {
		switch {
		case len(digits) == 10:
			return "+1" + digits
		case len(digits) == 11 && digits[0] == '1':
			return "+" + digits
		}
		return ""
	}
	if code, ok := trunk0Codes[cc]; ok {
		return withTrunk(digits, '0', code)
	}
	if code, ok := trunk8Codes[cc]; ok {
		return withTrunk(digits, '8', code)
	}
	return ""
}

func withTrunk(digits string, trunk byte, code string) string {
	if len(digits) > 1 && digits[0] == trunk {
		return "+" + code + digits[1:]
	}
	return ""
}
