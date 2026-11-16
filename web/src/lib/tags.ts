// Tags at the start of a siege title: "!dtvlt(r)  fix the login" shows a red
// "dtvlt" tag before "fix the login". A tag runs from the "!" to the first two
// spaces, so "!dtvlt fix" with one space is all tag. A color letter in
// parentheses at the tag's end picks its color; with none it takes the accent.
// Several tags can follow each other: "!dtvlt(r)  !web  fix the login".

export type TagColor = 'red' | 'blue' | 'green'

export interface Tag {
  name: string
  color: TagColor | null
}

const COLORS: Record<string, TagColor> = { r: 'red', b: 'blue', g: 'green' }

/** Splits a title into its leading tags and the rest. */
export function parseTitle(title: string): { tags: Tag[]; text: string } {
  const tags: Tag[] = []
  let s = title.trim()
  while (s.startsWith('!')) {
    const end = s.indexOf('  ')
    let name = (end < 0 ? s : s.slice(0, end)).slice(1).trim()
    let color: TagColor | null = null
    const m = /\(([rbg])\)$/i.exec(name)
    if (m) {
      color = COLORS[m[1].toLowerCase()]
      name = name.slice(0, -m[0].length).trim()
    }
    // a lone "!" or "!(r)" is not a tag, just text
    if (!name) break
    tags.push({ name, color })
    s = end < 0 ? '' : s.slice(end).trim()
  }
  return { tags, text: s }
}
