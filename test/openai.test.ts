import assert from "node:assert/strict";
import { test } from "node:test";
import { appendIdentifier, hashIdentifier, internationalPhone, normalizeByKind, type IdentityKind } from "../src/openai.ts";
import { loadVectors } from "./vectors.ts";

interface OpenAIVector {
  kind: IdentityKind;
  input: string;
  normalized: string | null;
  sha256: string | null;
  note: string;
}

const vectors = loadVectors<OpenAIVector>("openai-v1.json", "openai");

test("normalizeByKind: openai v1 vectors match expected", () => {
  for (const v of vectors) {
    assert.equal(normalizeByKind(v.kind, v.input), v.normalized, `${v.kind} ${JSON.stringify(v.input)} (${v.note})`);
  }
});

test("hashIdentifier: openai v1 vectors match expected", () => {
  for (const v of vectors) {
    assert.equal(hashIdentifier(v.kind, v.input), v.sha256, `${v.kind} ${JSON.stringify(v.input)} (${v.note})`);
  }
});

test("appendIdentifier: keeps the first three unique values", () => {
  let list: string[] = [];
  for (const raw of ["a@example.com", "A@example.com ", "bad", "b@example.com", "c@example.com", "d@example.com"]) {
    list = appendIdentifier(list, "email", raw);
  }
  assert.equal(list.length, 3);
  assert.equal(list[0], hashIdentifier("email", "a@example.com"));
});

test("internationalPhone: re-exported for address phones", () => {
  assert.equal(internationalPhone("0983 714 965", "VN"), "+84983714965");
});
