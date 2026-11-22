from . import palette as P
from .geometry import bounds, polygon, shift, stroke

GLYPHS = {
    "R": [[(0.9, 8.0), (0.9, 0.9), (3.9, 0.9), (4.8, 1.8), (4.8, 3.1), (3.9, 4.0), (0.9, 4.0)], [(2.7, 4.0), (4.8, 8.0)]],
    "A": [[(0.6, 8.0), (1.9, 0.9), (3.1, 0.9), (4.4, 8.0)], [(1.5, 5.3), (3.5, 5.3)]],
    "N": [[(0.9, 8.0), (0.9, 0.9), (4.1, 8.0), (4.1, 0.9)]],
    "G": [[(4.7, 2.0), (3.8, 0.9), (1.7, 0.9), (0.9, 1.8), (0.9, 7.1), (1.7, 8.0), (3.9, 8.0), (4.7, 7.1), (4.7, 4.6), (2.9, 4.6)]],
    "E": [[(4.6, 0.9), (0.9, 0.9), (0.9, 8.0), (4.6, 8.0)], [(0.9, 4.45), (3.9, 4.45)]],
    "S": [[(4.6, 1.9), (3.7, 0.9), (1.6, 0.9), (0.8, 1.8), (0.8, 3.1), (1.6, 4.0), (3.4, 4.9), (4.3, 5.8), (4.3, 7.1), (3.5, 8.0), (1.3, 8.0), (0.5, 7.0)]],
}
ADVANCE = 6.0


def glyph_pixels(letter, ox, oy, unit, weight):
    pixels = set()
    half = weight * unit / 2
    for line in GLYPHS[letter]:
        points = [(ox + x * unit, oy + y * unit) for x, y in line]
        pixels |= stroke(points, weight * unit)
        for index, (x, y) in enumerate(points):
            if 0 < index < len(points) - 1:
                pixels |= polygon([(x - half, y - half), (x + half, y - half), (x + half, y + half), (x - half, y + half)])
    return pixels


def word_pixels(word, ox, oy, unit, weight=1.9, gap=0.15):
    pixels = set()
    x = ox
    for letter in word:
        pixels |= glyph_pixels(letter, x, oy, unit, weight)
        x += (ADVANCE + gap) * unit
    return pixels


def title_width(word, unit, gap=0.15):
    return (ADVANCE + gap) * unit * len(word)


def paint_title(painter, word, ox, oy, unit, depth, face, side, weight=1.9):
    pixels = word_pixels(word, ox, oy, unit, weight)
    low_x, low_y, high_x, high_y = bounds(pixels)
    height = max(1, high_y - low_y)
    extrusion = set()
    for step in range(1, depth + 1):
        extrusion |= shift(pixels, step, step)
    extrusion -= pixels

    def side_bias(x, y):
        return -0.3 - 1.4 * ((y - low_y) / (height + depth))

    def face_bias(x, y):
        return 1.1 - 2.3 * ((y - low_y) / height)

    painter.add(extrusion, side, relief=2.0, gain=2.0, strength=2.0, ceiling=3.0, floor=1.0, dither=0.4, bias=side_bias)
    painter.add(pixels, face, relief=2.6, gain=4.6, strength=3.4, ceiling=5.0, floor=1.0, dither=0.3, bias=face_bias)
    return pixels


def paint_ribbon(painter, left, top, right, bottom, notch, fold, ramp, trim):
    mid = (top + bottom) / 2
    height = bottom - top
    body = polygon([(left + fold, top), (right - fold, top), (right - fold, bottom), (left + fold, bottom)])
    tail_left = polygon([(left, top + 4), (left + fold + 5, top + 4), (left + fold + 5, bottom + 4), (left, bottom + 4), (left + notch, mid + 4)])
    tail_right = polygon([(right, top + 4), (right - fold - 5, top + 4), (right - fold - 5, bottom + 4), (right, bottom + 4), (right - notch, mid + 4)])
    painter.add(tail_left, ramp, relief=3.0, gain=3.0, ceiling=3.0, bias=lambda x, y: -1.0)
    painter.add(tail_right, ramp, relief=3.0, gain=3.0, ceiling=3.0, bias=lambda x, y: -1.0)
    painter.add(body, ramp, relief=5.0, gain=3.6, ceiling=4.4, dither=0.6, bias=lambda x, y: 0.5 - 1.1 * (y - top) / height)
    for sign, edge in ((-1, left + fold), (1, right - fold)):
        shade = polygon([(edge, top), (edge + sign * 5, top), (edge + sign * 5, bottom), (edge, bottom)])
        painter.add(shade, ramp, flat=1.2, bias=lambda x, y: 0.0, contour=False, outline=False, dither=0.0)
    top_band = polygon([(left + fold, top), (right - fold, top), (right - fold, top + 2), (left + fold, top + 2)])
    low_band = polygon([(left + fold, bottom - 2), (right - fold, bottom - 2), (right - fold, bottom), (left + fold, bottom)])
    painter.add(top_band, trim, relief=1.5, gain=3.0, ceiling=5.0, contour=False)
    painter.add(low_band, trim, relief=1.5, gain=3.0, ceiling=4.0, contour=False)
    for row in (top + 5, bottom - 6):
        for x in range(left + fold + 4, right - fold - 3, 4):
            painter.add({(x, row), (x + 1, row)}, P.CREAM, flat=3, contour=False, outline=False)
    return body
