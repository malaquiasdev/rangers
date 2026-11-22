from . import palette as P
from .geometry import Frame, capsule, ellipse, grow, polygon, shift, symmetric

JAW = [(-0.95, 0.0), (-0.84, 0.55), (-0.56, 0.9), (-0.3, 1.02), (0.3, 1.02), (0.56, 0.9), (0.84, 0.55), (0.95, 0.0)]

VISORS = {
    "red": [(0.0, -0.4), (0.7, -0.5), (1.0, -0.34), (1.04, 0.2), (0.82, 0.52), (0.46, 0.22), (0.0, 0.3)],
    "blue": [(0.0, -0.42), (0.95, -0.44), (1.04, 0.0), (0.6, 0.4), (0.0, 0.66)],
    "pink": [(0.0, -0.16), (0.5, -0.34), (0.98, -0.62), (1.1, -0.5), (0.94, -0.02), (0.5, 0.3), (0.0, 0.16)],
    "green": [(0.0, -0.02), (0.3, -0.46), (1.0, -0.22), (0.94, 0.18), (0.4, 0.32), (0.0, 0.1)],
    "yellow": [(0.0, -0.22), (0.46, -0.42), (1.12, -0.36), (0.9, 0.0), (0.5, 0.16), (0.0, 0.2)],
    "black": [(0.0, -0.52), (0.86, -0.52), (0.98, -0.1), (0.9, 0.4), (0.0, 0.4)],
}

SPIKES = {
    "red": [[(0.0, -1.5), (0.32, -0.8), (-0.32, -0.8)]],
    "blue": [
        [(0.0, -1.5), (0.18, -0.9), (-0.18, -0.9)],
        [(0.62, -1.28), (0.72, -0.62), (0.34, -0.9)],
        [(-0.62, -1.28), (-0.72, -0.62), (-0.34, -0.9)],
    ],
    "pink": [
        [(0.46, -1.0), (1.12, -1.5), (0.94, -0.7)],
        [(-0.46, -1.0), (-1.12, -1.5), (-0.94, -0.7)],
        [(0.0, -1.34), (0.2, -0.9), (-0.2, -0.9)],
    ],
    "green": [
        [(0.7, -0.85), (1.2, -1.3), (1.02, -0.5)],
        [(-0.7, -0.85), (-1.2, -1.3), (-1.02, -0.5)],
    ],
    "yellow": [
        [(0.5, -0.95), (1.02, -1.62), (0.98, -0.5)],
        [(-0.5, -0.95), (-1.02, -1.62), (-0.98, -0.5)],
    ],
    "black": [
        [(0.0, -1.28), (0.34, -1.18), (0.5, -0.9), (-0.5, -0.9), (-0.34, -1.18)],
    ],
}

CHEEKS = {
    "red": [(0.86, -0.1), (1.12, 0.15), (0.9, 0.5)],
    "blue": [(0.9, -0.05), (1.22, 0.3), (0.84, 0.55)],
    "pink": [(0.9, 0.0), (1.08, 0.3), (0.84, 0.55)],
    "green": [(0.9, 0.0), (1.14, 0.4), (0.84, 0.6)],
    "yellow": [(0.9, 0.0), (1.2, 0.5), (0.84, 0.62)],
    "black": [(0.92, -0.2), (1.2, 0.0), (1.22, 0.55), (0.86, 0.6)],
}


def helmet_shell(cx, cy, radius):
    frame = Frame(cx, cy, radius)
    return ellipse(cx, cy - radius * 0.08, radius * 1.0, radius * 1.04) | polygon(frame.pts(JAW))


