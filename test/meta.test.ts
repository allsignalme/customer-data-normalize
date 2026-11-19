import assert from "node:assert/strict";
import { test } from "node:test";
import { hash, normalize, normalizeDateOfBirth, shouldHash } from "../src/meta.ts";
import { loadVectors } from "./vectors.ts";

interface MetaVector {
  field: string;
  input: string;
  country: string | null;
  normalized: string | null;
  sha256: string | null;
  note: string;
}

const vectors = loadVectors<MetaVector>("meta-v1.json", "meta");

test("normalize: meta v1 vectors match expected", () => {
  for (const v of vectors) {
    assert.equal(normalize(v.field, v.input, { country: v.country }), v.normalized, `${v.field} ${JSON.stringify(v.input)} (${v.note})`);
  }
});

test("hash: meta v1 vectors match expected", () => {
  for (const v of vectors) {
    assert.equal(hash(v.field, v.input, { country: v.country }), v.sha256, `${v.field} ${JSON.stringify(v.input)} (${v.note})`);
  }
});

test("normalize: unknown field or non-string input is dropped", () => {
  for (const field of ["", "email", "client_ip_address", "fbp"]) {
    assert.equal(normalize(field, "John@example.com"), null);
    assert.equal(hash(field, "62a14e44f765419d10fea99367361a727c12365e2520f32218d505ed9aa0f62f"), null);
  }
  assert.equal(normalize("em", undefined), null);
  assert.equal(hash("em", null), null);
});

test("shouldHash: matches the Meta parameter table", () => {
  for (const k of ["em", "ph", "fn", "ln", "db", "ge", "ct", "st", "zp", "country", "external_id"]) assert.ok(shouldHash(k), k);
  for (const k of ["client_ip_address", "client_user_agent", "fbc", "fbp", "subscription_id", "fb_login_id", "lead_id",
    "anon_id", "madid", "page_id", "page_scoped_user_id", "ctwa_clid", "ig_account_id", "ig_sid", "", "EM"]) {
    assert.ok(!shouldHash(k), k);
  }
});

test("normalizeDateOfBirth: year bound follows the current year", () => {
  const now = new Date(Date.UTC(2026, 5, 1));
  assert.equal(normalizeDateOfBirth("2027-01-01", now), "20270101");
  assert.equal(normalizeDateOfBirth("2028-01-01", now), "");
  assert.equal(normalizeDateOfBirth("1800-01-01", now), "18000101");
});
