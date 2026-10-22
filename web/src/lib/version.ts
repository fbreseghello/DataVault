/** "0.5.1" -> "v0.5.1"; local builds stay "dev". */
export function showVersion(v: string): string {
  return /^\d/.test(v) ? 'v' + v : v
}
