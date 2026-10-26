// The candle rest screen: a skull lit by a candle on a shelf, against a brick
// wall that only shows where the flame reaches. Everything is a grid of
// character cells; each frame, the light that reaches a cell from the flame
// decides how bright it is drawn and, for the skull, how fine its strokes are.
//
// The candle and the skull come from scripts/art (candle3d.py, skull_light.py):
// the candle is a 3D render for a few positions of the swaying flame, the skull
// three levels of detail plus the relief of the bone.
import wax from './candle-wax.json'
import bone from './candle-skull.json'

const RAMP = ' .:-=+*#%@'
const DIGITS = '0123456789abcdefghijklmnopqrstuvwxyz'
const FLAME_ROWS = 10
const COLS = 130
const ROWS = FLAME_ROWS + wax.rows + 4
const SHELF_Y = ROWS - 4

export const SIZE = { cols: COLS, rows: ROWS }

type Cell =
  | { kind: 'shelf' | 'wall'; ch: string }
  | { kind: 'candle'; r: number; c: number }
  | { kind: 'skull'; chars: string[]; nx: number; ny: number; nz: number; hollow: number }

const unq = (ch: string, lo: number, hi: number) => lo + (DIGITS.indexOf(ch) / 35) * (hi - lo)
const lastRow = (rows: string[], has: (row: string) => boolean) => rows.reduce((last, row, r) => (has(row) ? r : last), 0)

// ------------------------------------------------------------------ the scene
const cells: (Cell | null)[][] = Array.from({ length: ROWS }, () => new Array(COLS).fill(null))
const hidden: boolean[][] = Array.from({ length: ROWS }, () => new Array(COLS).fill(false))

function put(x: number, y: number, ch: string, kind: 'shelf' | 'wall') {
  if (y >= 0 && y < ROWS && x >= 0 && x < COLS && ch !== ' ') cells[y][x] = { kind, ch }
}

// the shelf: a plank and two brackets
for (let x = 1; x < COLS - 1; x++) put(x, SHELF_Y, '=', 'shelf')
for (let x = 2; x < COLS - 2; x++) put(x, SHELF_Y + 1, '-', 'shelf')
for (const bx of [8, COLS - 10]) {
  put(bx, SHELF_Y + 2, '|', 'shelf')
  put(bx + 1, SHELF_Y + 2, '|', 'shelf')
  put(bx, SHELF_Y + 3, "'", 'shelf')
  put(bx + 1, SHELF_Y + 3, "'", 'shelf')
}

// the candle at the left; its lowest row of wax sits right on the plank
const CX = 6
const CY = SHELF_Y - 1 - lastRow(wax.mask, (row) => row.includes('1'))
for (let r = 0; r < wax.rows; r++) {
  for (let c = 0; c < wax.cols; c++) {
    if (wax.mask[r][c] !== '1') continue
    cells[CY + r][CX + c] = { kind: 'candle', r, c }
    hidden[CY + r][CX + c] = true
  }
}
const WICK_X = CX + wax.wick[0]
const WICK_Y = CY + wax.wick[1]

// the skull at the right, its face to the candle
const SKX = CX + wax.cols + 2
const SKY = SHELF_Y - 1 - lastRow(bone.levels[0], (row) => row.trim() !== '')
for (let r = 0; r < bone.rows; r++) {
  for (let c = 0; c < bone.cols; c++) {
    if (bone.inside[r][c] === '1') hidden[SKY + r][SKX + c] = true
    const chars = bone.levels.map((level) => level[r][c] || ' ')
    if (chars.every((ch) => ch === ' ')) continue
    const nx = unq(bone.nx[r][c], -1, 1)
    const ny = unq(bone.ny[r][c], -1, 1)
    const nz = Math.sqrt(Math.max(0, 1 - nx * nx - ny * ny))
    cells[SKY + r][SKX + c] = { kind: 'skull', chars, nx, ny, nz, hollow: unq(bone.hollow[r][c], 0, 1) }
  }
}

// the flame covers the wall behind it
for (let y = Math.max(0, WICK_Y - FLAME_ROWS); y < WICK_Y; y++) {
  for (let x = WICK_X - 4; x <= WICK_X + 4; x++) hidden[y][x] = true
}

// bricks behind everything, staggered every other course
for (let y = 0; y < SHELF_Y; y++) {
  const off = Math.floor(y / 2) % 2 ? 5 : 0
  for (let x = 0; x < COLS; x++) {
    if (hidden[y][x] || cells[y][x]) continue
    put(x, y, y % 2 === 0 ? '_' : (x + off) % 10 === 0 ? '|' : ' ', 'wall')
  }
}

// ------------------------------------------------------------------ the flame
// Its strength wanders to a new level now and then, with a slow sway on top,
// and it leans left and right. `still` holds it upright at full strength.
export class Flame {
  private seed = 7
  private target = 1
  private level = 1
  private leanTarget = 0
  power = 1
  lean = 0

  private rand() {
    this.seed = (this.seed * 16807) % 2147483647
    return this.seed / 2147483647
  }

