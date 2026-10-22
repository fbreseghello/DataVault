"""Renders the rest screen: a tall broken Doric column, in 3D, then turns the
image into ASCII.

Everything is modelled in metres at real scale (the shaft is about 0.9 m across
and breaks off near 3 m) and raymarched with a signed distance field: soft
shadows, ambient occlusion, weathered stone. The break is ragged, with chips torn
out below it, cracks zigzag down the lit face and pieces of it lie on the steps
and the ground. The ground fades to black like a spotlight.

The character ramp is ordered by ink coverage measured in JetBrains Mono, read
from web/node_modules (run npm install in web first). Rendering takes a few
minutes.

    pip install numpy pillow
    python scripts/art/scene.py web/src/lib/pillar.txt [preview.pgm]

The optional second argument writes the grey image behind the ASCII (PGM).
"""
import math
import os
import sys

import numpy as np

COLS, ROWS = 130, 110          # output size in characters
SX, SY = 3, 5                 # samples per character (a cell is 0.6 x 1)
STEPS, EPS, FAR = 160, 4e-4, 20.0


# ---------------------------------------------------------------- helpers

def norm(v):
    return v / np.linalg.norm(v)


def length(p):
    return np.sqrt((p * p).sum(-1))


def smin(a, b, k):
    h = np.clip(0.5 + 0.5 * (b - a) / k, 0.0, 1.0)
    return b + (a - b) * h - k * h * (1.0 - h)


def smax(a, b, k):
    return -smin(-a, -b, k)


def sd_box(p, b):
    q = np.abs(p) - b
    return length(np.maximum(q, 0.0)) + np.minimum(q.max(-1), 0.0)


def hash3(i):
    h = (i[..., 0] * 374761393 + i[..., 1] * 668265263 + i[..., 2] * 2147483647) & 0xFFFFFFFF
    h = (h ^ (h >> 13)) * 1274126177 & 0xFFFFFFFF
    return ((h ^ (h >> 16)) & 0xFFFF) / 65535.0


def vnoise(p):
    """Smooth value noise in 3D, 0..1."""
    i = np.floor(p).astype(np.int64)
    f = p - i
    f = f * f * (3.0 - 2.0 * f)
    out = 0.0
    for dx in (0, 1):
        for dy in (0, 1):
            for dz in (0, 1):
                w = (f[..., 0] if dx else 1 - f[..., 0]) * (f[..., 1] if dy else 1 - f[..., 1]) * (f[..., 2] if dz else 1 - f[..., 2])
                out = out + w * hash3(i + np.array([dx, dy, dz]))
    return out


def fbm(p, octaves=4):
    v, a = 0.0, 0.5
    for _ in range(octaves):
        v = v + a * (vnoise(p) - 0.5)
        p = p * 2.03 + 1.7
        a *= 0.5
    return v


# ---------------------------------------------------------------- the column

CX, CZ = 0.0, 0.0
STEPS_BOX = [(0.68, 0.00, 0.09), (0.60, 0.09, 0.18), (0.52, 0.18, 0.27)]   # half width, bottom, top
SHAFT_BOTTOM = 0.27
R_BOTTOM, R_TOP = 0.44, 0.40
FLUTES = 20
FLUTE_DEPTH = 0.028


def column_top(qx, qz):
    """Height of the fracture: higher on the left, falling away to the right,
    with a deep bite out of the front and a rough, stepped edge."""
    u = qx / R_BOTTOM
    flat = np.stack([qx, qz, np.zeros_like(qx)], -1)
    bite = 0.42 * np.exp(-((qx - 0.08) ** 2 + (qz - 0.3) ** 2) / 0.03)
    return (2.72 - 0.34 * u - 0.42 * np.maximum(u - 0.1, 0) - bite
            + 0.5 * fbm(flat * 2.2 + 3.0, 3) + 0.12 * fbm(flat * 11.0, 2))


# Cracks on the lit face the camera sees, as paths of (angle around the shaft,
# height) and a half width in metres. Each one is a groove a few cm deep.
CRACKS = [
    # the main crack, zigzagging down from the break across the flutes
    ([(2.02, 2.56), (1.76, 2.36), (1.96, 2.16), (1.62, 1.96), (1.86, 1.76), (1.56, 1.52), (1.76, 1.3), (1.62, 1.1)], 0.015),
    # its branches
    ([(1.62, 1.96), (1.38, 1.86), (1.3, 1.64)], 0.01),
    ([(1.86, 1.76), (2.14, 1.62), (2.3, 1.42), (2.5, 1.32)], 0.011),
    # a shorter one near the top, and one in the lowest drum
    ([(1.36, 2.32), (1.52, 2.12), (1.32, 1.96)], 0.01),
    ([(2.24, 1.04), (2.06, 0.86), (2.22, 0.62)], 0.009),
]

