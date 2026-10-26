"""A melted candle in 3D, lit by its own flame, turned into brightness grids,
for the candle rest screen (web/src/lib/candle.ts).

    pip install numpy pillow
    python scripts/art/candle3d.py web/src/lib/candle-wax.json [preview.png]

For a few positions of the swaying flame it writes the brightness (0..1, base
36) of every character cell, a mask of where the candle is and the cell of the
wick. The page picks the grid that matches the flame and turns brightness into
characters. It is seen from a little below the middle, like the skull.
"""
import json
import math
import sys
from pathlib import Path

import numpy as np

sys.path.insert(0, str(Path(__file__).parent))
from scene import fbm, vnoise, length, norm, smin, smax  # noqa: E402

COLS, ROWS = 34, 46
SX, SY = 3, 5
TOP = 5.6          # rim height; radius of the candle is 1
SWAYS = [-0.14, -0.07, 0.0, 0.07, 0.14]
# Seen from a little below the middle, looking slightly up, like the skull:
# the rim shows as an edge from underneath and the hollow stays out of sight.
CAM = np.array([0.0, 2.2, 15.0])
LOOK = np.array([0.0, 3.0, 0.0])
FOV = 30

# drips over the rim: angle (radians, pi/2 faces us), length, thickness
DRIPS = [
    (0.55, 2.4, 0.27), (0.98, 1.1, 0.22), (1.32, 3.8, 0.31), (1.7, 1.6, 0.25),
    (2.05, 4.4, 0.32), (2.45, 2.2, 0.27), (2.8, 3.2, 0.26), (0.2, 1.4, 0.2),
]


def sd_capsule_y(p, x, z, y0, y1, r):
    """A vertical capsule from y0 down to y1; x and z may vary per point."""
    yy = np.clip(p[:, 1], y1, y0)
    return np.hypot(np.hypot(p[:, 0] - x, p[:, 2] - z), p[:, 1] - yy) - r


def rim_height(theta):
    # melted lower on the right, lumpy all round
    return TOP - 0.28 * np.cos(theta) + 0.08 * np.sin(theta * 5.0) + 0.05 * np.sin(theta * 11.0 + 1.0)


def sd_candle(p):
    x, y, z = p[:, 0], p[:, 1], p[:, 2]
    r = np.hypot(x, z)
    th = np.arctan2(z, x)
    # body: a slightly uneven cylinder, cut at the melted rim
    wobble = 0.03 * vnoise(np.stack([th * 3, y * 2, np.zeros_like(y)], -1))
    d = r - (1.0 + wobble)
    d = np.maximum(d, -y)
    d = np.maximum(d, y - rim_height(th))
    # the hollow the flame has eaten into the top, with a pool of wax
    bowl = length((p - np.array([0.0, TOP + 0.45, 0.0])) / np.array([0.8, 0.6, 0.8])) - 1.0
    d = smax(d, -bowl * 0.6, 0.08)
    # a lip of wax rolled over the rim
    lip = np.hypot(r - 0.97, y - (rim_height(th) - 0.05)) - (0.1 + 0.04 * np.sin(th * 7))
    d = smin(d, lip, 0.08)
    # drips: wax run down the side, each ending in a bead
    for a, ln, w in DRIPS:
        cx, cz = 1.0 * math.cos(a), 1.0 * math.sin(a)
        top = TOP - 0.28 * math.cos(a)
        wig = 0.04 * np.sin(y * 3.0 + a * 5)
        # the run thins as it goes down, then gathers into a bead
        out = 1.0 + w * 0.45
        k = np.clip((top - p[:, 1]) / ln, 0, 1)
        cap = sd_capsule_y(p, cx * out + wig * -math.sin(a), cz * out + wig * math.cos(a), top, top - ln, 0) - w * (1 - 0.4 * k)
        bead = length((p - np.array([cx * (out + 0.03), top - ln - w * 0.2, cz * (out + 0.03)])) / np.array([1.0, 1.35, 1.0])) - w * 1.1
        d = smin(d, smin(cap, bead, 0.16), 0.07)
    # a puddle of wax spread round the foot
    pud = np.maximum(r - (1.55 + 0.15 * np.sin(th * 3 + 0.7) + 0.08 * np.sin(th * 7)), y - 0.14)
    pud = np.maximum(pud, -y - 0.02)
    d = smin(d, pud, 0.12)
    return d


def sd_wick(p):
    # a short wick, curled a little at the tip
    t = np.clip((p[:, 1] - (TOP - 0.1)) / 0.65, 0, 1)
    cx = 0.12 * t * t
    return np.maximum(np.hypot(p[:, 0] - cx, p[:, 2]) - 0.045, np.abs(p[:, 1] - (TOP + 0.22)) - 0.33)


def scene(p):
    dc, dw = sd_candle(p), sd_wick(p)
    return np.minimum(dc, dw), dw < dc


def normal(p):
    e = 2e-3
    n = np.zeros_like(p)
    for i in range(3):
        o = np.zeros(3)
        o[i] = e
        n[:, i] = scene(p + o)[0] - scene(p - o)[0]
    return n / np.maximum(length(n)[:, None], 1e-9)