def paint_helmet(painter, kind, cx, cy, radius, ramp, glass=P.GLASS, gem=None, detail=1.0):
    frame = Frame(cx, cy, radius)
    relief = max(3.0, radius * 0.62)

    for spike in SPIKES[kind]:
        painter.add(polygon(frame.pts(spike)), ramp, relief=max(1.5, radius * 0.18), gain=3.0)

    shell = helmet_shell(cx, cy, radius)
    painter.add(shell, ramp, relief=relief, gain=4.6, strength=3.6, ceiling=4.6)

    cheek = CHEEKS[kind]
    painter.add(polygon(frame.pts(cheek)), ramp, relief=max(1.5, radius * 0.2), gain=3.2)
    painter.add(polygon(frame.pts([(-x, y) for x, y in cheek])), ramp, relief=max(1.5, radius * 0.2), gain=3.2)

    visor = polygon(frame.pts(symmetric([(x, y * 0.88 - 0.02) for x, y in VISORS[kind]])))
    rim = (grow(visor, 1) - visor) & shell
    painter.add(rim, ramp, flat=1, contour=False, outline=False)
    if radius >= 9:
        plate = (grow(shift(visor, 0, -2), 1) - grow(visor, 1)) & shell
        painter.add(plate, ramp, flat=3.6, bias=lambda x, y: (cy - y) * 0.02, dither=0.6, contour=False, outline=False)
    painter.add(visor, glass, relief=max(2.0, radius * 0.3), gain=2.2, strength=2.0, ceiling=2.4, floor=1.0, dither=0.12)

    streak_a = polygon(frame.pts([(-0.68, -0.34), (-0.46, -0.4), (-0.12, 0.2), (-0.34, 0.24)])) & visor
    streak_b = polygon(frame.pts([(-0.06, -0.4), (0.06, -0.41), (0.3, 0.06), (0.18, 0.08)])) & visor
    painter.add(streak_a, glass, flat=5, contour=False, outline=False)
    painter.add(streak_b, glass, flat=4, contour=False, outline=False)
    if radius >= 5:
        painter.add(ellipse(cx - radius * 0.34, cy - radius * 0.72, max(1.0, radius * 0.16), max(0.8, radius * 0.1)) & shell, ramp, flat=5, contour=False, outline=False)
    if radius >= 9:
        reflection = visor & polygon(frame.pts([(-1.2, 0.02), (1.2, -0.12), (1.2, 0.8), (-1.2, 0.8)]))
        painter.add(reflection, glass, flat=2, contour=False, outline=False)
        dome = ellipse(cx - radius * 0.46, cy - radius * 0.68, radius * 0.2, radius * 0.11, -30) & shell
        painter.add(dome, ramp, flat=5, contour=False, outline=False)
        painter.add(ellipse(cx - radius * 0.2, cy - radius * 0.82, radius * 0.07, radius * 0.07) & shell, ramp, flat=5, contour=False, outline=False)
    sill = visor - shift(visor, 0, -1)
    painter.add(sill, glass, flat=4 if radius >= 9 else 3, contour=False, outline=False)
    dot = ellipse(cx + radius * 0.62, cy - radius * 0.14, max(0.6, radius * 0.07), max(0.6, radius * 0.07))
    painter.add(dot & visor, glass, flat=5, contour=False, outline=False)

    mouth = polygon(frame.pts([(-0.46, 0.62), (0.46, 0.62), (0.36, 0.96), (-0.36, 0.96)]))
    painter.add((grow(mouth, 1) - mouth) & shell, ramp, flat=1, contour=False, outline=False)
    painter.add(mouth, P.WHITE, relief=max(1.5, radius * 0.2), gain=3.2, ceiling=4.6)
    if radius >= 5:
        for slot in (-0.2, -0.07, 0.07, 0.2):
            painter.add(
                polygon(frame.pts([(slot - 0.035, 0.68), (slot + 0.035, 0.68), (slot + 0.035, 0.92), (slot - 0.035, 0.92)])),
                P.WHITE, flat=1, contour=False, outline=False,
            )

    if kind == "green" and gem is not None:
        painter.add(ellipse(cx, cy - radius * 0.56, radius * 0.15, radius * 0.17), gem, relief=2.0, gain=4.0, ceiling=5.0)
    if kind == "yellow":
        for sign in (-1, 1):
            fang = [(sign * 0.34, 0.9), (sign * 0.18, 0.9), (sign * 0.26, 1.2)]
            painter.add(polygon(frame.pts(fang)), P.WHITE, relief=1.5, gain=3.0, ceiling=5.0)
    if kind == "black":
        for sign in (-1, 1):
            tusk = capsule(frame.pt(sign * 1.02, 0.5), frame.pt(sign * 1.3, 0.0), radius * 0.11, radius * 0.05)
            painter.add(tusk, P.WHITE, relief=1.5, gain=3.0, ceiling=5.0)
    return shell