# Chips torn out of the shaft near the break: shallow scoops, wide and ragged.
# (angle, height, half width, half height, depth)
SPALLS = [(2.05, 2.3, 0.15, 0.1, 0.05), (1.5, 2.08, 0.1, 0.07, 0.04), (2.5, 2.48, 0.12, 0.09, 0.045)]

# Fallen pieces on the steps and on the ground, angular: centre, half sizes,
# turn around the vertical and tilt (radians).
RUBBLE = [
    ((0.32, 0.33, 0.42), (0.08, 0.05, 0.06), 0.5, 0.3),
    ((-0.27, 0.31, 0.46), (0.05, 0.035, 0.045), -0.4, -0.2),
    ((0.82, 0.08, 0.74), (0.14, 0.08, 0.1), 0.9, 0.2),
    ((-0.78, 0.05, 0.62), (0.08, 0.05, 0.07), -0.7, 0.35),
    ((0.6, 0.04, 1.05), (0.05, 0.035, 0.04), 0.2, -0.4),
    ((-0.5, 0.03, 0.95), (0.04, 0.03, 0.035), 1.1, 0.5),
]


def seg_dist(px, py, ax, ay, bx, by):
    vx, vy = bx - ax, by - ay
    t = np.clip(((px - ax) * vx + (py - ay) * vy) / (vx * vx + vy * vy), 0, 1)
    return np.hypot(px - ax - vx * t, py - ay - vy * t)


def crack_groove(q):
    """How far each point is from the nearest crack, measured along the shaft
    surface (arc length and height); negative inside a crack."""
    yy = q[:, 1]
    # the edges wander a little, so a crack is never drawn with a ruler
    arc = np.arctan2(q[:, 2], q[:, 0]) * R_BOTTOM + 0.008 * fbm(np.stack([yy * 50, q[:, 0] * 50, q[:, 2] * 50], -1), 2)
    ragged = 0.55 + 0.9 * vnoise(np.stack([arc * 60, yy * 60, np.zeros_like(yy)], -1))
    groove = np.full(len(q), 1e9)
    for path, w in CRACKS:
        n = len(path) - 1
        for i, ((a0, y0), (a1, y1)) in enumerate(zip(path, path[1:])):
            dd = seg_dist(arc, yy, a0 * R_BOTTOM, y0, a1 * R_BOTTOM, y1)
            # widest where it starts, a hairline at its tip
            taper = 1.0 - 0.75 * (i + 0.5) / n
            groove = np.minimum(groove, dd - w * taper * ragged)
    return groove


def sd_ellipsoid(p, r):
    k0 = length(p / r)
    k1 = length(p / (r * r))
    return k0 * (k0 - 1.0) / np.maximum(k1, 1e-9)


def weathered(d, q, amount, scale):
    """Pits and chipped edges, only near the surface where they show."""
    near = np.abs(d) < 3 * amount
    if near.any():
        d = d.copy()
        d[near] += amount * (fbm(q[near] * scale, 3) + 0.5 * np.maximum(fbm(q[near] * scale * 0.5 + 7, 2), 0))
    return d


