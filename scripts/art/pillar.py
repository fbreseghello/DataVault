"""Renders the rest screen art: a broken Doric column and a jawless skull,
shaded like a photo turned into text.

    python scripts/art/pillar.py web/src/lib/pillar.txt

The scene is continuous (in character cells; a cell is about twice as tall as
wide). Each cell averages several samples, so silhouettes are antialiased, and
brightness 0..1 maps onto a long character ramp from sparse to dense.
"""
import math
import sys

W, H = 172, 84
RAMP = sys.argv[2] if len(sys.argv) > 2 else " .:-=+*#%@"
SUB_X, SUB_Y = 3, 5

cx = 52.0
ground = 78.0
step_h = 3.4
steps = [44.0, 38.0, 32.0]
plinth = ground - step_h * len(steps)
top_rows = 10.0
R_BOTTOM, R_TOP = 21.0, 18.6
FLUTES = 20
LIGHT = -0.5


def lerp(a, b, t):
    return a + (b - a) * t


def smooth(t):
    return t * t * (3 - 2 * t)


def hash2(ix, iy):
    n = (ix * 374761393 + iy * 668265263) & 0xFFFFFFFF
    n = (n ^ (n >> 13)) * 1274126177 & 0xFFFFFFFF
    return ((n ^ (n >> 16)) & 0xFFFF) / 0xFFFF


def vnoise(x, y):
    ix, iy = math.floor(x), math.floor(y)
    fx, fy = smooth(x - ix), smooth(y - iy)
    a, b = hash2(ix, iy), hash2(ix + 1, iy)
    c, d = hash2(ix, iy + 1), hash2(ix + 1, iy + 1)
    return lerp(lerp(a, b, fx), lerp(c, d, fx), fy)


def stone(x, y):
    """Weathered stone: a few octaves of smooth noise, centred on 0."""
    v = 0.0
    amp, f = 0.5, 0.18
    for _ in range(4):
        v += amp * (vnoise(x * f, y * f * 2.0) - 0.5)
        amp *= 0.5
        f *= 2.1
    return v


def interp(points, x):
    for (x0, y0), (x1, y1) in zip(points, points[1:]):
        if x0 <= x <= x1:
            return lerp(y0, y1, (x - x0) / (x1 - x0))
    return points[-1][1]


BREAK = [(-1.0, 17.0), (-0.86, 13.0), (-0.7, 11.0), (-0.55, 12.5), (-0.4, 10.0),
         (-0.22, 12.0), (-0.05, 15.5), (0.1, 14.5), (0.28, 18.0), (0.45, 16.5),
         (0.62, 20.0), (0.8, 19.0), (1.0, 23.0)]


def break_row(u):
    return interp(BREAK, u) + 1.2 * (vnoise(u * 9, 3.3) - 0.5) + 0.5 * (vnoise(u * 31, 7.1) - 0.5)


CRACK = [(cx + 6.0, 19.0), (cx + 3.0, 28.0), (cx + 4.5, 35.0), (cx - 1.0, 44.0),
         (cx + 1.0, 52.0), (cx - 3.0, 60.0)]


def crack_dist(x, y):
    for (x0, y0), (x1, y1) in zip(CRACK, CRACK[1:]):
        if y0 <= y <= y1:
            t = (y - y0) / (y1 - y0)
            return x - lerp(x0, x1, t)
    return 99.0


def radius(y):
    t = max(0.0, min(1.0, (plinth - y) / (plinth - top_rows)))
    return lerp(R_BOTTOM, R_TOP, t)


def shaft(x, y):
    if y > plinth or y < top_rows - 2:
        return None
    r = radius(y)
    u = (x - cx) / r
    if abs(u) >= 1:
        return None
    top = break_row(u)
    if y < top:
        return None
    th = math.asin(u)
    # Flutes are concave channels: inside each one the normal swings across.
    k = th * FLUTES / (2 * math.pi)
    f = k - math.floor(k) - 0.5          # -0.5..0.5 across a channel
    swing = -1.1 * f                       # the normal turns against the slope
    n = th + swing
    diffuse = max(0.0, math.cos(n - LIGHT))
    spec = max(0.0, math.cos(n - LIGHT * 0.6)) ** 24
    occl = 1.0 - 0.62 * (1 - abs(f) * 2) ** 2   # channel floors are darker
    arris = 1.0 + 0.35 * max(0.0, abs(f) * 2 - 0.85) / 0.15   # sharp edges glint
    b = (0.24 + 0.70 * diffuse) * occl * arris + 0.20 * spec
    b += 0.16 * stone(x, y)
    # The fracture surface faces up: lit, rough, a lip of a couple of rows.
    depth = y - top
    if depth < 2.2:
        b = 0.55 + 0.35 * (1 - (u + 1) / 2) + 0.25 * stone(x * 1.7, y * 1.7)
        b *= 0.75 + 0.25 * min(1.0, depth / 0.8)
    # Darker toward the base (occlusion), lighter toward the top.
    b *= 0.78 + 0.22 * min(1.0, (plinth - y) / 10.0)
    d = crack_dist(x, y)
    if abs(d) < 0.45:
        b *= 0.08
    elif 0 < d < 1.4:
        b *= 1.18           # the lip facing the light
    elif -1.6 < d < 0:
        b *= 0.6            # the lip in shadow
    return b


