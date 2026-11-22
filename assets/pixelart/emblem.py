from . import palette as P
from .geometry import Frame, ellipse, polygon

FACETS = [
    ([(0.0, -1.0), (0.46, -0.52), (0.0, -0.12)], 4.6),
    ([(0.0, -1.0), (-0.46, -0.52), (0.0, -0.12)], 3.4),
    ([(0.46, -0.52), (0.5, 0.12), (0.0, -0.12)], 3.0),
    ([(-0.46, -0.52), (-0.5, 0.12), (0.0, -0.12)], 4.0),
    ([(0.5, 0.12), (0.0, 1.0), (0.0, -0.12)], 2.0),
    ([(-0.5, 0.12), (0.0, 1.0), (0.0, -0.12)], 2.8),
]


def paint_gem(painter, cx, cy, height, gem=P.CYAN, metal=P.SILVER):
    unit = height / 2
    frame = Frame(cx, cy, unit)
    wing = [(0.5, -0.1), (1.25, -0.62), (1.7, -0.36), (1.36, 0.12), (0.78, 0.34), (0.5, 0.3)]
    painter.add(polygon(frame.pts(wing)), metal, relief=3.0, gain=4.0, ceiling=5.0)
    painter.add(polygon(frame.pts([(-x, y) for x, y in wing])), metal, relief=3.0, gain=4.0, ceiling=5.0)
    claw = [(0.0, -1.14), (0.56, -0.58), (0.62, 0.18), (0.0, 1.14), (-0.62, 0.18), (-0.56, -0.58)]
    painter.add(polygon(frame.pts(claw)), metal, relief=3.6, gain=4.2, ceiling=5.0)
    for points, tone in FACETS:
        inset = [(x * 0.86, y * 0.86) for x, y in points]
        painter.add(polygon(frame.pts(inset)), gem, flat=tone, contour=False, outline=False)
    glint = polygon(frame.pts([(-0.2, -0.66), (-0.1, -0.7), (0.04, -0.32), (-0.08, -0.28)]))
    painter.add(glint, P.WHITE, flat=5, contour=False, outline=False)
    painter.add(ellipse(*frame.pt(0.18, -0.46), 0.8, 0.8), P.WHITE, flat=5, contour=False, outline=False)