def sd_column(p):
    q = p - np.array([CX, 0.0, CZ])
    x, y, z = q[:, 0], q[:, 1], q[:, 2]
    d = np.full(len(p), 1e9)
    for hw, b, t in STEPS_BOX:
        c = np.array([0.0, (b + t) / 2, 0.0])
        d = np.minimum(d, sd_box(q - c, np.array([hw, (t - b) / 2 - 0.006, hw])) - 0.006)
    d = weathered(d, q, 0.012, 9.0)
    r2 = np.sqrt(x * x + z * z)
    th = np.arctan2(z, x)
    t = np.clip((y - SHAFT_BOTTOM) / 3.0, 0, 1)
    rad = R_BOTTOM + (R_TOP - R_BOTTOM) * t
    phase = (th * FLUTES / (2 * np.pi)) % 1.0
    channel = FLUTE_DEPTH * np.sqrt(np.maximum(0.0, 1.0 - (phase * 2 - 1) ** 2))
    d_cyl = r2 - (rad - channel)
    d_shaft = np.maximum(d_cyl, SHAFT_BOTTOM - y)
    # the joints between the drums
    for joint in (1.08, 1.9):
        d_shaft = smax(d_shaft, -(np.abs(y - joint) - 0.005), 0.004)
    # cracks, cut 6 cm into the shaft
    near = (r2 > rad - 0.09) & (r2 < rad + 0.05)
    if near.any():
        groove = crack_groove(q[near])
        cut = np.maximum(groove, (rad[near] - 0.06) - r2[near])
        d_shaft = d_shaft.copy()
        d_shaft[near] = smax(d_shaft[near], -cut, 0.002)
    # chips torn out near the break: shallow, ragged, rough inside
    for a, h, hw, hh, depth in SPALLS:
        c = np.array([R_BOTTOM * math.cos(a), h, R_BOTTOM * math.sin(a)])
        n_out = np.array([math.cos(a), 0.0, math.sin(a)])
        tang = np.array([-math.sin(a), 0.0, math.cos(a)])
        rel = q - c
        local_p = np.stack([(rel * tang).sum(-1), rel[:, 1], (rel * n_out).sum(-1)], -1)
        hole = sd_ellipsoid(local_p, np.array([hw, hh, depth])) - 0.03 * fbm(rel * 14.0 + 5.0, 3)
        d_shaft = smax(d_shaft, -hole, 0.006)
    d_shaft = np.maximum(d_shaft, (y - column_top(x, z)) * 0.6)
    d_shaft = weathered(d_shaft, q, 0.007, 14.0)
    # the fallen pieces
    rubble = np.full(len(p), 1e9)
    for c, half, turn, tilt in RUBBLE:
        rel = q - np.array(c)
        ct, st = math.cos(turn), math.sin(turn)
        rx, rz = ct * rel[:, 0] + st * rel[:, 2], -st * rel[:, 0] + ct * rel[:, 2]
        cl, sl = math.cos(tilt), math.sin(tilt)
        ry, rx = cl * rel[:, 1] - sl * rx, sl * rel[:, 1] + cl * rx
        rubble = np.minimum(rubble, sd_box(np.stack([rx, ry, rz], -1), np.array(half)) - 0.006)
    rubble = weathered(rubble, q, 0.018, 9.0)
    return np.minimum(np.minimum(d, d_shaft), rubble)


# ---------------------------------------------------------------- scene

GROUND, COLUMN = 0, 1


def scene(p):
    dg = p[:, 1]
    dc = sd_column(p)
    d = np.minimum(dg, dc)
    mat = np.where(d == dc, COLUMN, GROUND)
    return d, mat


def dist(p):
    return scene(p)[0]


def normal(p):
    e = 3e-4
    n = np.zeros_like(p)
    for i in range(3):
        o = np.zeros(3)
        o[i] = e
        n[:, i] = dist(p + o) - dist(p - o)
    return n / np.maximum(length(n)[:, None], 1e-9)


def soft_shadow(p, l):
    res = np.ones(len(p))
    t = np.full(len(p), 0.01)
    for _ in range(48):
        h = dist(p + l * t[:, None])
        res = np.minimum(res, 10.0 * h / t)
        t = t + np.clip(h, 0.004, 0.25)
    return np.clip(res, 0.0, 1.0)


def occlusion(p, n):
    occ = np.zeros(len(p))
    for i in range(1, 6):
        h = 0.012 * i
        occ = occ + (h - dist(p + n * h)) * (0.65 ** i)
    return np.clip(1.0 - 9.0 * occ, 0.0, 1.0)