  step(t: number) {
    if (this.rand() < 0.09) this.target = 0.74 + this.rand() * 0.32
    if (this.rand() < 0.05) this.leanTarget = this.rand() * 2 - 1
    this.level += (this.target - this.level) * 0.2
    this.lean += (this.leanTarget - this.lean) * 0.12
    this.power = this.level + Math.sin(t * 0.0011) * 0.04 + Math.sin(t * 0.0041) * 0.025
  }
}

// light reaching (x, y) from the flame, falling off over `reach` cell heights
// (a cell is 0.6 wide); it reaches further when the flame is high
function lightAt(x: number, y: number, f: Flame, reach: number) {
  const fx = WICK_X + f.lean
  const fy = WICK_Y - FLAME_ROWS * 0.4
  const d = Math.hypot((x - fx) * 0.6, y - fy) / (reach * f.power)
  return f.power * Math.exp(-d * d)
}

const shade = (i: number) => RAMP[Math.max(1, Math.min(RAMP.length - 1, Math.round(i * (RAMP.length - 1))))]

// ------------------------------------------------------------------ drawing
// Draws one frame. `cw` and `size` are a cell's width and height in pixels;
// `rgb` is the colour of the text, as "r,g,b".
export function draw(ctx: CanvasRenderingContext2D, f: Flame, t: number, cw: number, size: number, rgb: string) {
  // the candle frame rendered closest to where the flame leans
  const frame = wax.frames[Math.round(((f.lean + 1) / 2) * (wax.frames.length - 1))]
  for (let y = 0; y < ROWS; y++) {
    for (let x = 0; x < COLS; x++) {
      const cell = cells[y][x]
      if (!cell) continue
      let L: number
      let ch: string
      if (cell.kind === 'candle') {
        // a curve that deepens the shadows, so each drip keeps its edge
        const b = Math.pow(DIGITS.indexOf(frame[cell.r][cell.c]) / 35, 1.35) * (0.8 + 0.3 * f.power)
        ch = shade(b)
        L = 0.3 + 0.7 * Math.min(1, b)
      } else if (cell.kind === 'skull') {
        // how far the flame is, and how squarely the bone faces it; hollows
        // stay dark whichever way it leans
        const fx = WICK_X + f.lean * 1.2
        const fy = WICK_Y - FLAME_ROWS * (0.3 + 0.15 * f.power)
        const lx = (fx - x) * 0.6
        const ly = fy - y
        const lz = 10
        const facing = Math.max(0, (cell.nx * lx + cell.ny * ly + cell.nz * lz) / Math.hypot(lx, ly, lz))
        L = lightAt(x, y, f, 46) * (0.16 + 0.95 * facing) * (1 - 0.85 * cell.hollow)
        // the brighter the bone, the finer the strokes that show on it
        let lvl = L > 0.5 ? 2 : L > 0.3 ? 1 : L > 0.13 ? 0 : -1
        ch = ' '
        for (; lvl >= 0 && ch === ' '; lvl--) ch = cell.chars[lvl]
        if (ch === ' ') continue
        L += 0.2
      } else if (cell.kind === 'wall') {
        // a halo of bricks around the flame, fading into the dark
        L = lightAt(x, y, f, 13) * 0.5
        if (L < 0.05) continue
        ch = cell.ch
      } else {
        L = lightAt(x, y, f, 40) * 0.8
        if (L < 0.05) continue
        ch = cell.ch
      }
      ctx.fillStyle = `rgba(${rgb},${Math.max(0, Math.min(1, L)).toFixed(3)})`
      ctx.fillText(ch, x * cw, y * size)
    }
  }
  drawFlame(ctx, f, t, cw, size)
}

// A teardrop on the wick: wide and dim at the base, a bright core low in it,
// narrowing to a tip that trails behind the lean.
function drawFlame(ctx: CanvasRenderingContext2D, f: Flame, t: number, cw: number, size: number) {
  const h = FLAME_ROWS * (0.72 + 0.28 * f.power)
  const baseY = WICK_Y - 0.4
  for (let y = Math.floor(baseY - h); y <= Math.ceil(baseY); y++) {
    const u = (baseY - y) / h // 0 at the base, 1 at the tip
    if (u < 0 || u > 1) continue
    const halfW = 2.6 * Math.pow(Math.sin(Math.PI * Math.min(1, u * 1.15 + 0.08)), 0.9) * (1 - u) ** 0.35
    const cx = WICK_X + f.lean * 1.6 * u * u + Math.sin(t * 0.009 + u * 5) * 0.35 * u
    for (let x = Math.floor(cx - halfW - 1); x <= Math.ceil(cx + halfW + 1); x++) {
      const v = Math.abs(x - cx) / Math.max(halfW, 0.3)
      if (v > 1) continue
      const core = Math.exp(-(((u - 0.32) / 0.32) ** 2)) * (1 - v * v)
      let I = 0.35 * (1 - v) + 0.75 * core
      if (u < 0.08) I *= 0.5
      I = Math.min(1, I * (0.85 + 0.25 * f.power))
      ctx.fillStyle = `rgba(250,250,250,${(0.55 + 0.45 * I).toFixed(3)})`
      ctx.fillText(shade(I), x * cw, y * size)
    }
  }
}
