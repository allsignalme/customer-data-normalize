// Spec openai v1 (see SPEC.md): normalizes and hashes identifiers for the OpenAI Conversions API. Matches the OpenAI
// browser SDK (oaiq) so browser and server produce the same value.
import { sha256Hex } from "./sha256.js";
export { internationalPhone } from "./phone.js";
const MAX_IDENTIFIERS_PER_FIELD = 3;
const MAX_EMAIL_LENGTH = 254;
const MAX_NAME_LENGTH = 1024;
const MAX_GEO_LENGTH = 128;
const MAX_POSTAL_LENGTH = 31;
// SDK email pattern (case-insensitive; \w is ASCII).
const EMAIL = /^[\w!#$%&'*+/=?^`{|}~-]+(?:\.[\w!#$%&'*+/=?^`{|}~-]+)*@(?:[a-z0-9](?:[a-z0-9-]*[a-z0-9])?\.)+[a-z0-9](?:[a-z0-9-]*[a-z0-9])?$/i;
// Whitespace + the 32 ASCII punctuation characters.
const NAME_STRIP = /[\s!"#$%&'()*+,\-./:;<=>?@[\\\]^_`{|}~]/g;
const PHONE_STRIP = /[\s().-]/g;
const SHA256_HEX = /^[0-9a-f]{64}$/i;
// normalizeEmail: trim + lowercase; longer than 254 characters or not an email => null.
export function normalizeEmail(raw) {
    const email = raw.trim().toLowerCase();
    return email.length <= MAX_EMAIL_LENGTH && EMAIL.test(email) ? email : null;
}
// normalizePhone removes spaces, parentheses, dots and dashes; the rest must be digits (optionally a leading +), not
// "+0". Drops the + and all leading zeros. With an international prefix (+ or 00) a number starting with 1 must have
// exactly 11 digits (North America); otherwise 8-15 digits. It does NOT add a country code: use internationalPhone
// first for address phones.
export function normalizePhone(raw) {
    const phone = raw.trim().replace(PHONE_STRIP, "");
    const body = phone.startsWith("+") ? phone.slice(1) : phone;
    if (!body || phone.startsWith("+0") || !/^[0-9]+$/.test(body))
        return null;
    const international = phone.startsWith("+") || phone.startsWith("00");
    const digits = body.replace(/^0+/, "");
    if (international && digits.startsWith("1") && digits.length !== 11)
        return null;
    return digits.length >= 8 && digits.length <= 15 ? digits : null;
}
// normalizeName removes whitespace and ASCII punctuation, lowercases, keeps non-ASCII characters without Unicode
// normalization; must keep at least one non-digit character.
export function normalizeName(raw) {
    if (raw.length > MAX_NAME_LENGTH)
        return null;
    const name = raw.replace(NAME_STRIP, "").toLowerCase();
    return /[^0-9]/.test(name) ? name : null;
}
// normalizeExternalId trims only; case is kept.
export function normalizeExternalId(raw) {
    const id = raw.trim();
    return id ? id : null;
}
// normalizeGeo (city, region): lowercase, collapse whitespace, keep accents; 1-128 characters. Sent in clear.
export function normalizeGeo(raw) {
    const geo = raw.toLowerCase().split(/\s+/).filter(Boolean).join(" ");
    return geo && geo.length <= MAX_GEO_LENGTH ? geo : null;
}
export function normalizeCountry(raw) {
    const country = normalizeGeo(raw);
    return country && /^[a-z]{2}$/.test(country) ? country : null;
}
export function normalizePostalCode(raw) {
    const postal = normalizeGeo(raw);
    return postal && postal.length <= MAX_POSTAL_LENGTH && /^[a-z0-9 -]+$/.test(postal) ? postal : null;
}
export const isSha256Hex = (value) => SHA256_HEX.test(value);
const NORMALIZERS = {
    email: normalizeEmail,
    phone: normalizePhone,
    first_name: normalizeName,
    last_name: normalizeName,
    external_id: normalizeExternalId,
    city: normalizeGeo,
    region: normalizeGeo,
    postal_code: normalizePostalCode,
    country: normalizeCountry,
};
export function normalizeByKind(kind, raw) {
    return Object.prototype.hasOwnProperty.call(NORMALIZERS, kind) ? NORMALIZERS[kind](raw) : null;
}
const HASHED_KINDS = new Set(["email", "phone", "first_name", "last_name", "external_id"]);
// hashIdentifier normalizes by kind, then returns the lowercase hex SHA-256. Already 64 hex characters => lowercased,
// not hashed again. Geo kinds are not hashed => null. Empty / invalid => null (do not send the field).
export function hashIdentifier(kind, raw) {
    if (!HASHED_KINDS.has(kind) || typeof raw !== "string")
        return null;
    const trimmed = raw.trim();
    if (!trimmed)
        return null;
    if (isSha256Hex(trimmed))
        return trimmed.toLowerCase();
    const normalized = normalizeByKind(kind, trimmed);
    return normalized ? sha256Hex(normalized) : null;
}
// appendIdentifier hashes raw and appends it when valid, not already present, and the list has fewer than 3 values.
export function appendIdentifier(list, kind, raw) {
    if (list.length >= MAX_IDENTIFIERS_PER_FIELD)
        return [...list];
    const hashed = hashIdentifier(kind, raw);
    return hashed && !list.includes(hashed) ? [...list, hashed] : [...list];
}
