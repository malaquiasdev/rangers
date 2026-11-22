import math

from . import palette as P
from .canvas import mix
from .geometry import ellipse, grow, polygon
from .noise import fbm
from .painter import bayer


def pick(ramp, value, x, y, sharp=4.0):
    base = math.floor(value)
    amount = max(0.0, min(1.0, (value - base - 0.5) * sharp + 0.5))
    index = base + (1 if amount > bayer(x, y) + 0.5 else 0)
    return ramp[max(0, min(len(ramp) - 1, index))]


class Layout:
    def __init__(self, width, height, floor, mist_height, rings, tubes, slabs, panel_top, panel_bottom):
        self.width = width
        self.height = height
        self.floor = floor
        self.mist_height = mist_height
        self.rings = rings
        self.tubes = tubes
        self.slabs = slabs
        self.panel_top = panel_top
        self.panel_bottom = panel_bottom


def paint_base(canvas, layout):
    width, height = layout.width, layout.height
    cx = width / 2
    for y in range(height):
        for x in range(width):
            fall = (y / height) ** 0.9
            edge = abs(x - cx) / cx
            glow = math.exp(-(((x - cx) / (width * 0.36)) ** 2 + ((y - height * 0.3) / (height * 0.32)) ** 2))
            value = 0.7 + fall * 1.5 + glow * 1.5 - edge ** 2 * 0.9
            canvas.put(x, y, pick(P.NIGHT, value, x, y, 3.2))


