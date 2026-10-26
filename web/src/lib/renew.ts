// Renewal of keys and other entries that expire: how far off the due date is,
// in whole days of the viewer's calendar, and how to say it.

/** Entries due within this many days are flagged in the list. */
export const SOON_DAYS = 14

function today(): number {
  const d = new Date()
  return Date.UTC(d.getFullYear(), d.getMonth(), d.getDate())
}

/** Days from today to a YYYY-MM-DD date: 0 today, negative when past. */
export function daysUntil(due: string): number {
  const [y, m, d] = due.split('-').map(Number)
  return Math.round((Date.UTC(y, m - 1, d) - today()) / 86_400_000)
}

/** True when an entry is due soon or overdue, and should catch the eye. */
export function needsRenewal(due: string | undefined): boolean {
  return !!due && daysUntil(due) <= SOON_DAYS
}

/** A short mark for the list: "due 5d", "due today", "overdue". */
export function shortDue(due: string): string {
  const n = daysUntil(due)
  if (n < 0) return 'overdue'
  if (n === 0) return 'due today'
  return `due ${n}d`
}

/** A full sentence for the entry: "in 76 days", "today", "3 days overdue". */
export function longDue(due: string): string {
  const n = daysUntil(due)
  if (n === 0) return 'today'
  if (n === 1) return 'tomorrow'
  if (n > 0) return `in ${n} days`
  return n === -1 ? '1 day overdue' : `${-n} days overdue`
}

/** A date as the rest of the interface writes it: "12 Jan 2027". */
export function showDate(due: string): string {
  const [y, m, d] = due.split('-').map(Number)
  return new Date(y, m - 1, d).toLocaleDateString('en-GB', { day: 'numeric', month: 'short', year: 'numeric' })
}
