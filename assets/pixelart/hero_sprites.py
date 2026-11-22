from . import palette as P
from .canvas import Canvas, mix
from .figure import paint_hero
from .geometry import bounds
from .painter import Painter

SIZE = (48, 64)
RIM_LIGHT = (150, 200, 255)
RAMPS = {"blue": P.BLUE, "pink": P.PINK, "red": P.RED, "green": P.GREEN, "yellow": P.YELLOW, "black": P.BLACK}


def fit_scale(key):
    probe = Painter()
    paint_hero(probe, key, RAMPS[key], (0, 0), 1.0)
    low_x, low_y, high_x, high_y = bounds(set().union(*[part.pixels for part in probe.parts]))
    return min(1.0, (SIZE[0] - 2) / (high_x - low_x + 1), (SIZE[1] - 2) / (high_y - low_y + 1))


def build_hero_sprite(key):
    scale = fit_scale(key)
    probe = Painter()
    paint_hero(probe, key, RAMPS[key], (0, 0), scale)
    low_x, low_y, high_x, high_y = bounds(set().union(*[part.pixels for part in probe.parts]))
    origin = (SIZE[0] / 2 - (low_x + high_x + 1) / 2, SIZE[1] - 1 - (high_y + 1))
    painter = Painter()
    paint_hero(painter, key, RAMPS[key], origin, scale)
    canvas = Canvas(*SIZE)
    owner, colors = painter.colors()
    painter.draw(canvas)
    for (x, y), index in owner.items():
        if (x + 1, y) in owner or (x, y + 1) in owner:
            continue
        color = canvas.get(x, y)
        if color is not None and sum(color) > 120 and index >= 0:
            canvas.put(x, y, mix(color, RIM_LIGHT, 0.35))
    return canvas
