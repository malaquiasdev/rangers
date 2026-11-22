from . import palette as P
from .backdrop import Layout, halo, paint_backdrop, paint_platform, sparkle
from .canvas import Canvas, mix
from .emblem import paint_gem
from .figure import paint_hero
from .geometry import bounds, ellipse
from .lettering import paint_ribbon, paint_title, title_width, word_pixels
from .painter import Painter, bayer
from .panels import build_panel, frame_panel

ORDER = ["blue", "pink", "red", "green", "yellow", "black"]
RAMPS = {"blue": P.BLUE, "pink": P.PINK, "red": P.RED, "green": P.GREEN, "yellow": P.YELLOW, "black": P.BLACK}
WHITE = (252, 253, 255)


def place_hero(key, center_x, baseline, scale, gem=P.MAGENTA):
    probe = Painter()
    paint_hero(probe, key, RAMPS[key], (0, 0), scale, gem=gem)
    everything = set().union(*[part.pixels for part in probe.parts])
    low_x, low_y, high_x, high_y = bounds(everything)
    origin = (center_x - (low_x + high_x) / 2, baseline - high_y)
    painter = Painter()
    paint_hero(painter, key, RAMPS[key], origin, scale, gem=gem)
    return painter


def hero_rim(key):
    return (WHITE, RAMPS[key][4])


def ground_shadow(canvas, cx, cy, rx, ry):
    for x, y in ellipse(cx, cy, rx, ry):
        under = canvas.get(x, y)
        if under is None:
            continue
        depth = 1 - (((x + 0.5 - cx) / rx) ** 2 + ((y + 0.5 - cy) / ry) ** 2)
        if depth > bayer(x, y) * 0.9 + 0.2:
            canvas.put(x, y, mix(under, (2, 6, 14), 0.7))


def reflect(canvas, painter, baseline, depth):
    owner, colors = painter.colors()
    for (x, y), color in colors.items():
        if y < baseline - depth:
            continue
        mirrored = 2 * baseline - y + 1
        fade = 0.3 * (1 - (mirrored - baseline) / (depth + 2))
        under = canvas.get(x, mirrored)
        if under is not None and fade > 0 and fade > bayer(x, mirrored) * 0.2 + 0.08:
            canvas.put(x, mirrored, mix(under, color, fade))


def draw_heroes(canvas, spec, centers, baseline, scales, rim=True, depth=18):
    for key in spec:
        ground_shadow(canvas, centers[key], baseline - 1, 30 * scales[key] / 1.6, 5)
    painters = {key: place_hero(key, centers[key], baseline, scales[key]) for key in spec}
    for key in spec:
        reflect(canvas, painters[key], baseline, depth)
    for key in spec:
        painters[key].draw(canvas, rim=hero_rim(key) if rim else None)


def glow_title(canvas, word, left, top, unit, depth, color=(255, 170, 60)):
    pixels = word_pixels(word, left, top, unit)
    halo(canvas, pixels | {(x + depth, y + depth) for x, y in pixels}, color, 6, 0.4)


POSTER = Layout(
    width=320,
    height=480,
    floor=398,
    mist_height=236,
    rings=[(28, 44, 25, 7), (28, 92, 25, 7), (292, 44, 25, 7), (292, 92, 25, 7)],
    tubes=[(28, 0, 300, 2), (292, 0, 300, 2)],
    slabs=[
        ([(0, 196), (52, 200), (58, 438), (0, 480)], "left"),
        ([(320, 196), (268, 200), (262, 438), (320, 480)], "right"),
    ],
    panel_top=176,
    panel_bottom=372,
)


def paint_panels(canvas, top, bottom, left, width, gap, tall):
    for index, key in enumerate(ORDER):
        x = left + index * (width + gap)
        raised = tall if key in ("red", "green") else 0
        panel = build_panel(width, bottom - top + raised, key, RAMPS[key], width * 0.47, 10 + raised * 0.2, gem=P.MAGENTA)
        frame_panel(canvas, panel, x, top - raised, RAMPS[key])


