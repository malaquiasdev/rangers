import math
import random

from .canvas import Canvas, mix
from .hero_maps import HERO_MAPS
from .shade import banded, bayer, render_map
from .sprites import RANGER_ORDER, RANGER_RAMPS, ranger
from .title import extruded, sparkle, shadow, title_rows

WIDTH = 320
HEIGHT = 160
HORIZON = 127
RIM_WARM = (255, 170, 150)
AURA = {"zack": (150, 130, 230)}

SKY_STOPS = [
    (0, (9, 10, 38)),
    (22, (17, 16, 58)),
    (42, (30, 22, 88)),
    (62, (50, 30, 112)),
    (80, (80, 38, 128)),
    (96, (122, 46, 138)),
    (108, (172, 62, 140)),
    (118, (222, 98, 138)),
    (128, (250, 144, 128)),
]
MOON_RAMP = [(46, 36, 104), (84, 70, 150), (138, 122, 190), (196, 184, 226), (238, 228, 240), (255, 250, 236)]
LIGHT = (-0.5, -0.55, 0.67)
HERO_ORDER = ["billy", "kimberly", "jason", "zack", "trini", "tommy"]
HERO_X = {"billy": 64, "kimberly": 102, "jason": 141, "zack": 179, "trini": 218, "tommy": 256}
HERO_FEET = {"billy": 126, "kimberly": 128, "jason": 131, "zack": 131, "trini": 128, "tommy": 126}
TITLE_ORIGIN = (80, 12)


def sky_color(y, x):
    for index in range(len(SKY_STOPS) - 1):
        y0, c0 = SKY_STOPS[index]
        y1, c1 = SKY_STOPS[index + 1]
        if y0 <= y <= y1:
            t = (y - y0) / (y1 - y0)
            t = max(0.0, min(1.0, (t - 0.5) * 6.0 + 0.5))
            return c1 if t > bayer(x, y) + 0.5 else c0
    return SKY_STOPS[-1][1]


def lighten(color, amount=0.25, toward=(200, 180, 255)):
    return mix(color, toward, amount)


def paint_sky(canvas):
    for y in range(HEIGHT):
        for x in range(WIDTH):
            canvas.put(x, y, sky_color(y, x))


def stars(canvas, rng, avoid):
    for _ in range(120):
        x, y = rng.randrange(WIDTH), rng.randrange(0, 92)
        if avoid(x, y):
            continue
        brightness = rng.choice([(120, 130, 210), (160, 175, 240), (210, 220, 255), (255, 236, 180)])
        canvas.put(x, y, brightness)
    for _ in range(14):
        x, y = rng.randrange(6, WIDTH - 6), rng.randrange(4, 80)
        if avoid(x, y, 14):
            continue
        bright = rng.choice([(235, 240, 255), (255, 240, 200)])
        canvas.put(x, y, bright)
        dim = mix(bright, (60, 60, 140), 0.55)
        for dx, dy in ((1, 0), (-1, 0), (0, 1), (0, -1)):
            canvas.put(x + dx, y + dy, dim)
    for x, y, size in ((30, 14, 3), (292, 22, 3), (14, 74, 2), (300, 70, 2), (262, 12, 2), (70, 8, 2)):
        sparkle(canvas, x, y, size, (225, 235, 255))


