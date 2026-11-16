import { describe, expect, it } from 'vitest'
import { parseTitle } from './tags'

describe('parseTitle', () => {
  it('ends a tag at two spaces', () => {
    expect(parseTitle('!dtvlt(r)  fix login')).toEqual({ tags: [{ name: 'dtvlt', color: 'red' }], text: 'fix login' })
    expect(parseTitle('!dtvlt  fix login')).toEqual({ tags: [{ name: 'dtvlt', color: null }], text: 'fix login' })
  })

  it('reads one space as part of the tag', () => {
    expect(parseTitle('!dtvlt fix login')).toEqual({ tags: [{ name: 'dtvlt fix login', color: null }], text: '' })
  })

  it('reads several tags, colors in either case', () => {
    expect(parseTitle('!a(B)  !b(g)   rest  here')).toEqual({
      tags: [
        { name: 'a', color: 'blue' },
        { name: 'b', color: 'green' },
      ],
      text: 'rest  here',
    })
  })

  it('leaves text alone that is not a tag', () => {
    expect(parseTitle('plain !x  title')).toEqual({ tags: [], text: 'plain !x  title' })
    expect(parseTitle('!  title')).toEqual({ tags: [], text: '!  title' })
    expect(parseTitle('!(r)  title')).toEqual({ tags: [], text: '!(r)  title' })
  })

  it('keeps an unknown color letter in the name', () => {
    expect(parseTitle('!x(y)  t')).toEqual({ tags: [{ name: 'x(y)', color: null }], text: 't' })
  })
})
