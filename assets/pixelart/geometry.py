import math


def polygon(points):
    if len(points) < 3:
        return set()
    pixels = set()
    low = int(math.floor(min(y for _, y in points)))
    high = int(math.ceil(max(y for _, y in points)))
    count = len(points)
    for y in range(low, high + 1):
        center = y + 0.5
        crossings = []
        for index in range(count):
            x0, y0 = points[index]
            x1, y1 = points[(index + 1) % count]
            if (y0 <= center < y1) or (y1 <= center < y0):
                crossings.append(x0 + (center - y0) / (y1 - y0) * (x1 - x0))
        crossings.sort()
        for pair in range(0, len(crossings) - 1, 2):
            start = int(math.ceil(crossings[pair] - 0.5))
            stop = int(math.floor(crossings[pair + 1] - 0.5))
            for x in range(start, stop + 1):
                pixels.add((x, y))
    return pixels


def ellipse(cx, cy, rx, ry, angle=0.0):
    pixels = set()
    cos, sin = math.cos(math.radians(angle)), math.sin(math.radians(angle))
    reach = int(math.ceil(max(rx, ry))) + 1
    for y in range(int(cy) - reach, int(cy) + reach + 1):
        for x in range(int(cx) - reach, int(cx) + reach + 1):
            dx, dy = x + 0.5 - cx, y + 0.5 - cy
            u = dx * cos + dy * sin
            v = -dx * sin + dy * cos
            if (u / rx) ** 2 + (v / ry) ** 2 <= 1.0:
                pixels.add((x, y))
    return pixels


def capsule(p0, p1, r0, r1=None):
    r1 = r0 if r1 is None else r1
    (x0, y0), (x1, y1) = p0, p1
    reach = int(math.ceil(max(r0, r1))) + 1
    dx, dy = x1 - x0, y1 - y0
    length2 = dx * dx + dy * dy
    pixels = set()
    for y in range(int(min(y0, y1)) - reach, int(max(y0, y1)) + reach + 1):
        for x in range(int(min(x0, x1)) - reach, int(max(x0, x1)) + reach + 1):
            px, py = x + 0.5 - x0, y + 0.5 - y0
            t = 0.0 if length2 == 0 else max(0.0, min(1.0, (px * dx + py * dy) / length2))
            cx, cy = t * dx, t * dy
            radius = r0 + (r1 - r0) * t
            if math.hypot(px - cx, py - cy) <= radius:
                pixels.add((x, y))
    return pixels


def stroke(points, width):
    pixels = set()
    half = width / 2
    for index in range(len(points) - 1):
        (x0, y0), (x1, y1) = points[index], points[index + 1]
        length = math.hypot(x1 - x0, y1 - y0)
        if length == 0:
            continue
        nx, ny = -(y1 - y0) / length * half, (x1 - x0) / length * half
        pixels |= polygon([(x0 + nx, y0 + ny), (x1 + nx, y1 + ny), (x1 - nx, y1 - ny), (x0 - nx, y0 - ny)])
    return pixels


def mirrored(points, axis=0.0):
    return [(2 * axis - x, y) for x, y in reversed(points)]


def symmetric(points, axis=0.0):
    return points + mirrored(points, axis)


def shift(pixels, dx, dy):
    return {(x + dx, y + dy) for x, y in pixels}


def grow(pixels, amount=1, diagonal=True):
    result = set(pixels)
    steps = [(1, 0), (-1, 0), (0, 1), (0, -1)]
    if diagonal:
        steps += [(1, 1), (-1, 1), (1, -1), (-1, -1)]
    for _ in range(amount):
        result |= {(x + dx, y + dy) for x, y in result for dx, dy in steps}
    return result


def shrink(pixels, amount=1):
    result = set(pixels)
    steps = [(1, 0), (-1, 0), (0, 1), (0, -1)]
    for _ in range(amount):
        result = {(x, y) for x, y in result if all((x + dx, y + dy) in result for dx, dy in steps)}
    return result


def bounds(pixels):
    xs = [x for x, _ in pixels]
    ys = [y for _, y in pixels]
    return min(xs), min(ys), max(xs), max(ys)


class Frame:
    def __init__(self, ox, oy, scale):
        self.ox = ox
        self.oy = oy
        self.scale = scale

    def pt(self, x, y):
        return (self.ox + x * self.scale, self.oy + y * self.scale)

    def pts(self, points):
        return [self.pt(x, y) for x, y in points]

    def length(self, value):
        return value * self.scale
