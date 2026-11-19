// Spec phone-international v1 (see SPEC.md): converts a phone number taken from an address (often national format,
// "0983 714 965") to international format using the country of that same address. Converts only when certain: the
// number carries that country's national trunk prefix (0, 8, or North American numbering). Countries without a trunk
// prefix (IT, ES, PL...) or missing from the table => null: better no phone than a hash of a wrong number.
// Calling codes by ISO 3166-1 alpha-2 country, grouped by national trunk prefix.
const TRUNK_0 = {
    VN: "84", TH: "66", ID: "62", MY: "60", PH: "63", JP: "81", KR: "82", CN: "86", TW: "886", IN: "91", PK: "92",
    BD: "880", LK: "94", AU: "61", NZ: "64", GB: "44", IE: "353", FR: "33", DE: "49", NL: "31", BE: "32", AT: "43",
    CH: "41", SE: "46", FI: "358", TR: "90", IL: "972", AE: "971", SA: "966", EG: "20", ZA: "27", NG: "234", KE: "254",
    UA: "380",
};
const TRUNK_8 = { RU: "7", KZ: "7" };
const NANP = new Set(["US", "CA", "PR"]);
const STRIP = /[\s().-]/g;
// internationalPhone returns the number in international format (+<code>...), the input unchanged (trimmed) when it
// already starts with + or 00, or null when it cannot be converted safely.
export function internationalPhone(raw, countryCode) {
    const phone = (raw ?? "").trim();
    if (!phone)
        return null;
    if (/^(\+|00)/.test(phone))
        return phone;
    const digits = phone.replace(STRIP, "");
    const country = (countryCode ?? "").trim().toUpperCase();
    if (!/^[0-9]+$/.test(digits) || !country)
        return null;
    if (NANP.has(country)) {
        if (digits.length === 10)
            return `+1${digits}`;
        return digits.length === 11 && digits.startsWith("1") ? `+${digits}` : null;
    }
    const trunk0 = Object.prototype.hasOwnProperty.call(TRUNK_0, country) ? TRUNK_0[country] : undefined;
    if (trunk0)
        return digits.startsWith("0") && digits.length > 1 ? `+${trunk0}${digits.slice(1)}` : null;
    const trunk8 = Object.prototype.hasOwnProperty.call(TRUNK_8, country) ? TRUNK_8[country] : undefined;
    if (trunk8)
        return digits.startsWith("8") && digits.length > 1 ? `+${trunk8}${digits.slice(1)}` : null;
    return null;
}