def moon(canvas, cx, cy, radius):
    for extent, amount in ((radius + 14, 0.10), (radius + 8, 0.12), (radius + 3, 0.16)):
        for y in range(cy - extent, cy + extent + 1):
            for x in range(cx - extent, cx + extent + 1):
                distance = math.hypot(x + 0.5 - cx, y + 0.5 - cy)
                current = canvas.get(x, y)
                if current is None or distance > extent or distance <= radius:
                    continue
                edge = (extent - distance) / 3
                if edge > 1 or edge > bayer(x, y) + 0.5:
                    canvas.put(x, y, lighten(current, amount, (190, 170, 255)))
    for y in range(cy - radius - 1, cy + radius + 2):
        for x in range(cx - radius - 1, cx + radius + 2):
            distance = math.hypot(x + 0.5 - cx, y + 0.5 - cy)
            if distance > radius:
                continue
            nx, ny = (x + 0.5 - cx) / radius, (y + 0.5 - cy) / radius
            nz = math.sqrt(max(0.0, 1 - nx * nx - ny * ny))
            lambert = max(0.0, nx * LIGHT[0] + ny * LIGHT[1] + nz * LIGHT[2])
            value = lambert * 6.0 - 0.4
            for ccx, ccy, cr in CRATERS:
                relative = (x + 0.5 - (cx + ccx), y + 0.5 - (cy + ccy))
                inside = math.hypot(*relative) / cr
                if inside < 1:
                    value += -1.3 if relative[0] + relative[1] < cr * 0.4 else 0.6
                elif inside < 1.3 and relative[0] + relative[1] < 0:
                    value += 0.8
            index = max(0, min(5, banded(value + 0.5, x, y, 4.0)))
            canvas.put(x, y, MOON_RAMP[index])
    for y in range(cy - radius - 1, cy + radius + 2):
        for x in range(cx - radius - 1, cx + radius + 2):
            distance = math.hypot(x + 0.5 - cx, y + 0.5 - cy)
            if radius - 0.3 <= distance < radius + 0.7:
                canvas.put(x, y, (60, 44, 120) if x + y > cx + cy else (236, 226, 250))


CRATERS = [(-7, -6, 5), (6, -9, 3.5), (8, 5, 6), (-10, 8, 3.2), (-1, 11, 2.8), (2, -1, 2.4), (13, -2, 2.2)]


def ridge(rng, base, amplitude, step_range):
    points = []
    x = -10
    high = True
    while x < WIDTH + 30:
        height = amplitude * (rng.uniform(0.7, 1.0) if high else rng.uniform(0.05, 0.35))
        points.append((x, base - height))
        x += rng.randint(*step_range)
        high = not high
    return points


def ridge_y(points, x):
    for index in range(len(points) - 1):
        x0, y0 = points[index]
        x1, y1 = points[index + 1]
        if x0 <= x <= x1:
            return y0 + (y1 - y0) * (x - x0) / (x1 - x0)
    return points[-1][1]


def mountains(canvas, points, lit, shade, base, haze, haze_span):
    for x in range(WIDTH):
        top = ridge_y(points, x)
        slope = ridge_y(points, x + 3) - ridge_y(points, x - 3)
        face = lit if slope < 0 else shade
        for y in range(int(math.ceil(top)), base):
            depth = y - top
            color = face
            if depth < 1.5 and slope < 0:
                color = mix(face, (230, 200, 255), 0.45)
            fade = max(0.0, (y - (base - haze_span)) / haze_span)
            if fade > bayer(x, y) + 0.5:
                color = haze(y, x)
            canvas.put(x, y, color)


def beams(canvas):
    sources = [(56, 120, 34), (266, 120, -28), (206, 120, 18)]
    for bx, by, tilt in sources:
        for y in range(8, by):
            progress = (by - y) / (by - 8)
            axis = bx + tilt * progress
            half = 2.0 + 9.0 * progress
            for x in range(int(axis - half) - 1, int(axis + half) + 2):
                if not (0 <= x < WIDTH):
                    continue
                edge = 1 - abs(x - axis) / half
                if edge <= 0:
                    continue
                fade = 1 - progress * 0.55
                level = edge * fade
                current = canvas.get(x, y)
                if level > 0.5:
                    canvas.put(x, y, lighten(current, 0.20, (230, 214, 255)))
                elif level > 0.22 and (x + y) % 2 == 0:
                    canvas.put(x, y, lighten(current, 0.14, (230, 214, 255)))


