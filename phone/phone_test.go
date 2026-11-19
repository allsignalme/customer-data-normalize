package phone

import (
	"encoding/json"
	"os"
	"testing"
)

func TestInternational_PhoneInternationalV1Vectors_MatchExpected(t *testing.T) {
	data, err := os.ReadFile("../vectors/phone-international-v1.json")
	if err != nil {
		t.Fatalf("read vectors: %v", err)
	}
	var file struct {
		Spec    string `json:"spec"`
		Version int    `json:"version"`
		Vectors []struct {
			Input    string  `json:"input"`
			Country  *string `json:"country"`
			Expected *string `json:"expected"`
			Note     string  `json:"note"`
		} `json:"vectors"`
	}
	if err := json.Unmarshal(data, &file); err != nil {
		t.Fatalf("parse vectors: %v", err)
	}
	if file.Spec != "phone-international" || file.Version != 1 || len(file.Vectors) == 0 {
		t.Fatalf("unexpected vector file: spec=%q version=%d", file.Spec, file.Version)
	}
	for _, v := range file.Vectors {
		country := ""
		if v.Country != nil {
			country = *v.Country
		}
		want := ""
		if v.Expected != nil {
			want = *v.Expected
		}
		if got := International(v.Input, country); got != want {
			t.Errorf("International(%q, %q) (%s) = %q, want %q", v.Input, country, v.Note, got, want)
		}
	}
}
