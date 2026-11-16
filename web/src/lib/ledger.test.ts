import { describe, expect, it } from 'vitest'
import type { Bill } from './api'
import {
  byMethod,
  chargeAt,
  ended,
  lastCharge,
  matches,
  nextCharge,
  paidSoFar,
  parseAmount,
  paused,
  priceOn,
  totals,
} from './ledger'

const TODAY = '2026-11-16'
const monthly = { since: '2026-01-31', cycle: 'monthly' as const, amount: 100 }

function bill(o: Partial<Bill>): Bill {
  return { id: 'x', name: 'n', currency: 'BRL', method: '', note: '', created_at: '', updated_at: '', ...monthly, ...o }
}

describe('chargeAt', () => {
  it('holds a late day to the end of a short month', () => {
    expect(chargeAt(monthly, 1)).toBe('2026-02-28')
    expect(chargeAt(monthly, 2)).toBe('2026-03-31')
  })
  it('keeps 29 February in leap years only', () => {
    const leap = { since: '2024-02-29', cycle: 'yearly' as const }
    expect(chargeAt(leap, 1)).toBe('2025-02-28')
    expect(chargeAt(leap, 4)).toBe('2028-02-29')
  })
  it('steps weeks across a month', () => {
    expect(chargeAt({ since: '2026-11-30', cycle: 'weekly' }, 1)).toBe('2026-12-07')
  })
})

describe('nextCharge', () => {
  it('finds the next charge on each cycle', () => {
    expect(nextCharge(monthly, TODAY)).toBe('2026-11-30')
    expect(nextCharge({ since: '2024-02-29', cycle: 'yearly' }, TODAY)).toBe('2027-02-28')
    expect(nextCharge({ since: '2026-11-02', cycle: 'weekly' }, TODAY)).toBe('2026-11-16')
    expect(nextCharge({ since: '2026-11-03', cycle: 'weekly' }, TODAY)).toBe('2026-11-17')
  })
  it('is the first charge while that is ahead', () => {
    expect(nextCharge({ since: '2027-01-01', cycle: 'monthly' }, TODAY)).toBe('2027-01-01')
  })
  it('stops at the end', () => {
    expect(nextCharge({ since: '2026-01-10', cycle: 'yearly', until: '2027-01-10' }, TODAY)).toBeNull()
    expect(nextCharge({ since: '2026-01-10', cycle: 'yearly', until: '2027-01-11' }, TODAY)).toBe('2027-01-10')
  })
  it('skips a pause, or waits on one with no end', () => {
    expect(nextCharge({ ...monthly, pause_from: '2026-11-01', pause_until: '2027-01-15' }, TODAY)).toBe('2027-01-31')
    expect(nextCharge({ ...monthly, pause_from: '2026-11-01' }, TODAY)).toBeNull()
    expect(nextCharge({ ...monthly, pause_from: '2026-12-01' }, TODAY)).toBe('2026-11-30')
  })
})

describe('lastCharge', () => {
  it('is the charge before the end', () => {
    expect(lastCharge({ since: '2026-01-10', cycle: 'yearly', until: '2027-01-10' })).toBe('2026-01-10')
    expect(lastCharge({ ...monthly, until: '2026-06-15' })).toBe('2026-05-31')
  })
  it('skips a pause', () => {
    expect(lastCharge({ ...monthly, until: '2026-06-15', pause_from: '2026-05-01' })).toBe('2026-04-30')
  })
  it('is null with no end', () => {
    expect(lastCharge(monthly)).toBeNull()
  })
})

describe('ended and paused', () => {
  it('ends on the end day', () => {
    expect(ended({ until: TODAY }, TODAY)).toBe(true)
    expect(ended({ until: '' }, TODAY)).toBe(false)
  })
  it('pauses from the start up to the day it resumes', () => {
    const p = { pause_from: '2026-11-01', pause_until: '2026-12-01' }
    expect(paused(p, '2026-11-01')).toBe(true)
    expect(paused(p, '2026-12-01')).toBe(false)
    expect(paused({ pause_from: '2026-11-01' }, '2030-01-01')).toBe(true)
  })
})

describe('paidSoFar', () => {
  it('counts the charges up to today', () => {
    expect(paidSoFar(monthly, TODAY)).toEqual({ count: 10, cents: 1000 })
    expect(paidSoFar({ ...monthly, since: TODAY }, TODAY)).toEqual({ count: 1, cents: 100 })
  })
  it('stops at the end and skips a pause', () => {
    expect(paidSoFar({ ...monthly, until: '2026-04-30' }, TODAY)).toEqual({ count: 3, cents: 300 })
    expect(paidSoFar({ ...monthly, pause_from: '2026-05-01', pause_until: '2026-08-01' }, TODAY)).toEqual({ count: 7, cents: 700 })
  })
  it('charges each day its own price', () => {
    const history = [
      { amount: 50, until: '2026-03-15' },
      { amount: 80, until: '2026-06-01' },
    ]
    expect(priceOn({ amount: 100, history }, '2026-03-14')).toBe(50)
    expect(priceOn({ amount: 100, history }, '2026-06-01')).toBe(100)
    expect(paidSoFar({ ...monthly, history }, TODAY)).toEqual({ count: 10, cents: 50 * 2 + 80 * 3 + 100 * 5 })
  })
})

describe('totals and byMethod', () => {
  it('leaves out paused and ended bills', () => {
    const list = [bill({}), bill({ pause_from: '2026-11-01' }), bill({ until: '2026-10-01' })]
    expect(totals(list, TODAY)).toEqual([{ currency: 'BRL', month: 100 }])
  })
  it('turns weeks and years into months', () => {
    const list = [bill({ cycle: 'yearly', amount: 1200 }), bill({ cycle: 'weekly', amount: 12 })]
    expect(totals(list, TODAY)).toEqual([{ currency: 'BRL', month: 100 + 52 }])
  })
  it('groups cards whatever their case', () => {
    const groups = byMethod([bill({ method: 'Visa' }), bill({ method: 'visa ' }), bill({ method: '' })], TODAY)
    expect(groups.map((g) => [g.method, g.count])).toEqual([
      ['Visa', 2],
      ['', 1],
    ])
  })
})

describe('matches', () => {
  it('needs every word somewhere in the bill', () => {
    expect(matches(bill({ name: 'Netflix', method: 'visa' }), 'net VISA')).toBe(true)
    expect(matches(bill({ name: 'Netflix' }), 'spot')).toBe(false)
    expect(matches(bill({ name: 'Domain' }), 'brl')).toBe(true)
  })
})

describe('parseAmount', () => {
  it('reads both decimal marks', () => {
    expect(parseAmount('55,90')).toBe(5590)
    expect(parseAmount('55.90')).toBe(5590)
    expect(parseAmount('1.234,56')).toBe(123456)
    expect(parseAmount('1,234.56')).toBe(123456)
    expect(parseAmount('60')).toBe(6000)
    expect(parseAmount('abc')).toBeNull()
  })
})
