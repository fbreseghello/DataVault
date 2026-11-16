// The companions' illustrations. Each lives in ./art as <companion>.webp, made
// from the Midjourney originals in art/hollow (docs/PERSONAGEM.md, "Os
// arquivos"); until a companion has its own, it shows the stand-in.
import placeholder from './art/placeholder.svg'

const drawn = import.meta.glob<string>('./art/*.webp', { eager: true, query: '?url', import: 'default' })

/** The picture for a companion, or the stand-in while it has none. */
export function artFor(companion: string): string {
  return drawn[`./art/${companion}.webp`] ?? placeholder
}

/** True once a companion has its own illustration. */
export function hasArt(companion: string): boolean {
  return `./art/${companion}.webp` in drawn
}