def render(cam=(0.45, 1.25, 5.4), look=(0.0, 1.42, 0.0), fov_deg=34, cols=COLS, rows=ROWS):
    W, H = cols * SX, rows * SY
    cam = np.array(cam, float)
    look = np.array(look, float)
    fwd = norm(look - cam)
    right = norm(np.cross(fwd, [0, 1, 0]))
    up = np.cross(right, fwd)
    fov = math.radians(fov_deg)
    half_h = math.tan(fov / 2)
    half_w = half_h * (W / H)        # samples are square: a cell is 0.6 wide, 1 tall
    xs = (np.arange(W) + 0.5) / W * 2 - 1
    ys = 1 - (np.arange(H) + 0.5) / H * 2
    gx, gy = np.meshgrid(xs * half_w, ys * half_h)
    rd = fwd + gx.reshape(-1, 1) * right + gy.reshape(-1, 1) * up
    rd = rd / length(rd)[:, None]
    ro = np.broadcast_to(cam, rd.shape).copy()

    t = np.zeros(len(rd))
    alive = np.ones(len(rd), bool)
    hit = np.zeros(len(rd), bool)
    for _ in range(STEPS):
        idx = np.nonzero(alive)[0]
        if not len(idx):
            break
        d = dist(ro[idx] + rd[idx] * t[idx, None])
        t[idx] += d * 0.75
        done = d < EPS * (1 + t[idx])
        hit[idx[done]] = True
        alive[idx[done | (t[idx] > FAR)]] = False

    img = np.zeros(len(rd))
    idx = np.nonzero(hit)[0]
    p = ro[idx] + rd[idx] * t[idx, None]
    _, mat = scene(p)
    n = normal(p)
    light = norm(np.array([-0.8, 0.62, 0.2]))
    diff = np.clip((n * light).sum(-1), 0, 1)
    sh = soft_shadow(p + n * 0.002, light)
    ao = occlusion(p, n)
    grain = fbm(p * 38.0, 3)
    albedo = np.where(mat == COLUMN, 0.8 + 0.12 * grain, 0.1 + 0.06 * grain)
    # the inside of a crack is dark: little light gets that deep
    q = p - np.array([CX, 0.0, CZ])
    inside = (mat == COLUMN) & (q[:, 1] > SHAFT_BOTTOM)
    if inside.any():
        g = crack_groove(q[inside])
        albedo[inside] *= np.clip(0.18 + g / 0.012, 0.18, 1.0)
    spec = np.clip((n * norm(light - rd[idx])).sum(-1), 0, 1) ** 30 * (mat != GROUND)
    b = albedo * (0.10 * ao + 0.90 * diff * sh * (0.55 + 0.45 * ao)) + 0.18 * spec * sh
    # spotlight: the ground fades out away from the ruins
    r = np.hypot(p[:, 0] - CX, p[:, 2] - CZ - 0.3)
    b = np.where(mat == GROUND, b * np.exp(-(r / 1.1) ** 2), b)
    img[idx] = b
    return img.reshape(H, W)


FONT = os.path.join(os.path.dirname(__file__), "../../web/node_modules/@fontsource/jetbrains-mono/files/jetbrains-mono-latin-400-normal.woff")
CHARS = " .:-=+*#%@"


def glyphs():
    """Each character as SX x SY coverage values, drawn the way the page draws
    it: JetBrains Mono, line-height 1 (the em box is centred on the line)."""
    from PIL import Image, ImageDraw, ImageFont
    size = 40
    font = ImageFont.truetype(FONT, size)
    w = int(font.getlength("M"))
    ascent, descent = font.getmetrics()
    top = (ascent + descent - size) / 2
    out = []
    for ch in CHARS:
        im = Image.new("L", (w, size), 0)
        ImageDraw.Draw(im).text((0, -top), ch, font=font, fill=255)
        a = np.asarray(im, float) / 255.0
        a = a[: size // SY * SY, : w // SX * SX]
        out.append(a.reshape(SY, a.shape[0] // SY, SX, a.shape[1] // SX).mean(axis=(1, 3)).ravel())
    return np.array(out)


def tone(img):
    """Black point and a little gamma, so the dark ground goes to nothing."""
    b = np.clip((img - 0.04) / 0.8, 0, 1)
    return b ** 0.75


def to_ascii(img):
    """Each cell takes the character whose ink coverage, measured in the font
    the page uses, is closest to the brightness under it."""
    cov = glyphs().mean(1)
    order = np.argsort(cov)
    chars = [CHARS[i] for i in order]
    levels = cov[order] / cov.max()
    cells = tone(img).reshape(ROWS, SY, COLS, SX).mean(axis=(1, 3))
    pick = np.abs(cells[..., None] - levels).argmin(-1)
    lines = ["".join(chars[i] for i in row).rstrip() for row in pick]
    while lines and not lines[0]:
        lines.pop(0)
    while lines and not lines[-1]:
        lines.pop()
    pad = min(len(l) - len(l.lstrip()) for l in lines if l)
    return "\n".join(l[pad:] for l in lines) + "\n"


def main():
    img = render()
    text = to_ascii(img)
    open(sys.argv[1], "w", encoding="utf-8", newline="\n").write(text)
    rows = text.rstrip("\n").split("\n")
    print(len(rows), "rows,", max(len(l) for l in rows), "cols")
    if len(sys.argv) > 2:
        g = (np.clip(img, 0, 1) * 255).astype(np.uint8)
        with open(sys.argv[2], "wb") as f:
            f.write(b"P5 %d %d 255\n" % (g.shape[1], g.shape[0]))
            f.write(g.tobytes())


if __name__ == "__main__":
    main()
