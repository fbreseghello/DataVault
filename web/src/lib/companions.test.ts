import { describe, expect, it } from 'vitest'
import { CATEGORIES, categoryOf, ordered, roman } from './companions'

describe('roman', () => {
  it('numbers a card the way the tarot does', () => {
    expect([1, 4, 9, 14, 21, 40, 90, 2026].map(roman)).toEqual(['I', 'IV', 'IX', 'XIV', 'XXI', 'XL', 'XC', 'MMXXVI'])
    expect(roman(0)).toBe('')
  })
})

describe('the catalog', () => {
  it('has seven categories, each with a 16x16 seal and its own companions', () => {
    expect(CATEGORIES.map((c) => c.id)).toEqual(['wanderers', 'animals', 'shinigami', 'demons', 'bushi', 'fae', 'denizens'])
    const seen = new Set<string>()
    for (const c of CATEGORIES) {
      expect(c.seal).toHaveLength(16)
      for (const r of c.seal) expect(r).toMatch(/^[.#]{16}$/)
      for (const x of c.companions) {
        expect(seen.has(x.id)).toBe(false)
        seen.add(x.id)
      }
    }
    expect(seen.size).toBe(34)
  })

  it('finds a companion\'s category', () => {
    expect(categoryOf('raven')?.id).toBe('animals')
    expect(categoryOf('knight')).toBeUndefined()
  })

  it('puts a theme\'s categories first and keeps the rest', () => {
    const ids = ordered(['fae', 'wanderers', 'nope']).map((c) => c.id)
    expect(ids.slice(0, 2)).toEqual(['fae', 'wanderers'])
    expect(ids).toHaveLength(7)
  })
})
