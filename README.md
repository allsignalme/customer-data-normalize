# customer-data-normalize

Normalization and SHA-256 hashing of customer information (email, phone, name, address, date of birth…) for the
**Meta** and **OpenAI** Conversions APIs, with identical results in **Go** and **TypeScript**.

Matching works only when every system that sends an identifier normalizes it the same way. This repository is the
single source for those rules: [SPEC.md](SPEC.md) describes them, [`vectors/`](vectors/) holds the expected outputs,
and both implementations are tested against the same vectors.

- **Meta (spec meta v1):** reproduces Meta's own normalization code
  ([meta/capi-param-builder](https://github.com/meta/capi-param-builder)), plus a safe country-code step for phone
  numbers. Every example with a published hash in Meta's documentation is a test vector.
- **OpenAI (spec openai v1):** matches the OpenAI browser SDK, so browser and server produce the same value.
- No runtime dependencies. Go 1.22+. TypeScript output is plain ES2020 modules with a pure-JS SHA-256, so it runs in
  Node, browsers and sandboxes without `crypto.subtle` (e.g. Shopify web pixels).

## Go

```bash
go get github.com/allsignalme/customer-data-normalize@latest
```

```go
import (
	"github.com/allsignalme/customer-data-normalize/meta"
	"github.com/allsignalme/customer-data-normalize/openai"
)

em, ok := meta.Hash(meta.Email, " John_Smith@gmail.com ")
// em = "62a14e44f765419d10fea99367361a727c12365e2520f32218d505ed9aa0f62f", ok = true

ph, ok := meta.Hash(meta.Phone, "0912 345 678", meta.WithCountry("VN"))
// hash of "84912345678"

n, ok := meta.Normalize(meta.City, "New York") // "newyork" — the value before hashing

if !meta.ShouldHash("fbp") { /* send in clear */ }

h, ok := openai.HashIdentifier(openai.IdentEmail, "  John.Doe@Example.COM ")
```

`ok == false` means the value is empty or invalid: do not send the field. A value that is already a SHA-256 digest is
returned lowercased, not hashed twice.

## TypeScript

Install from a tag of this repository (`dist/` is committed, no build step on install):

```bash
npm install github:allsignalme/customer-data-normalize#v0.1.0
```

```ts
import { hash, normalize, shouldHash } from "@allsignalme/customer-data-normalize/meta";
import { hashIdentifier, internationalPhone } from "@allsignalme/customer-data-normalize/openai";

hash("em", " John_Smith@gmail.com ");             // "62a14e44…"
hash("ph", "0912 345 678", { country: "VN" });    // hash of "84912345678"
normalize("ct", "New York");                      // "newyork"
hashIdentifier("email", "  John.Doe@Example.COM "); // "836f82db…"
```

Functions return `null` when the field must not be sent.

## Development

```bash
make test    # Go + TypeScript tests (vectors)
make lint    # gofmt + go vet
make gen     # regenerate Meta lookup tables from data/meta-tables.json
make build   # compile src/ to dist/
make check   # everything CI checks, including that generated files and dist/ are committed
```

TypeScript tests run on Node 24+ (native type stripping). Changing a rule means changing the spec: read the
versioning section of [SPEC.md](SPEC.md) first.

## Release

1. Update `CHANGELOG.md` and the `version` in `package.json`.
2. `make check`, commit, then tag `vX.Y.Z` on `main` and push the tag. Go and npm consumers both use the tag.

## License

Parts of the Meta normalization (lookup tables, behaviour) derive from meta/capi-param-builder; see
[THIRD_PARTY_NOTICES.md](THIRD_PARTY_NOTICES.md).
