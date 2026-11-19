package meta

import (
	"testing"
	"time"
)

func TestNormalize_UnknownField_Dropped(t *testing.T) {
	for _, f := range []Field{"", "email", "client_ip_address", "fbp"} {
		if got, ok := Normalize(f, "John@example.com"); ok {
			t.Errorf("field %q: want dropped, got %q", f, got)
		}
		if got, ok := Hash(f, "62a14e44f765419d10fea99367361a727c12365e2520f32218d505ed9aa0f62f"); ok {
			t.Errorf("field %q: pre-hashed input must still be dropped, got %q", f, got)
		}
	}
}

func TestShouldHash_KnownKeys_MatchMetaTable(t *testing.T) {
	hashed := []string{"em", "ph", "fn", "ln", "db", "ge", "ct", "st", "zp", "country", "external_id"}
	for _, k := range hashed {
		if !ShouldHash(k) {
			t.Errorf("ShouldHash(%q) = false, want true", k)
		}
	}
	notHashed := []string{"client_ip_address", "client_user_agent", "fbc", "fbp", "subscription_id", "fb_login_id",
		"lead_id", "anon_id", "madid", "page_id", "page_scoped_user_id", "ctwa_clid", "ig_account_id", "ig_sid", "", "EM"}
	for _, k := range notHashed {
		if ShouldHash(k) {
			t.Errorf("ShouldHash(%q) = true, want false", k)
		}
	}
}

func TestNormalizeDateOfBirth_YearBound_FollowsCurrentYear(t *testing.T) {
	now := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	cases := []struct {
		in, want string
	}{
		{"2027-01-01", "20270101"},
		{"2028-01-01", ""},
		{"1800-01-01", "18000101"},
	}
	for _, c := range cases {
		if got := normalizeDateOfBirth(c.in, now); got != c.want {
			t.Errorf("normalizeDateOfBirth(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestIsSHA256Hex_Inputs_DetectsDigests(t *testing.T) {
	cases := map[string]bool{
		"62a14e44f765419d10fea99367361a727c12365e2520f32218d505ed9aa0f62f": true,
		"62A14E44F765419D10FEA99367361A727C12365E2520F32218D505ED9AA0F62F": true,
		"62a14e44f765419d10fea99367361a727c12365e2520f32218d505ed9aa0f62":  false,
		"0123456789abcdef0123456789abcdef":                                 false,
		"g2a14e44f765419d10fea99367361a727c12365e2520f32218d505ed9aa0f62f": false,
	}
	for in, want := range cases {
		if got := IsSHA256Hex(in); got != want {
			t.Errorf("IsSHA256Hex(%q) = %v, want %v", in, got, want)
		}
	}
}
