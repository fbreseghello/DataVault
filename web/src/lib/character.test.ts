import { describe, expect, it } from 'vitest'
import type { Deed } from './api'
import { bar, byDay, kindName, titleFor } from './character'

const TITLES = ['the nameless', 'the unkindled', 'the wayfarer']

describe('titleFor', () => {
  it('moves every five levels and keeps the last', () => {
    expect(titleFor(1, TITLES)).toBe('the nameless')
    expect(titleFor(4, TITLES)).toBe('the nameless')
    expect(titleFor(5, TITLES)).toBe('the unkindled')
    expect(titleFor(10, TITLES)).toBe('the wayfarer')
    expect(titleFor(40, TITLES)).toBe('the wayfarer')
    expect(titleFor(3, [])).toBe('')
  })
})

describe('bar', () => {
  it('lights the share of the width', () => {
    expect(bar(0, 100, 10)).toBe('..........')
    expect(bar(50, 100, 10)).toBe('#####.....')
    expect(bar(100, 100, 10)).toBe('##########')
    expect(bar(5, 0, 4)).toBe('....')
  })
})

describe('byDay and kindName', () => {
  const deed = (day: string, source: string): Deed => ({ xp: 10, attr: 'vigor', kind: 'task_conquered', source, day, at: '' })

  it('groups deeds by day in their order', () => {
    const groups = byDay([deed('2026-11-16', 'a'), deed('2026-11-16', 'b'), deed('2026-11-15', 'c')])
    expect(groups.map((g) => [g.day, g.deeds.map((d) => d.source)])).toEqual([
      ['2026-11-16', ['a', 'b']],
      ['2026-11-15', ['c']],
    ])
  })

  it('names a kind in words', () => {
    expect(kindName('day_cleared')).toBe('day cleared')
  })
})

describe('spritePaths', () => {
  it('draws each run of a tone as one rectangle', async () => {
    const { spritePaths } = await import('./character')
    expect(spritePaths(['.11.', '2.22'])).toEqual([
      { tone: 1, d: 'M1 0h2v1h-2z' },
      { tone: 2, d: 'M0 1h1v1h-1zM2 1h2v1h-2z' },
    ])
  })

  it('has every class drawn on a 32x32 grid of tones', async () => {
    const sprites = (await import('./sprites.json')).default as Record<string, string[]>
    const { CLASSES } = await import('./character')
    for (const c of CLASSES) {
      const rows = sprites[c.id]
      expect(rows).toHaveLength(32)
      for (const r of rows) expect(r).toMatch(/^[.1-4]{32}$/)
    }
  })
})
