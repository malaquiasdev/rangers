from .canvas import Canvas, mix
from .shade import Material, banded, tone
from .title_map import TITLE_MAP

TITLE_RAMP = [
    (255, 250, 214),
    (255, 232, 140),
    (255, 203, 64),
    (255, 160, 36),
    (244, 112, 28),
    (212, 68, 34),
    (166, 42, 48),
    (112, 31, 62),
    (61, 20, 66),
]
SIDE_RAMP = [(196, 60, 40), (160, 42, 52), (130, 34, 62), (98, 28, 70), (70, 22, 72)]
OUTLINE = (24, 10, 44)
DEPTH = 5
SHADOW_OFFSET = (7, 8)


def title_rows():
    return TITLE_MAP["rangers"]


def face(rows):
    height = len(rows)
    width = len(rows[0])
    materials = {"T": Material([(0, 0, 0)] * 6)}
    canvas = Canvas(width, height)
    for y in range(height):
        for x in range(width):
            if rows[y][x] == ".":
                continue
            bevel = tone(rows, materials, x, y) - 3
            band = banded(1.0 + (y / height) * 3.9 + 0.5, x, y, 4.0)
            value = band - bevel
            value -= 1 if (x - 0.9 * y) % 47 < 2.6 or (x - 0.9 * y) % 47 > 45.4 else 0
            index = max(0, min(len(TITLE_RAMP) - 1, int(value)))
            canvas.put(x, y, TITLE_RAMP[index])
    return canvas


def extruded(rows):
    base = face(rows)
    width = base.width + DEPTH + 2
    height = base.height + DEPTH + 2
    result = Canvas(width, height)
    for step in range(DEPTH, 0, -1):
        color = SIDE_RAMP[step - 1]
        for y in range(base.height):
            for x in range(base.width):
                if base.get(x, y) is not None:
                    result.put(x + 1 + step, y + 1 + step, color)
    result.blit(base, 1, 1)
    outlined = Canvas(width + 2, height + 2)
    outlined.blit(result, 1, 1)
    for y in range(-1, height + 1):
        for x in range(-1, width + 1):
            if result.get(x, y) is not None:
                continue
            if any(result.get(x + dx, y + dy) is not None for dx, dy in ((1, 0), (-1, 0), (0, 1), (0, -1))):
                outlined.put(x + 1, y + 1, OUTLINE)
    return outlined


def shadow(canvas, target, x, y, strength=0.62):
    dx, dy = SHADOW_OFFSET
    for row in range(canvas.height):
        for column in range(canvas.width):
            if canvas.get(column, row) is None:
                continue
            tx, ty = x + column + dx, y + row + dy
            current = target.get(tx, ty)
            if current is not None:
                target.put(tx, ty, mix(current, (8, 4, 28), strength))


def sparkle(canvas, x, y, size=3, color=(255, 255, 255)):
    tint = [color, mix(color, (255, 230, 150), 0.5), mix(color, (255, 160, 120), 0.6)]
    canvas.put(x, y, color)
    for step in range(1, size + 1):
        shade = tint[min(step - 1, 2)] if step > 1 else color
        for dx, dy in ((step, 0), (-step, 0), (0, step), (0, -step)):
            canvas.put(x + dx, y + dy, shade)
    canvas.put(x + 1, y + 1, mix(color, (255, 200, 140), 0.5))
    canvas.put(x - 1, y - 1, mix(color, (255, 200, 140), 0.5))
