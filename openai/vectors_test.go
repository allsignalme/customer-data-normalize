package openai

import (
	"encoding/json"
	"os"
	"testing"
)

type openaiVector struct {
	Kind       string  `json:"kind"`
	Input      string  `json:"input"`
	Normalized *string `json:"normalized"`
	SHA256     *string `json:"sha256"`
	Note       string  `json:"note"`
}

func loadVectors(t *testing.T) []openaiVector {
	t.Helper()
	data, err := os.ReadFile("../vectors/openai-v1.json")
	if err != nil {
		t.Fatalf("read vectors: %v", err)
	}
	var file struct {
		Spec    string         `json:"spec"`
		Version int            `json:"version"`
		Vectors []openaiVector `json:"vectors"`
	}
	if err := json.Unmarshal(data, &file); err != nil {
		t.Fatalf("parse vectors: %v", err)
	}
	if file.Spec != "openai" || file.Version != 1 || len(file.Vectors) == 0 {
		t.Fatalf("unexpected vector file: spec=%q version=%d vectors=%d", file.Spec, file.Version, len(file.Vectors))
	}
	return file.Vectors
}

func TestNormalizeByKind_OpenAIV1Vectors_MatchExpected(t *testing.T) {
	for _, v := range loadVectors(t) {
		want := ""
		if v.Normalized != nil {
			want = *v.Normalized
		}
		if got := NormalizeByKind(v.Kind, v.Input); got != want {
			t.Errorf("%s %q (%s): want %q, got %q", v.Kind, v.Input, v.Note, want, got)
		}
	}
}

func TestHashIdentifier_OpenAIV1Vectors_MatchExpected(t *testing.T) {
	for _, v := range loadVectors(t) {
		got, ok := HashIdentifier(v.Kind, v.Input)
		switch {
		case v.SHA256 == nil && ok:
			t.Errorf("%s %q (%s): want dropped, got %q", v.Kind, v.Input, v.Note, got)
		case v.SHA256 != nil && (!ok || got != *v.SHA256):
			t.Errorf("%s %q (%s): want %q, got %q ok=%v", v.Kind, v.Input, v.Note, *v.SHA256, got, ok)
		}
	}
}

func TestAppendIdentifier_DuplicatesAndLimit_KeepsFirstThreeUnique(t *testing.T) {
	var list []string
	for _, raw := range []string{"a@example.com", "A@example.com ", "bad", "b@example.com", "c@example.com", "d@example.com"} {
		list = AppendIdentifier(list, IdentEmail, raw)
	}
	if len(list) != 3 {
		t.Fatalf("len = %d, want 3: %v", len(list), list)
	}
	first, _ := HashIdentifier(IdentEmail, "a@example.com")
	if list[0] != first {
		t.Errorf("list[0] = %q, want hash of a@example.com", list[0])
	}
}

func TestInternationalPhone_NationalNumber_AddsCountryCode(t *testing.T) {
	if got := InternationalPhone("0983 714 965", "VN"); got != "+84983714965" {
		t.Errorf("InternationalPhone = %q, want +84983714965", got)
	}
}
