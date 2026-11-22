from . import palette as P
from .backdrop import pick
from .canvas import Canvas, mix
from .geometry import ellipse, polygon
from .geometry import grow
from .helmets import helmet_shell, paint_helmet
from .painter import Painter

WHITE = (255, 255, 255)


def pale(ramp):
    return [
        mix(ramp[1], WHITE, 0.30),
        mix(ramp[2], WHITE, 0.50),
        mix(ramp[3], WHITE, 0.62),
        mix(ramp[4], WHITE, 0.70),
        mix(ramp[4], WHITE, 0.84),
        WHITE,
    ]


def paint_backdrop_panel(canvas, width, height, ramp):
    tones = pale(ramp)
    for y in range(height):
        for x in range(width):
            vertical = 3.6 - 1.7 * (y / height)
            streak = (x * 0.9 + y * 0.35) % 34
            value = vertical + (0.9 if 4 < streak < 9 else 0.0) + (0.5 if x < 3 else 0.0) - (0.9 if x > width - 4 else 0.0)
            canvas.put(x, y, pick(tones, value, x, y, 3.0))


def paint_suit(painter, width, height, ramp, top, helmet_radius):
    cx = width / 2
    shoulders = polygon([(-16, top - 2), (width + 16, top - 2), (width + 10, height + 30), (-10, height + 30)])
    painter.add(shoulders, ramp, relief=18.0, gain=3.2, strength=1.8, ceiling=4.2, dither=0.5)
    for anchor in (-6, width + 6):
        painter.add(ellipse(anchor, top + 12, 17, 15), ramp, relief=11.0, gain=4.4, strength=2.6, ceiling=4.6)
    diamond_top = top + 16
    reach = 21
    diamond = polygon([(cx, diamond_top), (cx + reach, diamond_top + 34), (cx, diamond_top + 84), (cx - reach, diamond_top + 34)])
    painter.add(diamond, P.WHITE, flat=3, dither=0.0)
    right_face = polygon([(cx, diamond_top), (cx + reach, diamond_top + 34), (cx, diamond_top + 84)])
    painter.add(right_face, P.WHITE, flat=1.5, bias=lambda x, y: (x - cx) * -0.02 + (y - diamond_top) * 0.012, dither=0.7, contour=False, outline=False)
    left_face = polygon([(cx, diamond_top), (cx - reach, diamond_top + 34), (cx, diamond_top + 84)])
    painter.add(left_face, P.WHITE, flat=3.4, bias=lambda x, y: (cx - x) * -0.03 + (y - diamond_top) * -0.012, dither=0.7, contour=False, outline=False)
    ridge = polygon([(cx - 1, diamond_top + 3), (cx + 1, diamond_top + 3), (cx + 1, diamond_top + 80), (cx - 1, diamond_top + 80)])
    painter.add(ridge, P.WHITE, flat=4.5, contour=False, outline=False)
    rim = polygon([(cx, diamond_top + 10), (cx + 12, diamond_top + 34), (cx, diamond_top + 62), (cx - 12, diamond_top + 34)])
    painter.add(rim - polygon([(cx, diamond_top + 17), (cx + 8, diamond_top + 34), (cx, diamond_top + 53), (cx - 8, diamond_top + 34)]), ramp, flat=3.0, bias=lambda x, y: (cx - x) * -0.04, dither=0.5, contour=False, outline=False)


def build_panel(width, height, key, ramp, helmet_radius, helmet_top, gem=None):
    canvas = Canvas(width, height)
    paint_backdrop_panel(canvas, width, height, ramp)
    painter = Painter()
    cx = width / 2
    helmet_y = helmet_top + helmet_radius * 1.05
    paint_suit(painter, width, height, ramp, helmet_y + helmet_radius * 0.82, helmet_radius)
    shell = helmet_shell(cx, helmet_y, helmet_radius)
    under = {(x, y) for x, y in grow(shell, 3) - shell if y > helmet_y + helmet_radius * 0.5}
    painter.add(under, ramp, flat=1, contour=False, outline=False, bias=lambda x, y: -0.2 + 0.4 * ((x - cx) ** 2) / (helmet_radius ** 2), dither=0.7)
    paint_helmet(painter, key, cx, helmet_y, helmet_radius, ramp, gem=gem)
    painter.draw(canvas, outline=False)
    return canvas


def frame_panel(target, canvas, x, y, ramp):
    target.blit(canvas, x, y)
    width, height = canvas.width, canvas.height
    for column in range(-1, width + 1):
        target.put(x + column, y - 1, ramp[0])
        target.put(x + column, y + height, ramp[0])
    for row in range(-1, height + 1):
        target.put(x - 1, y + row, ramp[0])
        target.put(x + width, y + row, ramp[0])
    for column in range(width):
        target.put(x + column, y, mix(ramp[4], WHITE, 0.6))
    for row in range(height):
        target.put(x, y + row, mix(ramp[4], WHITE, 0.45))
        target.put(x + width - 1, y + row, ramp[1])