def city(canvas, rng, layer, base_y, color, edge, height_range, width_range, lit_chance, start=-4, towers=(), cap=99):
    x = start
    while x < WIDTH + 4:
        width = rng.randint(*width_range)
        height = rng.randint(*height_range)
        if any(abs(x - tx) < 12 for tx, _ in towers):
            height = max(height, dict(towers).get(min(towers, key=lambda t: abs(x - t[0]))[0], height))
        elif 52 < x < 272:
            height = min(height, cap)
        top = base_y - height
        for yy in range(top, base_y):
            for xx in range(x, x + width):
                tone_color = color
                if yy == top or xx == x:
                    tone_color = edge
                elif xx == x + width - 1:
                    tone_color = mix(color, (10, 6, 30), 0.35)
                canvas.put(xx, yy, tone_color)
        if layer == "near":
            for yy in range(top + 3, base_y - 2, 4):
                for xx in range(x + 2, x + width - 2, 3):
                    if rng.random() < lit_chance:
                        glow = rng.choice([(255, 214, 110), (255, 214, 110), (255, 168, 100), (140, 232, 255)])
                        canvas.put(xx, yy, glow)
                        canvas.put(xx + 1, yy, glow)
                        canvas.put(xx, yy + 1, mix(glow, color, 0.45))
                        canvas.put(xx + 1, yy + 1, mix(glow, color, 0.45))
            if rng.random() < 0.3:
                mast = x + width // 2
                for yy in range(top - 6, top):
                    canvas.put(mast, yy, edge)
                canvas.put(mast, top - 7, (255, 80, 90))
        x += width + rng.randint(0, 2)


def stage(canvas):
    floor = [(18, 22, 58), (24, 32, 76), (30, 42, 96), (38, 54, 112)]
    for y in range(HORIZON, HEIGHT):
        t = (y - HORIZON) / (HEIGHT - HORIZON)
        for x in range(WIDTH):
            position = t * (len(floor) - 1)
            index = int(position)
            fraction = position - index
            nxt = min(len(floor) - 1, index + 1)
            canvas.put(x, y, floor[nxt] if (fraction - 0.5) * 2.4 + 0.5 > bayer(x, y) + 0.5 else floor[index])
    line = (66, 88, 160)
    vanish = (160, 40)
    for offset in range(-12, 13):
        bottom_x = 160 + offset * 34
        canvas.line(int(vanish[0] + (bottom_x - vanish[0]) * (HORIZON - vanish[1]) / (HEIGHT - 12 - vanish[1])), HORIZON + 1, bottom_x, HEIGHT - 12, line)
    for index in range(1, 6):
        y = HORIZON + int(24 * (index / 5) ** 1.8) + 2
        canvas.line(0, y, WIDTH - 1, y, (50, 68, 134))
    for x in range(WIDTH):
        canvas.put(x, HORIZON - 1, (130, 236, 255))
        canvas.put(x, HORIZON, (214, 250, 255))
        canvas.put(x, HORIZON + 1, (92, 70, 170) if x % 2 else (60, 50, 140))
    for y in range(HEIGHT - 11, HEIGHT):
        for x in range(WIDTH):
            band = y - (HEIGHT - 11)
            color = (88, 110, 190) if band == 0 else (46, 58, 128) if band < 3 else (22, 24, 70)
            if band > 3 and x % 16 in (7, 8):
                color = (255, 120, 190) if band == 6 else (160, 60, 130)
            canvas.put(x, y, color)


def backlight(canvas, name):
    center = HERO_X[name]
    feet = HERO_FEET[name]
    ramp = RANGER_RAMPS[name]
    if name in AURA:
        ramp = [AURA[name]] * 6
    for y in range(58, feet):
        height = (feet - y) / (feet - 58)
        for x in range(center - 18, center + 19):
            level = (1 - abs(x - center) / 19) * (1.2 - height * 1.25)
            current = canvas.get(x, y)
            if current is None:
                continue
            if level > 0.5:
                canvas.put(x, y, mix(current, ramp[4], 0.5))
            elif level > 0.28:
                canvas.put(x, y, mix(current, ramp[3], 0.4))
            elif level > 0.1 and (x + y) % 2 == 0:
                canvas.put(x, y, mix(current, ramp[3], 0.34))


def stage_ring(canvas):
    cx, cy, rx, ry = 160, 142, 128, 11
    for y in range(cy - ry - 3, cy + ry + 4):
        for x in range(cx - rx - 4, cx + rx + 5):
            distance = math.hypot((x + 0.5 - cx) / rx, (y + 0.5 - cy) / ry)
            current = canvas.get(x, y)
            if current is None or y < HORIZON + 1:
                continue
            band = abs(distance - 1) * ry
            if band < 0.8:
                canvas.put(x, y, (150, 240, 255))
            elif band < 1.5:
                canvas.put(x, y, mix(current, (96, 200, 255), 0.42))
            elif band < 2.6 and (x + y) % 2 == 0:
                canvas.put(x, y, mix(current, (96, 200, 255), 0.26))


