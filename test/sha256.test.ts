import assert from "node:assert/strict";
import { createHash } from "node:crypto";
import { test } from "node:test";
import { sha256Hex } from "../src/sha256.ts";

const nodeSha256 = (s: string): string => createHash("sha256").update(s, "utf8").digest("hex");

test("sha256Hex: matches node:crypto for edge inputs", () => {
  const inputs = ["", "abc", "a".repeat(55), "a".repeat(56), "a".repeat(64), "a".repeat(1000), "nguyễnvăn", "정", "😀",
    "\ud800", "x\udc00y"];
  for (const s of inputs) assert.equal(sha256Hex(s), nodeSha256(s.toWellFormed()), JSON.stringify(s));
});

test("sha256Hex: matches node:crypto for random strings", () => {
  let seed = 42;
  const next = () => (seed = (seed * 1103515245 + 12345) % 2147483648);
  for (let i = 0; i < 500; i++) {
    const len = next() % 200;
    let s = "";
    for (let j = 0; j < len; j++) s += String.fromCodePoint(next() % 3 === 0 ? next() % 0x10ffff : next() % 0x800);
    assert.equal(sha256Hex(s), nodeSha256(s.toWellFormed()), JSON.stringify(s));
  }
});
