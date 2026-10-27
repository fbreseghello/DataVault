// The forge: passwords made from the browser's cryptographic randomness. Every
// character is drawn without bias, and each kind asked for appears at least once.

export interface ForgeOptions {
  length: number
  upper: boolean
  lower: boolean
  digits: boolean
  symbols: boolean
  /** Leave out characters that are easy to mistake for one another. */
  plain: boolean
}

export const MIN_LENGTH = 8
export const MAX_LENGTH = 128

export const DEFAULTS: ForgeOptions = { length: 24, upper: true, lower: true, digits: true, symbols: true, plain: false }

const SETS = {
  upper: 'ABCDEFGHIJKLMNOPQRSTUVWXYZ',
  lower: 'abcdefghijklmnopqrstuvwxyz',
  digits: '0123456789',
  symbols: '!#$%&*+-=?@^_~.:;,/|()[]{}<>',
}
const LOOKALIKES = new Set('0Oo1lI|')

function kinds(o: ForgeOptions): string[] {
  const out: string[] = []
  for (const k of ['upper', 'lower', 'digits', 'symbols'] as const) {
    if (!o[k]) continue
    const set = o.plain ? [...SETS[k]].filter((c) => !LOOKALIKES.has(c)).join('') : SETS[k]
    if (set) out.push(set)
  }
  return out
}

// a uniform integer in [0, n), rejecting the values that would bias the draw
function below(n: number): number {
  const limit = Math.floor(0x1_0000_0000 / n) * n
  const buf = new Uint32Array(1)
  for (;;) {
    crypto.getRandomValues(buf)
    if (buf[0] < limit) return buf[0] % n
  }
}

/** A password for the options, or "" when no kind of character is chosen. */
export function forge(o: ForgeOptions): string {
  const sets = kinds(o)
  if (!sets.length) return ''
  const length = Math.max(MIN_LENGTH, Math.min(MAX_LENGTH, Math.round(o.length)))
  const pool = sets.join('')
  // one of each kind first, the rest from the whole pool, then shuffled
  const chars = sets.map((s) => s[below(s.length)])
  while (chars.length < length) chars.push(pool[below(pool.length)])
  for (let i = chars.length - 1; i > 0; i--) {
    const j = below(i + 1)
    ;[chars[i], chars[j]] = [chars[j], chars[i]]
  }
  return chars.join('')
}

/** How hard the password is to guess, in bits: length × log2(pool size). */
export function bits(o: ForgeOptions): number {
  const pool = kinds(o).join('').length
  if (!pool) return 0
  return Math.floor(Math.max(MIN_LENGTH, Math.min(MAX_LENGTH, o.length)) * Math.log2(pool))
}
