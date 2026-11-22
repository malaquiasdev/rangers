import math

BAYER = [[0, 8, 2, 10], [12, 4, 14, 6], [3, 11, 1, 9], [15, 7, 13, 5]]
LIGHT = (-0.46, -0.62, 0.64)
NEIGHBORS = ((1, 0), (-1, 0), (0, 1), (0, -1))
AROUND = NEIGHBORS + ((1, 1), (-1, 1), (1, -1), (-1, -1))


def bayer(x, y):
    return (BAYER[y % 4][x % 4] + 0.5) / 16 - 0.5


def depth_map(pixels):
    low_x = min(x for x, _ in pixels) - 1
    low_y = min(y for _, y in pixels) - 1
    high_x = max(x for x, _ in pixels) + 2
    high_y = max(y for _, y in pixels) + 2
    width, height = high_x - low_x, high_y - low_y
    far = 10 ** 6
    grid = [[0] * width for _ in range(height)]
    for x, y in pixels:
        grid[y - low_y][x - low_x] = far
    for y in range(height):
        for x in range(width):
            if grid[y][x]:
                best = grid[y][x]
                if x > 0:
                    best = min(best, grid[y][x - 1] + 3)
                if y > 0:
                    best = min(best, grid[y - 1][x] + 3)
                    if x > 0:
                        best = min(best, grid[y - 1][x - 1] + 4)
                    if x < width - 1:
                        best = min(best, grid[y - 1][x + 1] + 4)
                grid[y][x] = best
    for y in range(height - 1, -1, -1):
        for x in range(width - 1, -1, -1):
            if grid[y][x]:
                best = grid[y][x]
                if x < width - 1:
                    best = min(best, grid[y][x + 1] + 3)
                if y < height - 1:
                    best = min(best, grid[y + 1][x] + 3)
                    if x < width - 1:
                        best = min(best, grid[y + 1][x + 1] + 4)
                    if x > 0:
                        best = min(best, grid[y + 1][x - 1] + 4)
                grid[y][x] = best
    return {(x, y): grid[y - low_y][x - low_x] / 3.0 for x, y in pixels}


def relief_map(pixels, relief):
    depth = depth_map(pixels)
    top = min(relief, max(depth.values()))
    heights = {}
    for point, value in depth.items():
        edge = min(value, top) / top
        heights[point] = math.sqrt(max(0.0, 1 - (1 - edge) ** 2)) if top > 1 else 1.0
    smooth = {}
    for (x, y) in heights:
        total = 0.0
        for dy in (-1, 0, 1):
            for dx in (-1, 0, 1):
                total += heights.get((x + dx, y + dy), 0.0)
        smooth[(x, y)] = total / 9
    return depth, smooth


def lighting(heights, x, y, strength):
    gx = (heights.get((x + 1, y), 0.0) - heights.get((x - 1, y), 0.0)) / 2
    gy = (heights.get((x, y + 1), 0.0) - heights.get((x, y - 1), 0.0)) / 2
    nx, ny, nz = -gx * strength, -gy * strength, 1.0
    norm = math.sqrt(nx * nx + ny * ny + nz * nz)
    return (nx * LIGHT[0] + ny * LIGHT[1] + nz * LIGHT[2]) / norm


class Part:
    def __init__(self, pixels, ramp, relief=4.0, gain=4.4, strength=3.2, dither=0.55, flat=None, ceiling=4.4,
                 floor=1.0, contour=True, outline=True, bounce=0.8, bias=None):
        self.pixels = pixels
        self.ramp = ramp
        self.relief = relief
        self.gain = gain
        self.strength = strength
        self.dither = dither
        self.flat = flat
        self.ceiling = ceiling
        self.floor = floor
        self.contour = contour
        self.outline = outline
        self.bounce = bounce
        self.bias = bias


class Painter:
    def __init__(self):
        self.parts = []

    def add(self, pixels, ramp, **style):
        pixels = set(pixels)
        if pixels:
            self.parts.append(Part(pixels, ramp, **style))
        return len(self.parts) - 1

    def labels(self):
        owner = {}
        for index, part in enumerate(self.parts):
            for point in part.pixels:
                owner[point] = index
        return owner

    def colors(self):
        owner = self.labels()
        result = {}
        for index, part in enumerate(self.parts):
            visible = [point for point in part.pixels if owner[point] == index]
            if not visible:
                continue
            if part.flat is not None:
                for x, y in visible:
                    tone = part.flat + (part.bias(x, y) if part.bias else 0.0)
                    step = int(math.floor(tone + 0.5 + bayer(x, y) * part.dither)) if part.bias else int(tone)
                    result[(x, y)] = part.ramp[max(0, min(5, step))]
                continue
            depth, heights = relief_map(part.pixels, part.relief)
            for x, y in visible:
                lam = lighting(heights, x, y, part.strength)
                tone = 3 + (lam - LIGHT[2]) * part.gain
                if depth[(x, y)] <= 1.2 and lam < 0.25:
                    tone += part.bounce
                if part.bias:
                    tone += part.bias(x, y)
                tone = max(part.floor, min(part.ceiling, tone))
                step = int(math.floor(tone + 0.5 + bayer(x, y) * part.dither))
                step = max(int(part.floor), min(int(math.ceil(part.ceiling)), step))
                result[(x, y)] = part.ramp[step]
        for (x, y), index in owner.items():
            part = self.parts[index]
            if not part.outline or part.flat is not None:
                continue
            for dx, dy in NEIGHBORS:
                other = owner.get((x + dx, y + dy))
                if other is not None and other > index and self.parts[other].contour:
                    result[(x, y)] = part.ramp[0]
                    break
        return owner, result

    def draw(self, canvas, rim=None, outline=True):
        owner, result = self.colors()
        for (x, y), color in result.items():
            canvas.put(x, y, color)
        if not outline:
            return result
        edge = {}
        for x, y in list(owner):
            for dx, dy in NEIGHBORS:
                point = (x + dx, y + dy)
                if point not in owner:
                    top = max(owner[(x, y)], edge.get(point, (-1,))[0])
                    edge[point] = (top,)
        for point, (index,) in edge.items():
            canvas.put(point[0], point[1], self.parts[index].ramp[0])
        if rim is not None:
            inner, outer = rim
            solid = set(owner) | set(edge)
            first = set()
            for x, y in solid:
                for dx, dy in AROUND:
                    point = (x + dx, y + dy)
                    if point not in solid:
                        first.add(point)
            second = set()
            for x, y in first:
                for dx, dy in AROUND:
                    point = (x + dx, y + dy)
                    if point not in solid and point not in first:
                        second.add(point)
            for x, y in first:
                canvas.put(x, y, inner)
            for x, y in second:
                canvas.put(x, y, outer)
        return result
