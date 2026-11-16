// The ledger as a calendar file (RFC 5545): each running bill is an all-day
// event that repeats on its cycle, so any calendar app shows when it charges.

import type { Bill } from './api'
import { chargeAt, ended, money, todayISO } from './ledger'

const compact = (day: string) => day.replaceAll('-', '')

function shift(day: string, n: number): string {
  const [y, m, d] = day.split('-').map(Number)
  return new Date(Date.UTC(y, m - 1, d + n)).toISOString().slice(0, 10)
}

// text values escape backslashes, commas, semicolons and line breaks
function text(s: string): string {
  return s.replace(/[\\;,]/g, (c) => '\\' + c).replace(/\r?\n/g, '\\n')
}

// lines longer than 75 bytes go on in lines that start with a space, never splitting a character
function fold(line: string): string {
  const enc = new TextEncoder()
  const parts: string[] = []
  let cur = ''
  let bytes = 0
  for (const ch of line) {
    const n = enc.encode(ch).length
    if (bytes + n > (parts.length ? 74 : 75)) {
      parts.push(cur)
      cur = ''
      bytes = 0
    }
    cur += ch
    bytes += n
  }
  parts.push(cur)
  return parts.join('\r\n ')
}

// the repeat rule; days past the 28th fall back to the month's last day, as the ledger does
function rule(b: Bill, last: string | null): string {
  const [, m, d] = b.since.split('-').map(Number)
  let r = `FREQ=${b.cycle.toUpperCase()}`
  if (b.cycle === 'monthly' && d > 28) r += `;BYMONTHDAY=${d},-1;BYSETPOS=1`
  if (b.cycle === 'yearly' && d > 28) r += `;BYMONTH=${m};BYMONTHDAY=${d},-1;BYSETPOS=1`
  if (last) r += `;UNTIL=${compact(last)}`
  return r
}

function event(b: Bill, stamp: string): string[] | null {
  // the last day it may charge: the day before its end, or before a pause that has no end
  const stops = [b.until, b.pause_from && !b.pause_until ? b.pause_from : ''].filter(Boolean) as string[]
  const stop = stops.sort()[0]
  if (stop && stop <= b.since) return null
  const last = stop ? shift(stop, -1) : null

  // the charges a pause with an end skips
  const skipped: string[] = []
  if (b.pause_from && b.pause_until) {
    for (let k = 0; k < 2000; k++) {
      const c = chargeAt(b, k)
      if (c >= b.pause_until || (last && c > last)) break
      if (c >= b.pause_from) skipped.push(compact(c))
    }
  }

  const lines = [
    'BEGIN:VEVENT',
    `UID:${b.id}@datavault`,
    `DTSTAMP:${stamp}`,
    `DTSTART;VALUE=DATE:${compact(b.since)}`,
    `DTEND;VALUE=DATE:${compact(shift(b.since, 1))}`,
    `RRULE:${rule(b, last)}`,
  ]
  if (skipped.length) lines.push(`EXDATE;VALUE=DATE:${skipped.join(',')}`)
  lines.push(`SUMMARY:${text(`${b.name} · ${money(b.amount, b.currency)}`)}`)
  const about = [b.method && `paid with ${b.method}`, b.note].filter(Boolean).join('\n')
  if (about) lines.push(`DESCRIPTION:${text(about)}`)
  lines.push('TRANSP:TRANSPARENT')
  lines.push('BEGIN:VALARM', 'ACTION:DISPLAY', `DESCRIPTION:${text(`${b.name} charges tomorrow`)}`, 'TRIGGER:-P1D', 'END:VALARM')
  lines.push('END:VEVENT')
  return lines
}

/** The bills that have not ended, as the text of a .ics file. */
export function calendar(bills: Bill[], today = todayISO()): string {
  const stamp = new Date().toISOString().replace(/[-:]/g, '').replace(/\.\d+/, '')
  const lines = ['BEGIN:VCALENDAR', 'VERSION:2.0', 'PRODID:-//DataVault//ledger//EN', 'CALSCALE:GREGORIAN', 'X-WR-CALNAME:ledger']
  for (const b of bills) {
    if (ended(b, today)) continue
    const ev = event(b, stamp)
    if (ev) lines.push(...ev)
  }
  lines.push('END:VCALENDAR')
  return lines.map(fold).join('\r\n') + '\r\n'
}
