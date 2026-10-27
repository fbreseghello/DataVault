// The ledger's arithmetic: when a bill comes due next, what it costs a month,
// and how money is written and read. Dates are YYYY-MM-DD on the viewer's calendar.

import type { Bill, Cycle } from './api'

/** Bills due within this many days are flagged in the list. */
export const SOON_DAYS = 7

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

/** The first charge on or after today (or after the first charge, if that is still ahead). */
export function nextCharge(b: Pick<Bill, 'since' | 'cycle'>, today = todayISO()): string {
  if (b.since >= today) return b.since
  const [y, m, d] = ymd(b.since)
  if (b.cycle === 'weekly') {
    const start = Date.UTC(y, m - 1, d)
    const [ty, tm, td] = ymd(today)
    const weeks = Math.ceil((Date.UTC(ty, tm - 1, td) - start) / (7 * 86_400_000))
    const next = new Date(start + weeks * 7 * 86_400_000)
    return iso(next.getUTCFullYear(), next.getUTCMonth() + 1, next.getUTCDate())
  }
  const [ty, tm] = ymd(today)
  if (b.cycle === 'yearly') {
    for (let year = ty; ; year++) {
      const c = clamped(year, m, d)
      if (c >= today) return c
    }
  }
  for (let k = 0; ; k++) {
    const month = tm - 1 + k
    const c = clamped(ty + Math.floor(month / 12), (month % 12) + 1, d)
    if (c >= today) return c
  }
}

/** What a bill costs a month, in cents. */
export function perMonth(b: Pick<Bill, 'amount' | 'cycle'>): number {
  if (b.cycle === 'yearly') return b.amount / 12
  if (b.cycle === 'weekly') return (b.amount * 52) / 12
  return b.amount
}

/** Monthly totals by currency, largest first. */
export function totals(bills: Bill[]): { currency: string; month: number }[] {
  const by = new Map<string, number>()
  for (const b of bills) by.set(b.currency, (by.get(b.currency) ?? 0) + perMonth(b))
  return [...by].map(([currency, month]) => ({ currency, month })).sort((a, b) => b.month - a.month)
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
