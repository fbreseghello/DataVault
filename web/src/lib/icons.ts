// Sidebar icons, drawn as 16x16 one-bit pixel art. `#` is a lit pixel; Icon.svelte
// turns each grid into a single SVG path. Edit the art here, not in the markup.

const grids = {
  // Grimoire: an open book, lines of writing cut into both pages.
  book: `
................
................
.###........###.
#######..#######
#######..#######
##...##..##...##
#######..#######
##...##..##...##
#######..#######
##..###..###..##
#######..#######
.######..######.
.......##.......
................
................
................`,

  // Broken heart: solid, split top to bottom by a jagged crack.
  heart: `
................
................
..####....####..
.######..######.
#######.########
########.#######
#######.########
######.#########
.######.#######.
..######.#####..
...#####.####...
....###.####....
.....###.##.....
......#.##......
.......##.......
................`,

  // Sword, point up and to the right, with a crossguard and pommel.
  sword: `
................
.............##.
............###.
...........###..
..........###...
.........###....
........###.....
.#.....###......
..#...###.......
...#.###........
....###.........
...#.#..........
..#...#.........
.##....#........
##..............
................`,

  // Seal: a padlock with a keyhole, for sealed grimoire entries.
  seal: `
................
.....######.....
....########....
....##....##....
....##....##....
....##....##....
..############..
..############..
..#####..#####..
..#####..#####..
..######.#####..
..############..
..############..
..############..
................
................`,
} as const

export type IconName = keyof typeof grids

// Each horizontal run of lit pixels becomes one rectangle in the path.
function toPath(grid: string): string {
  const rows = grid.trim().split('\n')
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

export const iconPaths = Object.fromEntries(
  Object.entries(grids).map(([name, grid]) => [name, toPath(grid)]),
) as Record<IconName, string>
