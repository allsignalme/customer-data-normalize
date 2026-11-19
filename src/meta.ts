// Spec meta v1 (see SPEC.md): normalizes and hashes customer information parameters for the Meta Conversions API.
// Normalization reproduces Meta's reference implementation (meta/capi-param-builder) so the value hashed here equals
// the value Meta's own browser code would hash.

import { COUNTRY_NAMES, FEMALE_WORDS, MALE_WORDS, STATE_NAMES } from "./meta-tables.gen.ts";
import { internationalPhone } from "./phone.ts";
import { sha256Hex } from "./sha256.ts";

// Hashed user_data keys.
export type MetaField = "em" | "ph" | "fn" | "ln" | "db" | "ge" | "ct" | "st" | "zp" | "country" | "external_id";

export interface MetaOptions {
  // ISO 3166-1 alpha-2 country of the customer (any case). Used by "ph" only: a national number with that country's
  // trunk prefix gets the country calling code (see phone.ts).
  country?: string | null;
}

const FIELDS: ReadonlySet<string> = new Set(["em", "ph", "fn", "ln", "db", "ge", "ct", "st", "zp", "country", "external_id"]);

// shouldHash reports whether a user_data key is one of the hashed fields. Keys such as client_ip_address,
// client_user_agent, fbc and fbp must be sent in clear.
export const shouldHash = (key: string): key is MetaField => FIELDS.has(key);

const SHA256_HEX = /^[0-9a-f]{64}$/i;

// isSha256Hex reports whether value is a SHA-256 hex digest (64 hex characters, either case).
export const isSha256Hex = (value: string): boolean => SHA256_HEX.test(value);