def occlusion(p, n):
    occ = np.zeros(len(p))
    for i in range(1, 6):
        h = 0.05 * i
        occ += (h - scene(p + n * h)[0]) * (0.6 ** i)
    return np.clip(1.0 - 4.5 * occ, 0.0, 1.0)


def render(sway):
    W, H = COLS * SX, ROWS * SY
    cam, look = CAM, LOOK
    fwd = norm(look - cam)
    right = norm(np.cross(fwd, [0, 1, 0]))
    up = np.cross(right, fwd)
    half_h = math.tan(math.radians(FOV) / 2)
    half_w = half_h * W / H
    xs = (np.arange(W) + 0.5) / W * 2 - 1
    ys = 1 - (np.arange(H) + 0.5) / H * 2
    gx, gy = np.meshgrid(xs * half_w, ys * half_h)
    rd = fwd + gx.reshape(-1, 1) * right + gy.reshape(-1, 1) * up
    rd /= length(rd)[:, None]
    t = np.zeros(len(rd))
    alive = np.ones(len(rd), bool)
    hit = np.zeros(len(rd), bool)
    for _ in range(120):
        idx = np.nonzero(alive)[0]
        if not len(idx):
            break
        d = scene(cam + rd[idx] * t[idx, None])[0]
        t[idx] += d * 0.8
        done = d < 1e-3
        hit[idx[done]] = True
        alive[idx[done | (t[idx] > 30)]] = False
    img = np.zeros(len(rd))
    idx = np.nonzero(hit)[0]
    p = cam + rd[idx] * t[idx, None]
    _, wick = scene(p)
    n = normal(p)
    flame = np.array([sway, TOP + 1.0, 0.0])
    tol = flame - p
    dist = length(tol)
    l = tol / dist[:, None]
    diff = np.clip((n * l).sum(-1), 0, 1)
    fall = 1.0 / (1.0 + (dist / 3.6) ** 2)
    # wax lets light through near the flame: the top glows
    glow = 0.55 * np.exp(-np.clip(TOP - p[:, 1], 0, None) / 0.9) * (1 - wick)
    # light bounced back from the room in front, weak, so every drip keeps
    # its volume even below the flame
    bounce = 0.34 * np.clip((n * norm(np.array([-0.45, 0.25, 0.85]))).sum(-1), 0, 1)
    ao = occlusion(p, n)
    amb = 0.05 + 0.04 * n[:, 1].clip(0)
    grain = 0.05 * fbm(p * 6.0, 2)
    b = ((0.95 * diff * fall * 1.9 + bounce) * (0.45 + 0.55 * ao) + glow + amb + grain) * np.where(wick, 0.25, 1.0)
    img[idx] = b
    cells = img.reshape(ROWS, SY, COLS, SX).mean(axis=(1, 3))
    cover = hit.reshape(ROWS, SY, COLS, SX).mean(axis=(1, 3))
    # local contrast: each cell against its neighbours, so a drip keeps its
    # edge even where the light is even
    pad = np.pad(cells, 1, mode="edge")
    around = sum(pad[1 + dy:1 + dy + ROWS, 1 + dx:1 + dx + COLS] for dy in (-1, 0, 1) for dx in (-1, 0, 1)) / 9
    cells = np.where(cover > 0.3, cells + 1.1 * (cells - around), cells)
    if sway == 0.0 and len(sys.argv) > 2:
        from PIL import Image
        Image.fromarray((np.clip(img, 0, 1).reshape(H, W) * 255).astype(np.uint8)).resize((W * 2, H * 2)).save(sys.argv[2])
    return np.clip(cells, 0, 1), cover


def main():
    out = {"cols": COLS, "rows": ROWS, "sways": SWAYS, "frames": []}
    cover = None
    for s in SWAYS:
        cells, cv = render(s)
        cover = cv if cover is None else np.maximum(cover, cv)
        out["frames"].append(["".join("0123456789abcdefghijklmnopqrstuvwxyz"[int(round(v * 35))] for v in row) for row in cells])
        print("sway", s, "done")
    out["mask"] = ["".join("1" if v > 0.3 else "0" for v in row) for row in cover]
    # where the wick ends, in cells: the flame stands on it
    out["wick"] = wick_cell()
    Path(sys.argv[1]).write_text(json.dumps(out), encoding="utf-8")


def wick_cell():
    # project the wick tip with the same camera
    cam, look = CAM, LOOK
    fwd = norm(look - cam)
    right = norm(np.cross(fwd, [0, 1, 0]))
    up = np.cross(right, fwd)
    half_h = math.tan(math.radians(FOV) / 2)
    half_w = half_h * (COLS * SX) / (ROWS * SY)
    tip = np.array([0.12, TOP + 0.55, 0.0]) - cam
    zc = (tip * fwd).sum()
    sx = (tip * right).sum() / zc / half_w
    sy = (tip * up).sum() / zc / half_h
    return [round((sx + 1) / 2 * COLS - 0.5), round((1 - sy) / 2 * ROWS - 0.5)]


if __name__ == "__main__":
    main()
