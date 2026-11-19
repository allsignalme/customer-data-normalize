package meta

import (
	"encoding/json"
	"os"
	"testing"
)

type metaVector struct {
	Field      string  `json:"field"`
	Input      string  `json:"input"`
	Country    *string `json:"country"`
	Normalized *string `json:"normalized"`
	SHA256     *string `json:"sha256"`
	Note       string  `json:"note"`
}

func loadVectors(t *testing.T) []metaVector {
	t.Helper()
	data, err := os.ReadFile("../vectors/meta-v1.json")
	if err != nil {
		t.Fatalf("read vectors: %v", err)
	}
	var file struct {
		Spec    string       `json:"spec"`
		Version int          `json:"version"`
		Vectors []metaVector `json:"vectors"`
	}
	if err := json.Unmarshal(data, &file); err != nil {
		t.Fatalf("parse vectors: %v", err)
	}
	if file.Spec != "meta" || file.Version != 1 || len(file.Vectors) == 0 {
		t.Fatalf("unexpected vector file: spec=%q version=%d vectors=%d", file.Spec, file.Version, len(file.Vectors))
	}
	return file.Vectors
}

func optionsFor(v metaVector) []Option {
	if v.Country == nil {
		return nil
	}
	return []Option{WithCountry(*v.Country)}
}

func TestNormalize_MetaV1Vectors_MatchExpected(t *testing.T) {
	for _, v := range loadVectors(t) {
		got, ok := Normalize(Field(v.Field), v.Input, optionsFor(v)...)
		switch {
		case v.Normalized == nil && ok:
			t.Errorf("%s %q (%s): want dropped, got %q", v.Field, v.Input, v.Note, got)
		case v.Normalized != nil && (!ok || got != *v.Normalized):
			t.Errorf("%s %q (%s): want %q, got %q ok=%v", v.Field, v.Input, v.Note, *v.Normalized, got, ok)
		}
	}
}

func TestHash_MetaV1Vectors_MatchExpected(t *testing.T) {
	for _, v := range loadVectors(t) {
		got, ok := Hash(Field(v.Field), v.Input, optionsFor(v)...)
		switch {
		case v.SHA256 == nil && ok:
			t.Errorf("%s %q (%s): want dropped, got %q", v.Field, v.Input, v.Note, got)
		case v.SHA256 != nil && (!ok || got != *v.SHA256):
			t.Errorf("%s %q (%s): want %q, got %q ok=%v", v.Field, v.Input, v.Note, *v.SHA256, got, ok)
		}
	}
}