def paint_dots(canvas, box, seed=1):
    x0, y0, x1, y1 = box
    for y in range(y0, y1):
        for x in range(x0, x1):
            if (x + (y // 2) * 2) % 4 == 0 and y % 2 == 0:
                fade = min((x - x0) / 10, (x1 - x) / 10, (y - y0) / 14, (y1 - y) / 24, 1.0)
                if fade > bayer(x, y) + 0.5:
                    canvas.put(x, y, mix((24, 10, 30), (150, 36, 74), 0.35 + 0.4 * ((x * 7 + y * 3) % 5) / 5))


def paint_mist(canvas, layout, seed=7):
    width = layout.width
    cx = width / 2
    height = layout.mist_height
    for y in range(height):
        spread = width * (0.36 - 0.2 * (y / height) ** 0.8)
        for x in range(width):
            dx = (x - cx) / spread
            if abs(dx) >= 1.25:
                continue
            envelope = max(0.0, 1.0 - dx * dx * 0.78)
            swirl = fbm(seed, x * 0.028 + y * 0.012, y * 0.034, 4)
            wisp = fbm(seed + 5, x * 0.06 - y * 0.02, y * 0.05, 3)
            top = 1.0 - (y / height) ** 1.4
            field = envelope * (0.3 + 0.85 * swirl) * (0.55 + 0.6 * top) + (wisp - 0.5) * 0.18
            level = (field - 0.22) * 8.0
            if level <= 0.35:
                continue
            under = canvas.get(x, y) or P.NIGHT[1]
            color = pick(P.MIST, min(5.0, level), x, y, 3.0)
            canvas.put(x, y, mix(under, color, 0.92))


def halo(canvas, pixels, color, reach=4, strength=0.55):
    ring = set(pixels)
    for step in range(reach):
        ring = grow(ring, 1, diagonal=False)
        alpha = strength * (1 - step / reach) ** 1.6
        for x, y in ring:
            if (x, y) in pixels:
                continue
            under = canvas.get(x, y)
            if under is None or alpha <= bayer(x, y) + 0.5 - 0.5 * (1 - alpha):
                continue
            canvas.put(x, y, mix(under, color, min(1.0, alpha * 1.4)))


def paint_tube(canvas, x, y0, y1, half=2, core=(210, 250, 255), body=(40, 150, 230), edge=(12, 40, 110)):
    pixels = set()
    for y in range(y0, y1):
        for dx in range(-half, half + 1):
            pixels.add((x + dx, y))
    halo(canvas, pixels, (30, 120, 230), 6, 0.6)
    for xx, yy in pixels:
        dx = xx - x
        if abs(dx) == half:
            canvas.put(xx, yy, edge)
        elif dx == 0 or (dx == -1 and half >= 2):
            canvas.put(xx, yy, core if dx == -1 or half < 2 else body)
        else:
            canvas.put(xx, yy, body)
    if half >= 2:
        for yy in range(y0, y1):
            canvas.put(x - 1, yy, core)
            canvas.put(x, yy, mix(core, body, 0.45))


def paint_ring(canvas, cx, cy, rx, ry, tube_x=None):
    outer = ellipse(cx, cy, rx, ry)
    inner = ellipse(cx, cy, rx - 2, ry - 1.5)
    band = outer - inner
    halo(canvas, band, (40, 140, 255), 5, 0.55)
    for x, y in band:
        front = y + 0.5 >= cy
        distance = abs(x + 0.5 - cx) / rx
        if front:
            color = (190, 238, 255) if distance < 0.55 else (84, 178, 250)
            if y + 0.5 > cy + ry - 1.6:
                color = (120, 200, 255)
        else:
            color = (26, 70, 170) if distance > 0.3 else (46, 110, 210)
        canvas.put(x, y, color)


def paint_slab(canvas, points, side):
    pixels = polygon(points)
    top = min(y for _, y in points)
    bottom = max(y for _, y in points)
    left = min(x for x, _ in points)
    right = max(x for x, _ in points)
    for x, y in pixels:
        across = (x - left) / max(1, right - left)
        if side == "right":
            across = 1 - across
        along = (y - top) / max(1, bottom - top)
        value = 1.4 + across * 2.6 + (1 - along) * 0.6
        if (x - left) % 11 < 2:
            value -= 0.9
        canvas.put(x, y, pick(P.MINT, value, x, y, 3.0))
    edge = grow(pixels, 1, diagonal=False) - pixels
    for x, y in edge:
        canvas.put(x, y, P.MINT[0])


def paint_floor(canvas, layout):
    width, height = layout.width, layout.height
    top = layout.floor
    cx = width / 2
    vanishing = (cx, top - (height - top) * 1.4)
    for y in range(top, height):
        depth = (y - top) / (height - top)
        for x in range(width):
            glow = math.exp(-(((x - cx) / (width * 0.42)) ** 2) - ((depth - 0.65) / 0.55) ** 2)
            value = 0.6 + depth * 0.9 + glow * 2.4
            canvas.put(x, y, pick(P.MIST, value * 0.9, x, y, 3.0) if value > 2.2 else pick(P.NIGHT, value * 1.4, x, y, 3.0))
    for line in range(1, 9):
        progress = (line / 9) ** 1.7
        y = int(top + progress * (height - top))
        for x in range(width):
            under = canvas.get(x, y)
            canvas.put(x, y, mix(under, (90, 200, 190), 0.42))
    for column in range(-14, 15):
        x_far = cx + column * 22
        for y in range(top, height):
            depth = (y - top) / (height - top)
            x = vanishing[0] + (x_far - vanishing[0]) * ((y - vanishing[1]) / (top - vanishing[1])) if top != vanishing[1] else x_far
            xi = int(round(x))
            if 0 <= xi < width:
                under = canvas.get(xi, y)
                canvas.put(xi, y, mix(under, (90, 200, 190), 0.3 * (0.4 + depth)))


def paint_platform(canvas, cx, cy, rx, ry):
    for scale, color, alpha in ((1.0, (90, 230, 220), 0.9), (0.86, (60, 170, 200), 0.7), (0.5, (60, 150, 190), 0.4)):
        outer = ellipse(cx, cy, rx * scale, ry * scale)
        inner = ellipse(cx, cy, rx * scale - 1.6, ry * scale - 0.9)
        band = outer - inner
        for x, y in band:
            under = canvas.get(x, y)
            if under is not None:
                canvas.put(x, y, mix(under, color, alpha))
        if scale == 1.0:
            halo(canvas, band, (60, 220, 210), 4, 0.5)


def sparkle(canvas, x, y, size, color=(235, 252, 255)):
    canvas.put(x, y, color)
    dim = mix(color, (60, 140, 200), 0.5)
    for step in range(1, size + 1):
        tone = color if step < size - 0 and step <= size // 2 + 1 else dim
        for dx, dy in ((step, 0), (-step, 0), (0, step), (0, -step)):
            canvas.put(x + dx, y + dy, tone)


def paint_stars(canvas, layout, seed=3):
    width = layout.width
    for index in range(70):
        x = int(fbm(seed, index * 3.1, 1.7) * (width - 2)) + 1
        y = int(fbm(seed + 1, index * 2.3, 9.2) * layout.panel_top * 0.8)
        under = canvas.get(x, y)
        if under is not None and sum(under) < 140:
            canvas.put(x, y, mix(under, (170, 200, 255), 0.7))


def paint_beams(canvas, cx, cy, reach, count=11, tint=(150, 235, 215)):
    for index in range(count):
        angle = math.radians(20 + index * (140 / (count - 1)))
        spread = math.radians(1.7 + 0.8 * ((index * 5) % 3))
        left = (cx + math.cos(angle - spread) * reach, cy + math.sin(angle - spread) * reach)
        right = (cx + math.cos(angle + spread) * reach, cy + math.sin(angle + spread) * reach)
        for x, y in polygon([(cx, cy), left, right]):
            under = canvas.get(x, y)
            if under is None:
                continue
            fade = 1 - math.hypot(x - cx, y - cy) / reach
            if fade * 0.9 > bayer(x, y) + 0.5 and (x + y) % 2 == 0:
                canvas.put(x, y, mix(under, tint, 0.28))


def paint_backdrop(canvas, layout):
    paint_base(canvas, layout)
    paint_stars(canvas, layout)
    paint_dots(canvas, (0, int(layout.height * 0.06), int(layout.width * 0.2), int(layout.panel_top)))
    paint_dots(canvas, (int(layout.width * 0.8), int(layout.height * 0.06), layout.width, int(layout.panel_top)))
    paint_mist(canvas, layout)
    paint_beams(canvas, layout.width // 2, int(layout.height * 0.13), layout.height * 0.55)
    for tube in layout.tubes:
        paint_tube(canvas, *tube)
    for ring in layout.rings:
        paint_ring(canvas, *ring)
    for points, side in layout.slabs:
        paint_slab(canvas, points, side)
    paint_floor(canvas, layout)
