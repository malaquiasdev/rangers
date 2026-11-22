import math
import random
import subprocess
import sys
from pathlib import Path

UNIT = 8
COLUMNS = 160
ROWS = 80

SKY = ["#0b0b2a", "#12123a", "#1a1850", "#26205f", "#3a2470", "#54307c", "#7a3c85", "#a04b8a"]
BAND = 7

FONT = {
    "R": ["11110", "10001", "10001", "11110", "10100", "10010", "10001"],
    "A": ["01110", "10001", "10001", "11111", "10001", "10001", "10001"],
    "N": ["10001", "11001", "10101", "10011", "10001", "10001", "10001"],
    "G": ["01110", "10001", "10000", "10111", "10001", "10001", "01111"],
    "E": ["11111", "10000", "10000", "11110", "10000", "10000", "11111"],
    "S": ["01111", "10000", "10000", "01110", "00001", "00001", "11110"],
    " ": ["00000"] * 7,
}

HELMET_HALF = [
    "...ooooo",
    "..ohhLLL",
    ".ohhLLLL",
    "ohLLLLLL",
    "oLLLLLLL",
    "oLLLLLLL",
    "oLLooooo",
    "oLLovvvv",
    "oLLovwwv",
    "oLLovvvv",
    "oLLooooo",
    "oLLLLLLL",
    ".oLsssss",
    "..oooooo",
]

HELMETS = ["#e53935", "#3a3a46", "#ff6fa8", "#2f7bff", "#ffd23f", "#2ecc71"]
TITLE_BANDS = ["#fff3b0", "#ffe066", "#ffd23f", "#ffb627", "#ff9f1c", "#e8710a", "#c1560a"]


def shade(color, factor):
    red, green, blue = (int(color[index:index + 2], 16) for index in (1, 3, 5))
    if factor < 1:
        channels = (red * factor, green * factor, blue * factor)
    else:
        mix = factor - 1
        channels = (red + (255 - red) * mix, green + (255 - green) * mix, blue + (255 - blue) * mix)
    return "#%02x%02x%02x" % tuple(max(0, min(255, round(value))) for value in channels)


class Canvas:
    def __init__(self):
        self.cells = {}

    def put(self, column, row, color):
        if 0 <= column < COLUMNS and 0 <= row < ROWS:
            self.cells[(column, row)] = color

    def rect(self, column, row, width, height, color):
        for y in range(row, row + height):
            for x in range(column, column + width):
                self.put(x, y, color)

    def svg(self):
        rows = {}
        for (column, row), color in self.cells.items():
            rows.setdefault(row, {})[column] = color
        parts = []
        for row in sorted(rows):
            line = rows[row]
            column = 0
            while column < COLUMNS:
                if column not in line:
                    column += 1
                    continue
                color = line[column]
                end = column
                while end + 1 in line and line[end + 1] == color:
                    end += 1
                parts.append(
                    '<rect x="%d" y="%d" width="%d" height="%d" fill="%s"/>'
                    % (column * UNIT, row * UNIT, (end - column + 1) * UNIT, UNIT, color)
                )
                column = end + 1
        width, height = COLUMNS * UNIT, ROWS * UNIT
        return (
            '<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="%d" viewBox="0 0 %d %d" '
            'shape-rendering="crispEdges">%s</svg>' % (width, height, width, height, "".join(parts))
        )


