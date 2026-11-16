// The hollow's arithmetic for the page: its title, the XP bar, the attribute
// bars and the deeds by day. The vault works out the XP and the level.

import type { Attr, Deed, DeedKind } from './api'

export const ATTRS: Attr[] = ['vigor', 'memory', 'ward', 'hoard', 'thrift']

/** The title for a level: one every five levels, the last kept from then on. */
export function titleFor(level: number, titles: string[]): string {
  if (!titles.length) return ''
  return titles[Math.min(Math.floor(level / 5), titles.length - 1)]
}

/** A bar of width cells, as the page draws it: "[######..........]". */
export function bar(value: number, max: number, width = 16): string {
  const lit = max > 0 ? Math.min(width, Math.max(0, Math.floor((value / max) * width))) : 0
  return '#'.repeat(lit) + '.'.repeat(width - lit)
}

/** What a deed was, in the page's words. */
export function kindName(k: DeedKind): string {
  return k.replaceAll('_', ' ')
}

/** Deeds grouped by the day they were earned, newest day first, keeping their order. */
export function byDay(deeds: Deed[]): { day: string; deeds: Deed[] }[] {
  const out: { day: string; deeds: Deed[] }[] = []
  for (const d of deeds) {
    const last = out.at(-1)
    if (last?.day === d.day) last.deeds.push(d)
    else out.push({ day: d.day, deeds: [d] })
  }
  return out
}

/** A sprite's grid as SVG paths, one per tone (1 to 4), each run of a row one rectangle. */
export function spritePaths(rows: string[]): { tone: number; d: string }[] {
  const by = new Map<number, string[]>()
  rows.forEach((row, y) => {
    let x = 0
    while (x < row.length) {
      const ch = row[x]
      let end = x
      while (end < row.length && row[end] === ch) end++
      const tone = Number(ch)
      if (tone >= 1 && tone <= 4) {
        if (!by.has(tone)) by.set(tone, [])
        by.get(tone)!.push(`M${x} ${y}h${end - x}v1h${x - end}z`)
      }
      x = end
    }
  })
  return [...by].sort((a, b) => a[0] - b[0]).map(([tone, parts]) => ({ tone, d: parts.join('') }))
}

/** What each class looks like and leans to, for the first choice. */
export const CLASSES: { id: 'knight' | 'scribe' | 'warden' | 'thief' | 'cleric'; attr: Attr; look: string }[] = [
  { id: 'knight', attr: 'vigor', look: 'broad shoulders, a closed helm, a sword' },
  { id: 'scribe', attr: 'memory', look: 'hunched under a hood, a book and a quill' },
  { id: 'warden', attr: 'ward', look: 'a great shield held in front, closed armor' },
  { id: 'thief', attr: 'hoard', look: 'thin, a low hood, a dagger in each hand' },
  { id: 'cleric', attr: 'thrift', look: 'a long robe, a lantern on a staff' },
]