def base(x, y):
    for k, hw in enumerate(steps):
        bottom = ground - step_h * k
        top = bottom - step_h
        if top <= y < bottom and abs(x - cx) <= hw:
            u = (x - cx) / hw
            on_top = y - top < 0.9
            if on_top:
                b = 0.80 - 0.30 * (u + 1) / 2
                # the shaft and the step above throw a shadow to the right
                if cx + R_BOTTOM - 2 < x < cx + R_BOTTOM + 16 - 3 * k:
                    b *= 0.42
            else:
                b = 0.46 - 0.26 * (u + 1) / 2
                b *= 0.85 + 0.15 * (bottom - y) / step_h
                # joints between the stone blocks
                if ((x - cx + hw + 3 * k) % 12.0) < 0.45:
                    b *= 0.25
            b += 0.12 * stone(x * 1.3, y * 1.3)
            if abs(u) > 0.975:
                b *= 0.55
            return b
    return None


# A fallen drum standing on its rim, its round face toward us, to the right.
DCX, DR = cx + 72.0, 13.0
DCY = ground - DR / 2.0
DEPTH_X, DEPTH_Y = 5.0, -1.6   # the back face is offset up and right


def drum(x, y):
    # front face: an ellipse (circle squashed by the cell aspect)
    ex, ey = (x - DCX) / DR, (y - DCY) / (DR / 2.0)
    rr = ex * ex + ey * ey
    if rr <= 1.0:
        ang = math.atan2(ey, ex)
        rim = math.sqrt(rr)
        # fluted rim: the outline is notched by the channels
        notch = 0.04 * (0.5 + 0.5 * math.cos(ang * FLUTES))
        if rim > 1.0 - notch:
            return None
        # broken face: rough, lit from the upper left
        b = 0.42 + 0.30 * (-ex * 0.6 - ey * 0.5) + 0.22 * stone(x * 1.6, y * 1.6)
        if rim > 0.86:
            b *= 0.8
        return b
    # the curved side seen between the front and back faces
    for t in (0.25, 0.5, 0.75, 1.0):
        bx, by = DCX + DEPTH_X * t, DCY + DEPTH_Y * t
        ex2, ey2 = (x - bx) / DR, (y - by) / (DR / 2.0)
        if ex2 * ex2 + ey2 * ey2 <= 1.0:
            ang = math.atan2(ey2, ex2)
            lit = max(0.0, math.cos(ang - math.radians(-120)))
            channel = 0.5 + 0.5 * math.cos(ang * FLUTES)
            return (0.08 + 0.55 * lit) * (0.7 + 0.3 * channel) + 0.08 * stone(x, y)
    return None


ROCKS = [(cx - 47, 1.2, 2.4), (cx + 47, 1.6, 3.0), (cx + 53, 1.0, 1.8),
         (cx + 95, 1.3, 2.4), (cx - 52, 0.8, 1.4), (cx + 41, 0.7, 1.2)]


def rocks(x, y):
    for rx, ry, rw in ROCKS:
        dx, dy = (x - rx) / rw, (y - (ground - ry)) / ry
        if dx * dx + dy * dy <= 1:
            return 0.22 + 0.45 * max(0.0, -dx * 0.6 - dy * 0.8) + 0.1 * stone(x * 2, y * 2)
    return None


def ground_band(x, y):
    if ground <= y < ground + 2.5 and cx - 58 < x < cx + 98:
        fade = 1 - (y - ground) / 2.5
        b = (0.13 + 0.10 * stone(x * 0.7, y * 3)) * fade
        if cx + steps[0] < x < cx + steps[0] + 14:
            b *= 0.3
        if SKX + 1.0 * SKU < x < SKX + 1.7 * SKU:
            b *= 0.3
        return b if b > 0.03 else None
    return None


# A skull without its jaw, resting on the ground right of the column and turned
# a little to its left (our left): the features slide left of the cranium, the
# far (left) orbit is foreshortened and the right temple turns into shadow.
# u, v are normalised: the cranium is about 2 wide; v grows downward.
SKU = 22.0                                  # columns per unit
SKX = cx + 84.0
SKY = ground - 1.12 * SKU / 2.0             # teeth rest on the ground
TURN = -0.16


