from .backdrop import halo, paint_platform, sparkle
from .canvas import mix
from .geometry import ellipse, polygon
from .painter import bayer


def reflect_panels(canvas, spans, bottom, reach, strength=0.42):
    for left, right in spans:
        for depth in range(1, reach + 1):
            fade = strength * (1 - depth / (reach + 1)) ** 1.3
            for x in range(left, right):
                source = canvas.get(x, bottom - depth)
                under = canvas.get(x, bottom + depth)
                if source is None or under is None or fade * 1.5 <= bayer(x, depth) + 0.5:
                    continue
                canvas.put(x, bottom + depth, mix(under, source, fade * 1.6))


def glow_pools(canvas, centers, y, rx, ry, tones):
    for cx, tone in zip(centers, tones):
        for px, py in ellipse(cx, y, rx, ry):
            under = canvas.get(px, py)
            if under is None:
                continue
            depth = 1 - (((px + 0.5 - cx) / rx) ** 2 + ((py + 0.5 - y) / ry) ** 2)
            if depth * 1.1 > bayer(px, py) + 0.5:
                canvas.put(px, py, mix(under, tone, min(0.7, 0.25 + depth * 0.7)))


def floor_beams(canvas, xs, bottom, top, tint=(150, 235, 230)):
    for x in xs:
        wedge = polygon([(x - 2, bottom), (x + 2, bottom), (x + 7, top), (x - 7, top)])
        for px, py in wedge:
            under = canvas.get(px, py)
            along = (py - top) / max(1, bottom - top)
            fade = min(1.0, along * 2.2) * min(1.0, (1 - along) * 3 + 0.35) * along ** 0.3
            if under is not None and fade * 1.3 > bayer(px, py) + 0.5:
                canvas.put(px, py, mix(under, tint, 0.75 * fade))


def paint_stage(canvas, cx, cy, rx, ry):
    paint_platform(canvas, cx, cy, rx, ry)
    inner = ellipse(cx, cy, rx * 0.22, ry * 0.22)
    halo(canvas, inner, (90, 240, 230), 6, 0.6)
    for x, y in inner:
        canvas.put(x, y, mix(canvas.get(x, y), (210, 255, 250), 0.6))


def scatter_sparkles(canvas, points):
    for x, y, size in points:
        sparkle(canvas, x, y, size)