def place_hero(canvas, name):
    rows = HERO_MAPS[name]
    sprite = render_map(rows, ranger(RANGER_RAMPS[name]), RANGER_ORDER, rim=RIM_WARM)
    x = HERO_X[name] - sprite.width // 2
    feet = HERO_FEET[name]
    y = feet - sprite.height + 1
    glow = RANGER_RAMPS[name][4]
    for dy in range(-4, 12):
        for dx in range(-26, 27):
            distance = math.hypot(dx / 26, dy / 9)
            if distance < 1:
                tx, ty = HERO_X[name] + dx, feet + 2 + dy
                current = canvas.get(tx, ty)
                if current is not None and ty >= HORIZON and (1 - distance) * 0.95 > bayer(tx, ty) + 0.5:
                    canvas.put(tx, ty, mix(current, glow, 0.5))
    for row in range(18):
        source = sprite.height - 1 - row
        if row > 11 and row % 3:
            continue
        if row > 5 and row % 2:
            continue
        for column in range(sprite.width):
            color = sprite.get(column, source)
            if color is None:
                continue
            tx, ty = x + column, feet + 2 + row
            base = canvas.get(tx, ty)
            if base is not None:
                canvas.put(tx, ty, mix(base, color, 0.5 - row * 0.015))
    canvas.blit(sprite, x, y)


def vignette(canvas):
    for y in range(HEIGHT):
        for x in range(WIDTH):
            dx = (x - WIDTH / 2) / (WIDTH / 2)
            dy = (y - HEIGHT / 2) / (HEIGHT / 2)
            strength = max(0.0, math.hypot(dx, dy * 1.1) - 0.7) * 3.2
            level = min(3, banded(strength, x, y, 3.0))
            if level > 0:
                canvas.put(x, y, mix(canvas.get(x, y), (6, 4, 24), 0.16 * level))


def build_banner():
    rng = random.Random(7)
    canvas = Canvas(WIDTH, HEIGHT)
    paint_sky(canvas)
    title = extruded(title_rows())
    tx, ty = TITLE_ORIGIN

    def avoid(x, y, margin=3):
        near_moon = math.hypot(x - 46, y - 46) < 44
        in_title = tx - margin <= x <= tx + title.width + 8 + margin and ty - margin <= y <= ty + title.height + 9 + margin
        return near_moon or in_title

    stars(canvas, rng, avoid)
    moon(canvas, 46, 46, 22)
    far = ridge(rng, 98, 26, (16, 26))
    mountains(canvas, far, (126, 80, 170), (78, 52, 138), 112, lambda y, x: sky_color(y - 14, x), 14)
    near = ridge(rng, 108, 20, (14, 22))
    mountains(canvas, near, (92, 56, 140), (50, 32, 100), 120, lambda y, x: (96, 46, 128), 10)
    beams(canvas)
    city(canvas, rng, "far", HORIZON, (58, 40, 112), (112, 86, 176), (14, 34), (8, 16), 0.0, start=-6, cap=22)
    towers = ((8, 62), (30, 52), (292, 58), (312, 66))
    city(canvas, rng, "near", HORIZON, (26, 18, 64), (84, 62, 150), (8, 22), (10, 18), 0.3, towers=towers, cap=12)
    stage(canvas)
    stage_ring(canvas)
    vignette(canvas)
    for name in HERO_ORDER:
        backlight(canvas, name)
    for name in sorted(HERO_ORDER, key=lambda n: HERO_FEET[n]):
        place_hero(canvas, name)
    shadow(title, canvas, tx, ty)
    canvas.blit(title, tx, ty)
    for x, y, size in ((tx + 3, ty + 2, 4), (tx + 70, ty - 1, 3), (tx + title.width - 2, ty + 12, 4), (tx + 112, ty + title.height - 2, 2), (tx + 34, ty + 20, 2)):
        sparkle(canvas, x, y, size)
    return canvas
