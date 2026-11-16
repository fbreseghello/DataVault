// The ledger's arithmetic: when a bill comes due next, what it costs a month,
// and how money is written and read. Dates are YYYY-MM-DD on the viewer's calendar.

import type { Bill, Cycle } from './api'

/** Bills due within this many days are flagged in the list. */
export const SOON_DAYS = 7

/** Bills ending within this many days are flagged too, in time to renew or cancel. */
export const ENDING_DAYS = 30

export const CYCLES: { cycle: Cycle; label: string; short: string }[] = [
  { cycle: 'monthly', label: 'month', short: 'mo' },
  { cycle: 'yearly', label: 'year', short: 'yr' },
  { cycle: 'weekly', label: 'week', short: 'wk' },
]

const pad = (n: number) => String(n).padStart(2, '0')

export function todayISO(): string {
  const d = new Date()
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}`
}

function ymd(s: string): [number, number, number] {
  const [y, m, d] = s.split('-').map(Number)
  return [y, m, d]
}

function iso(y: number, m: number, d: number): string {
  return `${y}-${pad(m)}-${pad(d)}`
}

// the day-th of a month, held to the month's last day (the 31st in February is the 28th or 29th)
function clamped(y: number, m: number, day: number): string {
  const last = new Date(Date.UTC(y, m, 0)).getUTCDate()
  return iso(y, m, Math.min(day, last))
}

/** The k-th charge (0 is the first), on the cycle from the first charge. */
export function chargeAt(b: Pick<Bill, 'since' | 'cycle'>, k: number): string {
  const [y, m, d] = ymd(b.since)
  if (b.cycle === 'weekly') {
    const t = new Date(Date.UTC(y, m - 1, d) + k * 7 * 86_400_000)
    return iso(t.getUTCFullYear(), t.getUTCMonth() + 1, t.getUTCDate())
  }
  if (b.cycle === 'yearly') return clamped(y + k, m, d)
  const month = m - 1 + k
  return clamped(y + Math.floor(month / 12), (month % 12) + 1, d)
}

// the index of the first charge on or after day, found by estimate and then a step either way
function indexFrom(b: Pick<Bill, 'since' | 'cycle'>, day: string): number {
  if (day <= b.since) return 0
  const [y, m, d] = ymd(b.since)
  const [ty, tm, td] = ymd(day)
  let k =
    b.cycle === 'weekly'
      ? Math.floor((Date.UTC(ty, tm - 1, td) - Date.UTC(y, m - 1, d)) / (7 * 86_400_000))
      : b.cycle === 'yearly'
        ? ty - y
        : (ty - y) * 12 + (tm - m)
  k = Math.max(0, k - 1)
  while (chargeAt(b, k) < day) k++
  return k
}

/** What decides when a bill charges. */
type Schedule = Pick<Bill, 'since' | 'cycle' | 'until' | 'pause_from' | 'pause_until'>

/** True once the bill's end has come: it charges no more. */
export function ended(b: Pick<Bill, 'until'>, today = todayISO()): boolean {
  return !!b.until && b.until <= today
}

/** True when the bill's pause covers day. */
export function paused(b: Pick<Bill, 'pause_from' | 'pause_until'>, day = todayISO()): boolean {
  return !!b.pause_from && b.pause_from <= day && (!b.pause_until || day < b.pause_until)
}

// past this many steps a search gives up: about forty years of weekly charges
const MAX_STEPS = 2000

/**
 * The first charge on or after today (or the first charge, if that is still
 * ahead), skipping a pause. Null when the bill ends, or stays paused, before it.
 */
export function nextCharge(b: Schedule, today = todayISO()): string | null {
  for (let k = indexFrom(b, today), n = 0; n < MAX_STEPS; k++, n++) {
    const c = chargeAt(b, k)
    if (b.until && c >= b.until) return null
    if (!paused(b, c)) return c
    if (!b.pause_until) return null
  }
  return null
}

/** The last charge before the end, skipping a pause; null when the bill has no end or never charges. */
export function lastCharge(b: Schedule): string | null {
  if (!b.until) return null
  for (let k = indexFrom(b, b.until) - 1; k >= 0; k--) {
    const c = chargeAt(b, k)
    if (!paused(b, c)) return c
  }
  return null
}

/** Every charge from the first up to (not including) stop, skipping a pause. */
export function chargesBefore(b: Schedule, stop: string): string[] {
  const end = b.until && b.until < stop ? b.until : stop
  const out: string[] = []
  for (let k = 0, n = indexFrom(b, end); k < n && k < MAX_STEPS * 5; k++) {
    const c = chargeAt(b, k)
    if (!paused(b, c)) out.push(c)
  }
  return out
}

/** What a charge on day cost, by the price history. */
export function priceOn(b: Pick<Bill, 'amount' | 'history'>, day: string): number {
  return b.history?.find((h) => day < h.until)?.amount ?? b.amount
}

/** How many charges have fallen so far, today's included, and what they add up to at the price of each day. */
export function paidSoFar(b: Schedule & Pick<Bill, 'amount' | 'history'>, today = todayISO()): { count: number; cents: number } {
  const charges = chargesBefore(b, addDay(today))
  return { count: charges.length, cents: charges.reduce((sum, c) => sum + priceOn(b, c), 0) }
}

function addDay(day: string, n = 1): string {
  const [y, m, d] = ymd(day)
  const t = new Date(Date.UTC(y, m - 1, d + n))
  return iso(t.getUTCFullYear(), t.getUTCMonth() + 1, t.getUTCDate())
}

/** What a bill costs a month, in cents. */
export function perMonth(b: Pick<Bill, 'amount' | 'cycle'>): number {
  if (b.cycle === 'yearly') return b.amount / 12
  if (b.cycle === 'weekly') return (b.amount * 52) / 12
  return b.amount
}

/** True when a bill counts in the totals: not ended and not paused today. */
export function counts(b: Bill, today = todayISO()): boolean {
  return !ended(b, today) && !paused(b, today)
}

/** Monthly totals by currency of the bills that count today, largest first. */
export function totals(bills: Bill[], today = todayISO()): { currency: string; month: number }[] {
  const by = new Map<string, number>()
  for (const b of bills) if (counts(b, today)) by.set(b.currency, (by.get(b.currency) ?? 0) + perMonth(b))
  return [...by].map(([currency, month]) => ({ currency, month })).sort((a, b) => b.month - a.month)
}

/** The bills that count today, grouped by the card or account they are paid with, the costliest first. */
export function byMethod(bills: Bill[], today = todayISO()): { method: string; count: number; sums: { currency: string; month: number }[] }[] {
  const groups = new Map<string, Bill[]>()
  for (const b of bills) {
    if (!counts(b, today)) continue
    const key = b.method.trim()
    const found = [...groups.keys()].find((k) => k.toLowerCase() === key.toLowerCase())
    groups.set(found ?? key, [...(groups.get(found ?? key) ?? []), b])
  }
  return [...groups]
    .map(([method, list]) => ({ method, count: list.length, sums: totals(list, today) }))
    .sort((a, b) => (b.sums[0]?.month ?? 0) - (a.sums[0]?.month ?? 0) || a.method.localeCompare(b.method))
}

/** True when a bill matches what is typed in the search: name, card or account, note or currency. */
export function matches(b: Bill, query: string): boolean {
  const words = query.toLowerCase().split(/\s+/).filter(Boolean)
  const text = `${b.name} ${b.method} ${b.note} ${b.currency}`.toLowerCase()
  return words.every((w) => text.includes(w))
}

/** Cents as the viewer writes money: "R$ 55,90", "US$ 60.00". */
export function money(cents: number, currency: string): string {
  try {
    return new Intl.NumberFormat(undefined, { style: 'currency', currency }).format(cents / 100)
  } catch {
    return `${currency} ${(cents / 100).toFixed(2)}`
  }
}

/**
 * Reads an amount typed by hand into cents: "55,90", "55.90", "1.234,56",
 * "1,234.56" and "60" all work. The last separator followed by one or two
 * digits is the decimal one. Null when it is not a number.
 */
export function parseAmount(s: string): number | null {
  const t = s.replace(/[^\d.,]/g, '')
  if (!/\d/.test(t)) return null
  const m = /^(.*?)[.,](\d{1,2})$/.exec(t)
  const whole = (m ? m[1] : t).replace(/[.,]/g, '')
  const cents = m ? m[2].padEnd(2, '0') : '00'
  const n = Number(whole || '0') * 100 + Number(cents)
  return Number.isSafeInteger(n) ? n : null
}

/** Cents as the amount field shows them for editing: "55,90" or "55.90". */
export function amountText(cents: number): string {
  const sep = (1.5).toLocaleString().includes(',') ? ',' : '.'
  return `${Math.floor(cents / 100)}${sep}${pad(cents % 100)}`
}
