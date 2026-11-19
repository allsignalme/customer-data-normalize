GO_ENV ?= GOMAXPROCS=2 GOMEMLIMIT=1500MiB

.PHONY: test test-go test-ts lint gen build check

test: test-go test-ts

test-go:
	$(GO_ENV) go test -count=1 ./...

test-ts:
	npm test

lint:
	@test -z "$$(gofmt -l .)" || (gofmt -l . && exit 1)
	go vet ./...

# Regenerates the Meta lookup tables (Go + TypeScript) from data/meta-tables.json.
gen:
	node scripts/gen-tables.mjs
	gofmt -w meta/tables_gen.go

# Compiles src/ to dist/. dist/ is committed so consumers can install from git without a build step.
build:
	npm run build

# What CI checks: tests, lint, generated files and dist up to date.
check: test lint gen build
	git diff --exit-code -- meta/tables_gen.go src/meta-tables.gen.ts dist
