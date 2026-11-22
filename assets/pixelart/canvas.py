import struct
import zlib
from pathlib import Path


def rgb(color):
    if isinstance(color, tuple):
        return color
    return tuple(int(color[index:index + 2], 16) for index in (1, 3, 5))


def mix(first, second, amount):
    first, second = rgb(first), rgb(second)
    return tuple(round(a + (b - a) * amount) for a, b in zip(first, second))


class Canvas:
    def __init__(self, width, height, fill=None):
        self.width = width
        self.height = height
        self.pixels = [fill] * (width * height)

    def inside(self, x, y):
        return 0 <= x < self.width and 0 <= y < self.height

    def get(self, x, y):
        if not self.inside(x, y):
            return None
        return self.pixels[y * self.width + x]

    def put(self, x, y, color):
        if self.inside(x, y):
            self.pixels[y * self.width + x] = None if color is None else rgb(color)

    def rect(self, x, y, width, height, color):
        for row in range(y, y + height):
            for column in range(x, x + width):
                self.put(column, row, color)

    def line(self, x0, y0, x1, y1, color):
        dx, dy = abs(x1 - x0), -abs(y1 - y0)
        step_x = 1 if x0 < x1 else -1
        step_y = 1 if y0 < y1 else -1
        error = dx + dy
        while True:
            self.put(x0, y0, color)
            if x0 == x1 and y0 == y1:
                return
            doubled = 2 * error
            if doubled >= dy:
                error += dy
                x0 += step_x
            if doubled <= dx:
                error += dx
                y0 += step_y

    def ellipse(self, cx, cy, rx, ry, color):
        for y in range(int(cy - ry) - 1, int(cy + ry) + 2):
            for x in range(int(cx - rx) - 1, int(cx + rx) + 2):
                if ((x + 0.5 - cx) / rx) ** 2 + ((y + 0.5 - cy) / ry) ** 2 <= 1:
                    self.put(x, y, color)

    def blit(self, other, x, y, flip=False):
        for row in range(other.height):
            for column in range(other.width):
                color = other.pixels[row * other.width + column]
                if color is not None:
                    target = x + (other.width - 1 - column if flip else column)
                    self.put(target, y + row, color)

    def flipped(self):
        result = Canvas(self.width, self.height)
        for row in range(self.height):
            for column in range(self.width):
                result.pixels[row * self.width + column] = self.pixels[row * self.width + self.width - 1 - column]
        return result

    def scaled(self, factor):
        result = Canvas(self.width * factor, self.height * factor)
        for row in range(self.height):
            for column in range(self.width):
                color = self.pixels[row * self.width + column]
                if color is None:
                    continue
                for dy in range(factor):
                    base = (row * factor + dy) * result.width + column * factor
                    for dx in range(factor):
                        result.pixels[base + dx] = color
        return result

    def save(self, path, factor=1):
        source = self.scaled(factor) if factor != 1 else self
        raw = bytearray()
        for row in range(source.height):
            raw.append(0)
            for color in source.pixels[row * source.width:(row + 1) * source.width]:
                raw.extend((0, 0, 0, 0) if color is None else (*color, 255))

        def chunk(kind, data):
            body = kind + data
            return struct.pack(">I", len(data)) + body + struct.pack(">I", zlib.crc32(body))

        header = struct.pack(">IIBBBBB", source.width, source.height, 8, 6, 0, 0, 0)
        png = b"\x89PNG\r\n\x1a\n" + chunk(b"IHDR", header) + chunk(b"IDAT", zlib.compress(bytes(raw), 9)) + chunk(b"IEND", b"")
        Path(path).parent.mkdir(parents=True, exist_ok=True)
        Path(path).write_bytes(png)
