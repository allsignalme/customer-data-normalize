// Package jsstr reproduces the JavaScript string behaviour the specs rely on (whitespace, trim, toLowerCase,
// UTF-16 length), so the Go and TypeScript implementations produce byte-identical output.
package jsstr

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"unicode"
	"unicode/utf16"
)

// IsSpace reports whether r matches JavaScript `\s` (WhiteSpace + LineTerminator). Unlike unicode.IsSpace it
// includes U+FEFF and excludes U+0085.
func IsSpace(r rune) bool {
	switch r {
	case '\t', '\n', '\v', '\f', '\r', ' ', '\u00a0', '\u1680', '\u2028', '\u2029', '\u202f', '\u205f', '\u3000', '\ufeff':
		return true
	}
	return r >= '\u2000' && r <= '\u200a'
}

// Trim is String.prototype.trim.
func Trim(s string) string { return strings.TrimFunc(s, IsSpace) }

// RemoveSpace removes every JavaScript whitespace character.
func RemoveSpace(s string) string {
	return strings.Map(func(r rune) rune {
		if IsSpace(r) {
			return -1
		}
		return r
	}, s)
}

// ToLower is String.prototype.toLowerCase: Unicode full lowercase mapping without language rules. It differs from
// strings.ToLower for U+0130 (becomes "i\u0307") and for capital sigma at the end of a word (becomes final sigma).
func ToLower(s string) string {
	ascii := true
	for i := 0; i < len(s); i++ {
		if s[i] >= 0x80 {
			ascii = false
			break
		}
	}
	if ascii {
		return strings.ToLower(s)
	}
	runes := []rune(s)
	var b strings.Builder
	b.Grow(len(s))
	for i, r := range runes {
		switch r {
		case '\u0130':
			b.WriteString("i\u0307")
		case '\u03a3':
			if isFinalSigma(runes, i) {
				b.WriteRune('\u03c2')
			} else {
				b.WriteRune('\u03c3')
			}
		default:
			b.WriteRune(unicode.ToLower(r))
		}
	}
	return b.String()
}

// isFinalSigma implements the Final_Sigma condition of Unicode SpecialCasing: preceded by a cased letter and not
// followed by one, ignoring case-ignorable characters in between.
func isFinalSigma(rs []rune, i int) bool {
	j := i - 1
	for j >= 0 && isCaseIgnorable(rs[j]) {
		j--
	}
	if j < 0 || !isCased(rs[j]) {
		return false
	}
	k := i + 1
	for k < len(rs) && isCaseIgnorable(rs[k]) {
		k++
	}
	return k >= len(rs) || !isCased(rs[k])
}

func isCased(r rune) bool {
	return unicode.In(r, unicode.Lu, unicode.Ll, unicode.Lt, unicode.Other_Lowercase, unicode.Other_Uppercase)
}

func isCaseIgnorable(r rune) bool {
	switch r { // Word_Break = MidLetter, MidNumLet or Single_Quote
	case '\'', '.', ':', '\u00b7', '\u0387', '\u055f', '\u05f4', '\u2018', '\u2019', '\u2024', '\u2027', '\ufe13',
		'\ufe52', '\ufe55', '\uff07', '\uff0e', '\uff1a':
		return true
	}
	return unicode.In(r, unicode.Mn, unicode.Me, unicode.Cf, unicode.Lm, unicode.Sk)
}

// UTF16Len is the JavaScript `.length` of s.
func UTF16Len(s string) int { return len(utf16.Encode([]rune(s))) }

// IsSHA256Hex reports whether s is exactly 64 hexadecimal characters (either case).
func IsSHA256Hex(s string) bool {
	if len(s) != sha256.Size*2 {
		return false
	}
	for i := 0; i < len(s); i++ {
		c := s[i]
		if !(c >= '0' && c <= '9' || c >= 'a' && c <= 'f' || c >= 'A' && c <= 'F') {
			return false
		}
	}
	return true
}

// SHA256Hex returns the lowercase hex SHA-256 of the UTF-8 bytes of s.
func SHA256Hex(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:])
}
