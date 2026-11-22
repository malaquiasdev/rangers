import math

from .canvas import Canvas, mix

BAYER = [[0, 8, 2, 10], [12, 4, 14, 6], [3, 11, 1, 9], [15, 7, 13, 5]]
LIT_STEPS = {1: 2.0, 2: 1.0}
SHADE_STEPS = {1: 2.0, 2: 1.0}
HALF_STEP = 0.5
EMPTY = "."


def bayer(x, y):
    return (BAYER[y % 4][x % 4] + 0.5) / 16 - 0.5


def banded(value, x, y, sharpness=5.0):
    base = math.floor(value)
    amount = max(0.0, min(1.0, (value - base - 0.5) * sharpness + 0.5))
    return base + (1 if amount > bayer(x, y) + 0.5 else 0)


class Material:
    def __init__(self, ramp, flat=None, lit=1.0, shade=1.0, dither=0.0, outline=True, glow=False):
        self.ramp = ramp
        self.flat = flat
        self.lit = lit
        self.shade = shade
        self.dither = dither
        self.outline = outline
        self.glow = glow


def distance(rows, x, y, step_x, step_y, limit=4):
    char = rows[y][x]
    count = 0
    while count < limit:
        x += step_x
        y += step_y
        count += 1
        if not (0 <= y < len(rows) and 0 <= x < len(rows[y])) or rows[y][x] != char:
            return count
    return limit + 1


def tone(rows, materials, x, y):
    material = materials[rows[y][x]]
    if material.flat is not None:
        return float(material.flat)
    toward_light = distance(rows, x, y, -1, -1)
    away_light = distance(rows, x, y, 1, 1)
    lit = LIT_STEPS.get(toward_light, HALF_STEP if toward_light == 3 and material.dither else 0.0) * material.lit
    shade = SHADE_STEPS.get(away_light, HALF_STEP if away_light == 3 and material.dither else 0.0) * material.shade
    return 3 + lit - shade


def quantize(value, x, y, strength):
    return max(0, min(5, int(value + 0.5 + bayer(x, y) * strength * 0.98)))


def render_map(rows, materials, order, rim=None, pad=1):
    height = len(rows)
    width = len(rows[0])
    for index, row in enumerate(rows):
        if len(row) != width:
            raise ValueError(f"row {index} has {len(row)} columns, expected {width}")
    for char in set("".join(rows)) - {EMPTY}:
        if char not in materials:
            raise ValueError(f"no material for {char!r}")
    depth = {char: index for index, char in enumerate(order)}
    canvas = Canvas(width + pad * 2, height + pad * 2)

    def char_at(x, y):
        if 0 <= y < height and 0 <= x < width:
            return rows[y][x]
        return EMPTY

    for y in range(height):
        for x in range(width):
            char = rows[y][x]
            if char == EMPTY:
                continue
            material = materials[char]
            index = quantize(tone(rows, materials, x, y), x, y, material.dither)
            color = material.ramp[index]
            border = [char_at(x + dx, y + dy) for dx, dy in ((1, 0), (-1, 0), (0, 1), (0, -1))]
            covered = any(other != EMPTY and other != char and depth[other] > depth[char] for other in border)
            if covered and material.outline:
                color = material.ramp[0]
            elif rim is not None and index <= 3 and (char_at(x + 1, y) == EMPTY or char_at(x, y + 1) == EMPTY):
                color = mix(material.ramp[3], rim, 0.4)
            canvas.put(x + pad, y + pad, color)

    for y in range(-1, height + 1):
        for x in range(-1, width + 1):
            if char_at(x, y) != EMPTY:
                continue
            neighbors = [char_at(x + dx, y + dy) for dx, dy in ((1, 0), (-1, 0), (0, 1), (0, -1))]
            filled = [other for other in neighbors if other != EMPTY]
            if filled:
                top = max(filled, key=lambda other: depth[other])
                canvas.put(x + pad, y + pad, materials[top].ramp[0])
    return canvas