def skull(x, y):
    u = (x - SKX) / SKU
    v = (y - SKY) * 2.0 / SKU
    t = TURN

    cranium = ((u - 0.07) / 1.0) ** 2 + ((v + 0.33) / 0.97) ** 2 <= 1.0
    face = False
    if 0.0 <= v <= 0.92:
        if v < 0.48:
            hw = 0.86 + 0.08 * math.sin(v / 0.48 * math.pi / 2)   # out to the cheekbones
        else:
            hw = 0.94 - (v - 0.48) / 0.44 * 0.40                  # in to the teeth
        face = abs(u - t * 0.55) <= hw
    tu = u - t
    teeth = 0.86 <= v <= 1.11 + 0.03 * math.sin(tu * 40) and abs(tu) <= 0.44
    if not (cranium or face or teeth):
        return None

    # One rounded volume for the light: bone faces us, lit from the upper left.
    nx, ny = (u - 0.04) / 1.05, (v - 0.12) / 1.30
    nz = math.sqrt(max(0.0, 1 - min(1.0, nx * nx + ny * ny)))
    lx, ly, lz = -0.50, -0.42, 0.76
    diffuse = max(0.0, nx * lx + ny * ly + nz * lz)
    b = 0.22 + 0.74 * diffuse + 0.09 * stone(x * 1.5, y * 1.5)

    if teeth and not face:
        if (tu + 0.02) % 0.105 < 0.03:
            return 0.03                                         # gaps between teeth
        return 0.70 + 0.22 * (1 - (v - 0.86) / 0.25) + 0.05 * stone(x * 3, y * 3)

    # Orbits: deep holes, the far one narrower; their rims fall into shadow.
    for ex, rx in ((t - 0.40, 0.235), (t + 0.42, 0.30)):
        o = ((u - ex) / rx) ** 2 + ((v - 0.27) / 0.255) ** 2
        if o <= 1.0:
            return 0.02 + 0.08 * o ** 4
        if o <= 1.45:
            b *= 0.45 + 0.55 * (o - 1.0) / 0.45
        brow = ((u - ex) / (rx + 0.06)) ** 2 + ((v + 0.04) / 0.07) ** 2
        if brow <= 1.0:
            b = min(1.0, b + 0.1)

    # Nasal cavity: an upside-down pear below and between the orbits.
    if 0.50 <= v <= 0.80:
        k = (v - 0.50) / 0.30
        w = 0.03 + 0.10 * math.sin(k * math.pi * 0.78)
        if abs(u - t + 0.01) <= w:
            return 0.03
    # Hollows: under the cheekbones, and the temple turned away on the right.
    for hx, hy, hrx, hry, f in ((t - 0.66, 0.66, 0.16, 0.16, 0.7), (t + 0.70, 0.66, 0.2, 0.18, 0.6),
                                (0.86, 0.08, 0.16, 0.30, 0.55)):
        if ((u - hx) / hrx) ** 2 + ((v - hy) / hry) ** 2 <= 1.0:
            b *= f
    # A suture line across the dome.
    if cranium and v < -0.35 and abs(v + 0.75 - 0.18 * (u - 0.2) - 0.03 * math.sin(u * 30)) < 0.022:
        b *= 0.3
    return b


LAYERS = [skull, shaft, rocks, base, ground_band]


def sample(x, y):
    for f in LAYERS:
        b = f(x, y)
        if b is not None:
            return max(0.0, min(1.0, b))
    return None


lines = []
for row in range(H):
    out = []
    for col in range(W):
        total, hits = 0.0, 0
        for sy in range(SUB_Y):
            for sx in range(SUB_X):
                b = sample(col + (sx + 0.5) / SUB_X, row + (sy + 0.5) / SUB_Y)
                if b is not None:
                    total += b
                    hits += 1
        n = SUB_X * SUB_Y
        if hits / n < 0.25:
            out.append(" ")
            continue
        b = total / n
        i = max(1, min(len(RAMP) - 1, int(round(b * (len(RAMP) - 1)))))
        out.append(RAMP[i])
    lines.append("".join(out).rstrip())

while lines and not lines[0]:
    lines.pop(0)
while lines and not lines[-1]:
    lines.pop()
pad = min(len(l) - len(l.lstrip()) for l in lines if l)
text = "\n".join(l[pad:] for l in lines) + "\n"
open(sys.argv[1], "w", encoding="utf-8", newline="\n").write(text)
print(len(lines), "rows,", max(len(l) for l in lines) - pad, "cols")