const EMAIL = /^[\w!#$%&'*+/=?^`{|}~-]+(?:\.[\w!#$%&'*+/=?^`{|}~-]+)*@(?:[a-z0-9](?:[a-z0-9-]*[a-z0-9])?\.)+[a-z0-9](?:[a-z0-9-]*[a-z0-9])?$/i;
const NAME_STRIP = /[!"#$%&'()*+,\-./:;<=>?@[\\\]^_`{|}~\s]+/g;

const normalizeEmail = (raw: string): string => {
  const email = raw.toLowerCase().trim();
  return EMAIL.test(email) ? email : "";
};

// normalizePhone keeps digits and drops leading zeros. With a country, a national number is first converted to
// international format when that can be done safely (deviation D3).
const normalizePhone = (raw: string, country: string | null | undefined): string => {
  const intl = country ? internationalPhone(raw, country) : null;
  return (intl ?? raw).replace(/[^0-9]/g, "").replace(/^0+/, "");
};

// normalizeName removes whitespace and ASCII punctuation, then lowercases. Other characters are kept.
const normalizeName = (raw: string): string => raw.replace(NAME_STRIP, "").toLowerCase();

const pad = (n: number, width: number): string => String(n).padStart(width, "0");

function validDate(year: number, month: number, day: number, now: Date): boolean {
  if (year < 1800 || year > now.getUTCFullYear() + 1 || month < 1 || month > 12 || day < 1) return false;
  return day <= new Date(Date.UTC(year, month, 0)).getUTCDate();
}

// normalizeDateOfBirth reads the first three digit groups as a date (year is the group > 31; month/day order is
// decided by which value is > 12, defaulting to month first), or a single 8-digit group as is. Years from 1800 to
// next year (UTC).
export function normalizeDateOfBirth(raw: string, now: Date = new Date()): string {
  const groups = raw.replace(/\D/g, " ").trim().split(" ").filter((g) => g.length > 0);
  if (groups.length >= 3) {
    const [n1, n2, n3] = groups.slice(0, 3).map((g) => parseInt(g, 10)) as [number, number, number];
    let year: number;
    let month: number;
    let day: number;
    if (n1 > 31) {
      year = n1;
      month = n2 > 12 ? n3 : n2;
      day = n2 > 12 ? n2 : n3;
    } else if (n2 > 31) {
      year = n2;
      month = n3 > 12 ? n1 : n3;
      day = n3 > 12 ? n3 : n1;
    } else {
      year = n3;
      month = n1 > 12 ? n2 : n1;
      day = n1 > 12 ? n1 : n2;
    }
    return validDate(year, month, day, now) ? `${pad(year, 4)}${pad(month, 2)}${pad(day, 2)}` : "";
  }
  const only = groups[0];
  return groups.length === 1 && only !== undefined && only.length === 8 ? only : "";
}

const GENDER: ReadonlyMap<string, string> = new Map([
  ...MALE_WORDS.map((w): [string, string] => [w, "m"]),
  ...FEMALE_WORDS.map((w): [string, string] => [w, "f"]),
]);

// normalizeGender maps a gender word (male, female, woman, m, f...) to "m" or "f". Unknown words are dropped.
const normalizeGender = (raw: string): string => GENDER.get(raw.toLowerCase().trim()) ?? "";

// normalizeCity keeps only ASCII letters and digits, lowercased; the result must start with a letter.
const normalizeCity = (raw: string): string => {
  const city = raw.replace(/[^a-zA-Z0-9]+/g, "").toLowerCase();
  return /^[a-z]/.test(city) ? city : "";
};

const STATE_INDEX: ReadonlyMap<string, string> = new Map(STATE_NAMES);
const COUNTRY_INDEX: ReadonlyMap<string, string> = new Map(COUNTRY_NAMES);

// normalizeAbbreviation (state, country): lowercase, keep a-z, then a two-letter value as is, a known name (exact,
// then the first name contained in the value, in table order) as its code, anything else cut to two letters.
function normalizeAbbreviation(
  raw: string,
  rows: ReadonlyArray<readonly [string, string]>,
  index: ReadonlyMap<string, string>,
): string {
  let value = raw.toLowerCase().trim().replace(/[^a-z]/g, "");
  if (value.length !== 2) {
    const code = index.get(value);
    if (code !== undefined) {
      value = code;
    } else {
      const hit = rows.find(([name]) => value.includes(name));
      if (hit) value = hit[1];
    }
  }
  return value.slice(0, 2);
}

// normalizeZip lowercases, keeps the part before the first dash, trims; at least 2 characters.
const normalizeZip = (raw: string): string => {
  const zip = (raw.toLowerCase().split("-", 1)[0] ?? "").trim();
  return zip.length >= 2 ? zip : "";
};

const normalizeExternalId = (raw: string): string => raw.replace(/\s+/g, "").toLowerCase();

// normalize returns the normalized value of raw for field, or null when it is empty, invalid or the field is unknown
// (do not send it). A value that is already a SHA-256 hex digest (after trimming) is returned lowercased as is.
export function normalize(field: string, raw: string | null | undefined, options: MetaOptions = {}): string | null {
  if (typeof raw !== "string" || raw === "" || !shouldHash(field)) return null;
  const trimmed = raw.trim();
  if (SHA256_HEX.test(trimmed)) return trimmed.toLowerCase();
  let value: string;
  switch (field) {
    case "em":
      value = normalizeEmail(raw);
      break;
    case "ph":
      value = normalizePhone(raw, options.country);
      break;
    case "fn":
    case "ln":
      value = normalizeName(raw);
      break;
    case "db":
      value = normalizeDateOfBirth(raw);
      break;
    case "ge":
      value = normalizeGender(raw);
      break;
    case "ct":
      value = normalizeCity(raw);
      break;
    case "st":
      value = normalizeAbbreviation(raw, STATE_NAMES, STATE_INDEX);
      break;
    case "zp":
      value = normalizeZip(raw);
      break;
    case "country":
      value = normalizeAbbreviation(raw, COUNTRY_NAMES, COUNTRY_INDEX);
      break;
    case "external_id":
      value = normalizeExternalId(raw);
      break;
  }
  return value ? value : null;
}

// hash returns the lowercase hex SHA-256 of the normalized value, or null (do not send the field). A value that is
// already a SHA-256 hex digest is returned lowercased, not hashed again.
export function hash(field: string, raw: string | null | undefined, options: MetaOptions = {}): string | null {
  const value = normalize(field, raw, options);
  if (value === null) return null;
  return SHA256_HEX.test((raw as string).trim()) ? value : sha256Hex(value);
}
