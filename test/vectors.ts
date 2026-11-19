import { readFileSync } from "node:fs";

export function loadVectors<T>(name: string, spec: string): T[] {
  const file = JSON.parse(readFileSync(new URL(`../vectors/${name}`, import.meta.url), "utf8"));
  if (file.spec !== spec || file.version !== 1 || !Array.isArray(file.vectors) || file.vectors.length === 0) {
    throw new Error(`unexpected vector file ${name}: spec=${file.spec} version=${file.version}`);
  }
  return file.vectors as T[];
}
