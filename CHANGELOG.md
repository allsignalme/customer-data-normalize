# Changelog

## 0.1.0 — unreleased

First version. Implements spec meta v1, openai v1 and phone-international v1 (see SPEC.md) in Go and TypeScript.

- **meta v1**: normalization reproduced from meta/capi-param-builder (commit `08cf920`) with deviations D1–D4; vectors
  generated from that reference, including every hashed example of Meta's documentation.
- **openai v1** and **phone-international v1**: taken from the existing Go and TypeScript implementations and their
  shared vectors, with two Go-only fixes so Go always equals TypeScript:
  - trimming and whitespace use the JavaScript whitespace set (U+FEFF is whitespace, U+0085 is not);
  - lowercasing follows JavaScript (`İ` => `i̇`, word-final `Σ` => `ς`), and the email pattern no longer matches
    U+017F / U+212A through Go's case-insensitive flag.
