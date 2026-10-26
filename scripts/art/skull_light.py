"""The skull for the candle rest screen: three levels of detail traced from a
model picture, and a relief of the bone so the page can light it from wherever
the flame is (web/src/lib/candle.ts).

    pip install numpy pillow
    python scripts/art/skull_light.py MODEL.png web/src/lib/candle-skull.json [cols] [preview.png]

MODEL.png is the line-art skull the drawing is traced from (three-quarter view,
face to the right, mandible below row 168); it is not kept in the repository.

Writes skull.json with, on one grid of character cells (mirrored, the face to
the left):
  levels  three drawings, from the sparsest strokes to the finest detail
  nx, ny  which way the bone faces in each cell (base 36, 0..35 -> -1..1)
  hollow  how deep in a hollow (orbit, nose, gaps) each cell is (0..35)
"""
import json
import sys
from pathlib import Path

import numpy as np
from PIL import Image, ImageDraw, ImageFilter, ImageFont

REF = Path(sys.argv[1])
FONT = str(Path(__file__).parent / "../../web/node_modules/@fontsource/jetbrains-mono/files/jetbrains-mono-latin-400-normal.woff")
COLS = int(sys.argv[3]) if len(sys.argv) > 3 else 84
SX, SY = 6, 10
CHARS = r""" .,'`-_/\|()"""
GAINS = [1.4, 1.95, 2.7]
DIG = "0123456789abcdefghijklmnopqrstuvwxyz"
SWAP = str.maketrans("/\\()[]{}<>", "\\/)(][}{><")

# ---------------------------------------------------------------- the model
im = np.asarray(Image.open(REF).convert("L"), float)
im = im[:, 5:]
im = np.clip((im - np.median(im) - 6) / 120.0, 0, 1)
H, W = im.shape
yy, xx = np.mgrid[0:H, 0:W]
xx = xx + 5
keep = (yy <= 139) | ((xx >= 76) & (yy <= 150)) | ((xx >= 108) & (yy <= 168))
im = im * keep
ys, _ = np.nonzero(im > 0.05)
im = im[: ys.max() + 2, :]
H, W = im.shape
cw = W / COLS
ROWS = int(round(H / (cw / 0.6)))


def to_img(a):
    return Image.fromarray((np.clip(a, 0, 1) * 255).astype(np.uint8))


def from_img(i):
    return np.asarray(i, float) / 255.0


def blur(a, r):
    return from_img(to_img(a).filter(ImageFilter.GaussianBlur(r)))


# ---------------------------------------------------------------- strokes
def glyphs():
    size = 40
    font = ImageFont.truetype(FONT, size)
    w = int(font.getlength("M"))
    asc, desc = font.getmetrics()
    top = (asc + desc - size) / 2
    out = []
    for c in CHARS:
        g = Image.new("L", (w, size), 0)
        ImageDraw.Draw(g).text((0, -top), c, font=font, fill=255)
        a = np.asarray(g, float) / 255.0
        a = a[: size // SY * SY, : w // SX * SX]
        out.append(a.reshape(SY, a.shape[0] // SY, SX, a.shape[1] // SX).mean(axis=(1, 3)).ravel())
    return np.array(out)


GL = glyphs()
UP = np.asarray(to_img(blur(im, 0.9)).resize((COLS * SX, ROWS * SY), Image.BILINEAR), float) / 255.0
CELLS = UP.reshape(ROWS, SY, COLS, SX).transpose(0, 2, 1, 3).reshape(ROWS * COLS, SY * SX)


def trace(gain):
    t = np.clip(CELLS * gain, 0, 1)
    pick = ((GL[None] - t[:, None]) ** 2).sum(-1).argmin(1).reshape(ROWS, COLS)
    ink = pick > 0
    n = sum(np.roll(np.roll(ink, dy, 0), dx, 1) for dy in (-1, 0, 1) for dx in (-1, 0, 1) if dy or dx)
    pick[ink & (n < 2)] = 0
    return ["".join(CHARS[i] for i in r) for r in pick]


levels = [trace(g) for g in GAINS]

# ---------------------------------------------------------------- relief
# silhouette: each row filled between its first and last mark
ink = blur(im, 3.0) > 0.05
sil = np.zeros_like(ink)
for r in range(H):
    cols = np.nonzero(ink[r])[0]
    if len(cols):
        sil[r, cols[0]:cols[-1] + 1] = True
sil = blur(sil.astype(float), 1.5) > 0.5
# the dome: the silhouette blurred several times rises toward its middle
dome = sum(blur(sil.astype(float), r) for r in (4, 9, 16, 26)) / 4.0
dome = dome * sil
# hollows, measured on the model (pixels, before the 5 px crop): the two
# orbits, the nasal opening and the gap under the cheekbone arch
HOLLOWS = [  # centre x, centre y, half width, half height, depth
    (74, 77, 30, 23, 1.0),     # near orbit
    (154, 80, 18, 21, 0.9),    # far orbit
    (126, 116, 11, 12, 1.0),   # nasal opening
    (45, 148, 26, 6, 0.6),     # under the arch
]
hollow = np.zeros_like(im)
for hx, hy, hw, hh, depth in HOLLOWS:
    e = ((xx[: H, :] - hx) / hw) ** 2 + ((yy[: H, :] - hy) / hh) ** 2
    hollow = np.maximum(hollow, depth * np.clip(1.0 - e, 0, 1) ** 0.6)
hollow = blur(hollow, 2.0) * sil
height = dome - 0.55 * hollow


def cellify(a):
    return np.asarray(to_img(a).resize((COLS, ROWS), Image.BOX), float) / 255.0


hc = cellify(np.clip(height, 0, 1))
holc = cellify(hollow)
silc = cellify(sil.astype(float))
gy, gx = np.gradient(hc)
gx = gx / 0.6          # a cell is 0.6 wide: slopes across are steeper
k = 9.0
n = np.stack([-gx * k, -gy * k, np.ones_like(hc)], -1)
n /= np.linalg.norm(n, axis=-1, keepdims=True)


def q36(v, lo, hi):
    return DIG[int(round((np.clip(v, lo, hi) - lo) / (hi - lo) * 35))]


# ---------------------------------------------------------------- mirror and write
def flip_text(rows):
    return [r[::-1].translate(SWAP) for r in rows]


out = {
    "cols": COLS,
    "rows": ROWS,
    "levels": [flip_text(l) for l in levels],
    "nx": ["".join(q36(-v, -1, 1) for v in row[::-1]) for row in n[..., 0]],
    "ny": ["".join(q36(v, -1, 1) for v in row[::-1]) for row in n[..., 1]],
    "hollow": ["".join(q36(v, 0, 1) for v in row[::-1]) for row in holc],
    "inside": ["".join("1" if v > 0.5 else "0" for v in row[::-1]) for row in silc],
}
Path(sys.argv[2]).write_text(json.dumps(out), encoding="utf-8")

# a debug picture: relief, hollows, and the light from the left
if len(sys.argv) > 4:
    L = np.array([-0.7, -0.2, 0.68])
    lam = np.clip((n * L).sum(-1), 0, 1) * (1 - 0.8 * holc) * (silc > 0.5)
    tiles = [hc, holc, lam]
    w, h = COLS * 4, ROWS * 6
    pic = Image.new("L", (w * 3 + 20, h), 0)
    for i, t in enumerate(tiles):
        pic.paste(to_img(t[:, ::-1]).resize((w, h), Image.NEAREST), (i * (w + 10), 0))
    pic.save(sys.argv[4])
print(ROWS, "rows", COLS, "cols")
