import assert from "node:assert/strict";
import { test } from "node:test";
import { internationalPhone } from "../src/phone.ts";
import { loadVectors } from "./vectors.ts";

interface PhoneVector {
  input: string;
  country: string | null;
  expected: string | null;
  note: string;
}

test("internationalPhone: phone-international v1 vectors match expected", () => {
  for (const v of loadVectors<PhoneVector>("phone-international-v1.json", "phone-international")) {
    assert.equal(internationalPhone(v.input, v.country), v.expected, `${JSON.stringify(v.input)} ${v.country} (${v.note})`);
  }
});
