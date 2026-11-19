export type MetaField = "em" | "ph" | "fn" | "ln" | "db" | "ge" | "ct" | "st" | "zp" | "country" | "external_id";
export interface MetaOptions {
    country?: string | null;
}
export declare const shouldHash: (key: string) => key is MetaField;
export declare const isSha256Hex: (value: string) => boolean;
export declare function normalizeDateOfBirth(raw: string, now?: Date): string;
export declare function normalize(field: string, raw: string | null | undefined, options?: MetaOptions): string | null;
export declare function hash(field: string, raw: string | null | undefined, options?: MetaOptions): string | null;
