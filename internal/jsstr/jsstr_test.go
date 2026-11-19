package jsstr

import "testing"

func TestToLower_SpecialCases_MatchJavaScript(t *testing.T) {
	cases := []struct{ in, want string }{
		{"ABC", "abc"},
		{"İlker", "i\u0307lker"},
		{"ΟΔΥΣΣΕΑΣ", "οδυσσεας"},
		{"ΣΑΣ ΣΑΣ", "σας σας"},
		{"Σ", "σ"},
		{"ΑΣ'Α", "ασ'α"},
		{"\u212aelvin", "kelvin"},
		{"ſ", "ſ"},
		{"NGUYỄN", "nguyễn"},
	}
	for _, c := range cases {
		if got := ToLower(c.in); got != c.want {
			t.Errorf("ToLower(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestIsSpace_Runes_MatchJavaScript(t *testing.T) {
	for _, r := range []rune{'\t', '\n', '\v', '\f', '\r', ' ', '\u00a0', '\u1680', '\u2000', '\u200a', '\u2028',
		'\u2029', '\u202f', '\u205f', '\u3000', '\ufeff'} {
		if !IsSpace(r) {
			t.Errorf("IsSpace(%U) = false, want true", r)
		}
	}
	for _, r := range []rune{'\u0085', '\u200b', '\u180e', 'a', '0'} {
		if IsSpace(r) {
			t.Errorf("IsSpace(%U) = true, want false", r)
		}
	}
}

func TestUTF16Len_Astral_CountsSurrogates(t *testing.T) {
	if got := UTF16Len("a😀é"); got != 4 {
		t.Errorf("UTF16Len = %d, want 4", got)
	}
}
