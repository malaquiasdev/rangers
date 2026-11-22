from . import palette as P
from .backdrop import Layout, halo, paint_backdrop
from .canvas import Canvas
from .emblem import paint_gem
from .geometry import ellipse
from .lettering import paint_ribbon, paint_title, title_width, word_pixels
from .painter import Painter
from .panels import build_panel, frame_panel
from .stage import floor_beams, glow_pools, paint_stage, reflect_panels, scatter_sparkles

ORDER = ["blue", "pink", "red", "green", "yellow", "black"]
RAMPS = {"blue": P.BLUE, "pink": P.PINK, "red": P.RED, "green": P.GREEN, "yellow": P.YELLOW, "black": P.BLACK}


def glow_title(canvas, word, left, top, unit, depth, color=(255, 170, 60)):
    pixels = word_pixels(word, left, top, unit)
    halo(canvas, pixels | {(x + depth, y + depth) for x, y in pixels}, color, 6, 0.4)


POSTER = Layout(
    width=320,
    height=480,
    floor=404,
    mist_height=236,
    rings=[(28, 44, 25, 7), (28, 92, 25, 7), (292, 44, 25, 7), (292, 92, 25, 7)],
    tubes=[(28, 0, 300, 2), (292, 0, 300, 2)],
    slabs=[
        ([(0, 196), (52, 200), (58, 438), (0, 480)], "left"),
        ([(320, 196), (268, 200), (262, 438), (320, 480)], "right"),
    ],
    panel_top=192,
    panel_bottom=400,
)


def paint_panels(canvas, top, bottom, left, width, gap, tall, crop=10):
    for index, key in enumerate(ORDER):
        x = left + index * (width + gap)
        raised = tall if key in ("red", "green") else 0
        panel = build_panel(width, bottom - top + raised, key, RAMPS[key], width * 0.47, crop + raised * 0.2, gem=P.MAGENTA)
        frame_panel(canvas, panel, x, top - raised, RAMPS[key])


def paint_crest(canvas, cx, gem_y, gem_height):
    glow = set(ellipse(cx, gem_y, 20, gem_height * 0.55))
    halo(canvas, glow, (80, 230, 255), 12, 0.5)


def panel_spans(left, width, gap):
    return [(left + index * (width + gap), left + index * (width + gap) + width) for index in range(len(ORDER))]


def paint_scene(canvas, layout, panel_left, ribbon, word_top, unit, depth, gem, floor_ring, reflection):
    paint_backdrop(canvas, layout)
    spans = panel_spans(panel_left, 46, 3)
    tall = 8 if layout is POSTER else 0
    paint_panels(canvas, layout.panel_top, layout.panel_bottom, panel_left, 46, 3, tall, 10 if tall else -6)
    reflect_panels(canvas, spans, layout.panel_bottom, reflection)
    centers = [(left + right) // 2 for left, right in spans]
    glow_pools(canvas, centers, layout.panel_bottom + 6, 27, floor_ring[3] * 0.55, [RAMPS[key][4] for key in ORDER])
    paint_stage(canvas, *floor_ring[:3], floor_ring[3])

    word = "RANGERS"
    left = (layout.width - title_width(word, unit)) / 2 + 1
    band = Painter()
    paint_ribbon(band, *ribbon, P.CRIMSON, P.GOLD)
    band.draw(canvas)
    glow_title(canvas, word, left, word_top, unit, depth)
    title = Painter()
    paint_title(title, word, left, word_top, unit, depth, P.GOLD, P.CRIMSON)
    title.draw(canvas)
    paint_crest(canvas, gem[0], gem[1], gem[2])
    crest = Painter()
    paint_gem(crest, *gem)
    crest.draw(canvas)


def build_poster():
    canvas = Canvas(POSTER.width, POSTER.height)
    paint_scene(canvas, POSTER, 12, (16, 112, 304, 146, 9, 14), 90, 4.4, 4, (160, 62, 56), (160, 442, 150, 30), 40)
    floor_beams(canvas, (14, 52, 108, 160, 212, 268, 306), 480, 372)
    scatter_sparkles(canvas, [(90, 96, 4), (232, 92, 3), (252, 140, 2), (60, 148, 2), (118, 70, 2), (210, 54, 3), (160, 30, 3),
                              (40, 430, 3), (280, 424, 3), (110, 452, 2), (214, 458, 2), (160, 470, 2), (70, 466, 2), (250, 468, 3), (20, 452, 2), (300, 450, 2), (130, 428, 2), (196, 432, 2)])
    return canvas


BANNER = Layout(
    width=320,
    height=160,
    floor=136,
    mist_height=96,
    rings=[(22, 14, 17, 5), (22, 40, 17, 5), (298, 14, 17, 5), (298, 40, 17, 5)],
    tubes=[(22, 0, 100, 2), (298, 0, 100, 2)],
    slabs=[
        ([(0, 60), (26, 62), (30, 140), (0, 160)], "left"),
        ([(320, 60), (294, 62), (290, 140), (320, 160)], "right"),
    ],
    panel_top=62,
    panel_bottom=132,
)


def build_banner():
    canvas = Canvas(BANNER.width, BANNER.height)
    paint_scene(canvas, BANNER, 14, (52, 42, 268, 58, 7, 10), 24, 3.5, 3, (160, 15, 30), (160, 150, 150, 9), 12)
    floor_beams(canvas, (24, 296), 160, 138)
    scatter_sparkles(canvas, [(8, 100, 2), (310, 104, 2), (60, 148, 2), (262, 150, 2), (160, 156, 2)])
    return canvas