def darken_foot(canvas, left, right, top, bottom, reach):
    for y in range(bottom - reach, bottom):
        amount = ((y - (bottom - reach)) / reach) ** 1.4 * 0.7
        for x in range(left, right):
            under = canvas.get(x, y)
            if under is not None and amount * 1.2 > bayer(x, y) + 0.5 - 0.15:
                canvas.put(x, y, mix(under, (6, 10, 22), min(0.8, amount * 1.1)))


def paint_crest(canvas, cx, gem_y, gem_height):
    glow = set(ellipse(cx, gem_y, 20, gem_height * 0.55))
    halo(canvas, glow, (80, 230, 255), 12, 0.5)


def build_poster():
    canvas = Canvas(POSTER.width, POSTER.height)
    paint_backdrop(canvas, POSTER)
    paint_platform(canvas, 160, 456, 150, 22)
    paint_panels(canvas, POSTER.panel_top, POSTER.panel_bottom, 12, 46, 3, 8)
    darken_foot(canvas, 8, 312, POSTER.panel_top, POSTER.panel_bottom, 56)

    unit = 4.4
    word = "RANGERS"
    left = (POSTER.width - title_width(word, unit)) / 2 + 1
    ribbon = Painter()
    paint_ribbon(ribbon, 16, 112, 304, 146, 9, 14, P.CRIMSON, P.GOLD)
    ribbon.draw(canvas)
    glow_title(canvas, word, left, 90, unit, 4)
    title = Painter()
    paint_title(title, word, left, 90, unit, 4, P.GOLD, P.CRIMSON)
    title.draw(canvas)
    paint_crest(canvas, 160, 62, 56)
    crest = Painter()
    paint_gem(crest, 160, 62, 56)
    crest.draw(canvas)
    for x, y, size in ((90, 96, 4), (232, 92, 3), (252, 140, 2), (60, 148, 2), (118, 70, 2), (210, 54, 3), (160, 30, 3)):
        sparkle(canvas, x, y, size)

    centers = {"blue": 36, "pink": 84, "red": 134, "green": 188, "yellow": 234, "black": 288}
    scales = {key: 1.86 for key in ORDER}
    scales["red"] = 2.02
    scales["black"] = 1.8
    order = ["blue", "black", "pink", "yellow", "green", "red"]
    draw_heroes(canvas, order, centers, 458, scales)
    return canvas


BANNER = Layout(
    width=320,
    height=160,
    floor=132,
    mist_height=96,
    rings=[(22, 14, 17, 5), (22, 40, 17, 5), (298, 14, 17, 5), (298, 40, 17, 5)],
    tubes=[(22, 0, 100, 2), (298, 0, 100, 2)],
    slabs=[
        ([(0, 60), (26, 62), (30, 140), (0, 160)], "left"),
        ([(320, 60), (294, 62), (290, 140), (320, 160)], "right"),
    ],
    panel_top=52,
    panel_bottom=112,
)


def build_banner():
    canvas = Canvas(BANNER.width, BANNER.height)
    paint_backdrop(canvas, BANNER)
    paint_platform(canvas, 160, 150, 150, 9)
    paint_panels(canvas, 62, 122, 14, 46, 3, 0)
    darken_foot(canvas, 10, 310, 62, 122, 24)

    unit = 3.5
    word = "RANGERS"
    left = (BANNER.width - title_width(word, unit)) / 2 + 1
    ribbon = Painter()
    paint_ribbon(ribbon, 52, 42, 268, 58, 7, 10, P.CRIMSON, P.GOLD)
    ribbon.draw(canvas)
    glow_title(canvas, word, left, 24, unit, 3)
    title = Painter()
    paint_title(title, word, left, 24, unit, 3, P.GOLD, P.CRIMSON)
    title.draw(canvas)
    crest = Painter()
    paint_gem(crest, 160, 15, 30)
    crest.draw(canvas)

    centers = {"blue": 36, "pink": 84, "red": 134, "green": 188, "yellow": 234, "black": 288}
    scales = {key: 1.0 for key in ORDER}
    scales["red"] = 1.1
    order = ["blue", "black", "pink", "yellow", "green", "red"]
    draw_heroes(canvas, order, centers, 156, scales)
    return canvas
