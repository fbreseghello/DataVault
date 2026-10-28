// The pixel art itself lives in the themes (internal/theme/builtin/*.json, or
// an imported file): 16 rows of 16 characters, '#' lit and '.' dark. Icon.svelte
// draws whichever grid the current theme gives a slot.

export type { IconSlot as IconName } from './api'

// Each horizontal run of lit pixels becomes one rectangle in the path.
export function toPath(rows: readonly string[]): string {
  let d = ''
  rows.forEach((row, y) => {
    let x = 0
    while (x < row.length) {
      if (row[x] !== '#') {
        x++
        continue
      }
      const start = x
      while (row[x] === '#') x++
      d += `M${start} ${y}h${x - start}v1h-${x - start}z`
    }
  })
  return d
}
