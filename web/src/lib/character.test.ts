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
