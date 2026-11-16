import { describe, expect, it } from 'vitest'
import type { Bill } from './api'
import { calendar } from './ics'

const TODAY = '2026-11-16'

function bill(o: Partial<Bill>): Bill {
  return {
    id: 'x',
    name: 'n',
    amount: 100,
    currency: 'BRL',
    cycle: 'monthly',
    since: '2026-01-31',
    method: '',
    note: '',
    created_at: '',
    updated_at: '',
    ...o,
  }
}

// the lines of one event, folded lines joined back
function lines(ics: string): string[] {
  return ics.replace(/\r\n /g, '').split('\r\n')
}

describe('calendar', () => {
  it('is a calendar with CRLF line ends', () => {
    const ics = calendar([], TODAY)
    expect(ics.startsWith('BEGIN:VCALENDAR\r\n')).toBe(true)
    expect(ics.endsWith('END:VCALENDAR\r\n')).toBe(true)
  })

  it('falls back to the last day of short months', () => {
    const l = lines(calendar([bill({})], TODAY))
    expect(l).toContain('DTSTART;VALUE=DATE:20260131')
    expect(l).toContain('RRULE:FREQ=MONTHLY;BYMONTHDAY=31,-1;BYSETPOS=1')
  })

  it('keeps 29 February on the last day of February', () => {
    const l = lines(calendar([bill({ cycle: 'yearly', since: '2024-02-29', until: '2028-01-01' })], TODAY))
    expect(l).toContain('RRULE:FREQ=YEARLY;BYMONTH=2;BYMONTHDAY=29,-1;BYSETPOS=1;UNTIL=20271231')
  })

  it('leaves out the charges of a pause', () => {
    const l = lines(calendar([bill({ pause_from: '2026-05-01', pause_until: '2026-07-01' })], TODAY))
    expect(l).toContain('EXDATE;VALUE=DATE:20260531,20260630')
  })

  it('ends at a pause with no end, and skips one that starts with the bill', () => {
    expect(lines(calendar([bill({ pause_from: '2026-05-01' })], TODAY))).toContain(
      'RRULE:FREQ=MONTHLY;BYMONTHDAY=31,-1;BYSETPOS=1;UNTIL=20260430',
    )
    expect(calendar([bill({ pause_from: '2026-01-31' })], TODAY)).not.toContain('BEGIN:VEVENT')
  })

  it('leaves out ended bills', () => {
    expect(calendar([bill({ until: '2026-06-01' })], TODAY)).not.toContain('BEGIN:VEVENT')
  })

  it('escapes text and folds long lines', () => {
    const ics = calendar([bill({ name: 'Stream; HD, 4K', note: 'x'.repeat(200) })], TODAY)
    expect(lines(ics).some((l) => l.startsWith('SUMMARY:Stream\\; HD\\, 4K'))).toBe(true)
    for (const l of ics.split('\r\n')) expect(new TextEncoder().encode(l).length).toBeLessThanOrEqual(75)
  })
})