def sky(canvas):
    for row in range(ROWS):
        band = min(row // BAND, len(SKY) - 1)
        following = min(band + 1, len(SKY) - 1)
        for column in range(COLUMNS):
            color = SKY[band]
            if row % BAND == BAND - 1 and (column + row) % 2 == 0:
                color = SKY[following]
            canvas.put(column, row, color)


def clear_of_text(column, row):
    return not (34 <= column <= 126 and 8 <= row <= 38)


def stars(canvas, generator):
    for _ in range(60):
        column, row = generator.randrange(COLUMNS), generator.randrange(46)
        if clear_of_text(column, row):
            canvas.put(column, row, generator.choice(["#ffffff", "#fff3b0", "#bfe3ff"]))
    for _ in range(9):
        column, row = generator.randrange(3, COLUMNS - 3), generator.randrange(2, 40)
        if not clear_of_text(column, row):
            continue
        for dx, dy in ((0, 0), (1, 0), (-1, 0), (0, 1), (0, -1)):
            canvas.put(column + dx, row + dy, "#ffffff" if dx == dy == 0 else "#bfe3ff")


def moon(canvas):
    centre_column, centre_row, radius = 132, 13, 7
    for row in range(centre_row - radius, centre_row + radius + 1):
        for column in range(centre_column - radius, centre_column + radius + 1):
            if (column - centre_column) ** 2 + (row - centre_row) ** 2 <= radius * radius:
                canvas.put(column, row, "#fff6d5")
    for column, row in ((129, 10), (135, 14), (131, 17), (136, 10)):
        canvas.rect(column, row, 2, 2, "#e6d9a8")


def hills(canvas):
    for column in range(COLUMNS):
        height = 9 + int(3 * math.sin(column / 11) + 2 * math.sin(column / 4.3))
        canvas.rect(column, ROWS - height - 12, 1, height + 12, "#5a2f7d")
    for column in range(COLUMNS):
        height = 7 + int(2 * math.sin(column / 7 + 2) + 2 * math.sin(column / 3.1))
        canvas.rect(column, ROWS - height - 10, 1, height + 10, "#3d2368")


def city(canvas, generator):
    for layer, color, floor, limits in (
        (0, "#1d1446", 62, (8, 16)),
        (1, "#0e0a24", 66, (6, 14)),
    ):
        column = generator.randrange(0, 4)
        while column < COLUMNS:
            width = generator.randrange(5, 11)
            height = generator.randrange(*limits)
            top = ROWS - height
            canvas.rect(column, top, width, ROWS - top, color)
            if layer == 1:
                for window_row in range(top + 2, ROWS - 3, 3):
                    for window_column in range(column + 1, column + width - 1, 2):
                        if generator.random() < 0.28:
                            canvas.put(window_column, window_row, "#ffd23f")
            column += width + generator.randrange(0, 3)
    canvas.rect(0, ROWS - 3, COLUMNS, 3, "#0a0818")


def write(canvas, text, column, row, scale, colors, shadow=None):
    for index, letter in enumerate(text):
        glyph = FONT[letter]
        origin = column + index * 6 * scale
        for glyph_row, bits in enumerate(glyph):
            for glyph_column, bit in enumerate(bits):
                if bit == "1":
                    color = colors[glyph_row % len(colors)] if isinstance(colors, list) else colors
                    x, y = origin + glyph_column * scale, row + glyph_row * scale
                    if shadow:
                        canvas.rect(x + 1, y + 1, scale, scale, shadow)
                    canvas.rect(x, y, scale, scale, color)
    return len(text) * 6 * scale - scale


def title(canvas):
    scale = 2
    glyph_width = 5 * scale
    full = len("RANGERS") * 6 * scale - scale
    column = (COLUMNS - full) // 2
    for index, letter in enumerate("RANGERS"):
        origin = column + index * 6 * scale
        for glyph_row, bits in enumerate("".join(FONT[letter])):
            pass
    shadow = "#1b1030"
    for index, letter in enumerate("RANGERS"):
        origin = column + index * 6 * scale
        for glyph_row, bits in enumerate(FONT[letter]):
            for glyph_column, bit in enumerate(bits):
                if bit == "1":
                    x = origin + glyph_column * scale
                    y = 16 + glyph_row * scale
                    canvas.rect(x + 1, y + 1, scale, scale, shadow)
    for index, letter in enumerate("RANGERS"):
        origin = column + index * 6 * scale
        for glyph_row, bits in enumerate(FONT[letter]):
            for glyph_column, bit in enumerate(bits):
                if bit == "1":
                    canvas.rect(origin + glyph_column * scale, 16 + glyph_row * scale, scale, scale, TITLE_BANDS[glyph_row])
    return glyph_width


def helmet(canvas, column, row, color):
    palette = {
        "o": "#14142b",
        "L": color,
        "h": shade(color, 1.45),
        "v": "#101830",
        "w": "#9fd0ff",
        "s": "#c8ccd8",
    }
    for sprite_row, half in enumerate(HELMET_HALF):
        full = half + half[::-1]
        for sprite_column, symbol in enumerate(full):
            if symbol == ".":
                continue
            fill = palette[symbol]
            if symbol == "L" and sprite_column >= 11:
                fill = shade(color, 0.78)
            canvas.put(column + sprite_column, row + sprite_row, fill)
    for offset in range(-1, 17):
        canvas.put(column + offset, row + len(HELMET_HALF), "#0a0818")


def helmets(canvas):
    spacing = 4
    total = 6 * 16 + 5 * spacing
    column = (COLUMNS - total) // 2
    for color in HELMETS:
        helmet(canvas, column, 40, color)
        column += 16 + spacing


def main():
    folder = Path(sys.argv[1])
    generator = random.Random(1990)
    canvas = Canvas()
    sky(canvas)
    stars(canvas, generator)
    moon(canvas)
    hills(canvas)
    city(canvas, generator)
    title(canvas)
    helmets(canvas)
    svg = folder / "banner.svg"
    svg.write_text(canvas.svg())
    subprocess.run(["rsvg-convert", "-w", "1280", "-h", "640", "-o", str(folder / "banner.png"), str(svg)], check=True)


main()
