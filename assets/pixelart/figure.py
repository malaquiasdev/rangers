import math

from . import palette as P
from .geometry import capsule, ellipse, polygon
from .helmets import paint_helmet

HEAD = 7.3
TORSO = 16.6
SHOULDER = 8.4
WAIST = 4.6
THIGH = 11.6
SHIN = 11.4
UPPER = 9.0
FORE = 7.8


def heading(angle):
    radians = math.radians(angle)
    return math.sin(radians), math.cos(radians)


def along(point, angle, length):
    dx, dy = heading(angle)
    return point[0] + dx * length, point[1] + dy * length


def between(a, b, amount):
    return a[0] + (b[0] - a[0]) * amount, a[1] + (b[1] - a[1]) * amount


POSES = {
    "red": dict(lean=0, legs=((-28, -4, -68), (28, 4, 68)), arms=((-70, 172), (70, -172)), spread=3.4),
    "blue": dict(lean=10, legs=((-40, -6, -70), (14, 2, 66)), arms=((-60, 158), (40, 100)), spread=3.4),
    "pink": dict(lean=-6, legs=((-8, 0, -60), (78, 30, 115)), arms=((-104, -168), (50, 110)), spread=3.4),
    "green": dict(lean=-4, legs=((-26, -2, -68), (32, 6, 72)), arms=((-50, 160), (64, 150)), spread=3.4),
    "yellow": dict(lean=-10, legs=((-32, -8, -66), (20, 4, 66)), arms=((-30, 130), (112, 160)), spread=3.4),
    "black": dict(lean=8, legs=((-30, -4, -68), (34, 16, 70)), arms=((-70, 172), (60, 130)), spread=3.4),
}


def skeleton(key, scale):
    pose = POSES[key]
    lean = pose["lean"]
    up = (math.sin(math.radians(lean)), -math.cos(math.radians(lean)))
    side = (math.cos(math.radians(lean)), math.sin(math.radians(lean)))
    hip = (0.0, 0.0)

    def on_torso(height, offset=0.0):
        return (
            hip[0] + up[0] * height * scale + side[0] * offset * scale,
            hip[1] + up[1] * height * scale + side[1] * offset * scale,
        )

    joints = {"up": up, "side": side, "hip": hip, "on_torso": on_torso}
    joints["neck"] = on_torso(TORSO + 1.0)
    joints["shoulders"] = (on_torso(TORSO - 1.0, -SHOULDER), on_torso(TORSO - 1.0, SHOULDER))
    joints["waist"] = (on_torso(4.0, -WAIST), on_torso(4.0, WAIST))
    joints["hips"] = (on_torso(0.0, -pose["spread"]), on_torso(0.0, pose["spread"]))
    legs = []
    for index, (thigh, shin, foot) in enumerate(pose["legs"]):
        root = joints["hips"][index]
        knee = along(root, thigh, THIGH * scale)
        ankle = along(knee, shin, SHIN * scale)
        toe = along(ankle, foot, 5.0 * scale)
        legs.append((root, knee, ankle, toe))
    joints["legs"] = legs
    arms = []
    for index, (upper, fore) in enumerate(pose["arms"]):
        root = joints["shoulders"][index]
        root = (root[0] + (1.5 if index == 0 else -1.5) * scale, root[1] + 1.0 * scale)
        elbow = along(root, upper, UPPER * scale)
        wrist = along(elbow, fore, FORE * scale)
        arms.append((root, elbow, wrist))
    joints["arms"] = arms
    joints["head"] = on_torso(TORSO + 1.0 + HEAD * 0.78)
    return joints


def glint(painter, ramp, x, y, scale):
    painter.add(ellipse(x, y, max(0.9, 0.7 * scale), max(0.7, 0.5 * scale), -35), ramp, flat=5, contour=False, outline=False)


def paint_leg(painter, leg, ramp, scale, back):
    root, knee, ankle, toe = leg
    painter.add(capsule(root, knee, 4.0 * scale, 3.3 * scale), ramp, relief=3.2 * scale, gain=4.2)
    painter.add(capsule(knee, ankle, 3.3 * scale, 2.7 * scale), ramp, relief=2.8 * scale, gain=4.2)
    painter.add(ellipse(knee[0], knee[1], 3.4 * scale, 3.2 * scale), ramp, relief=3.0 * scale, gain=4.6, ceiling=4.8)
    glint(painter, ramp, knee[0] - 1.0 * scale, knee[1] - 1.1 * scale, scale)
    cuff = between(knee, ankle, 0.62)
    painter.add(capsule(cuff, ankle, 2.9 * scale, 2.8 * scale), P.WHITE, relief=2.8 * scale, gain=3.4, ceiling=4.6)
    sole_end = (toe[0], toe[1] + 0.4 * scale)
    painter.add(capsule(ankle, sole_end, 2.5 * scale, 2.1 * scale), P.WHITE, relief=2.4 * scale, gain=3.4, ceiling=4.6)
    painter.add(capsule((ankle[0], ankle[1] + 1.8 * scale), (sole_end[0], sole_end[1] + 1.4 * scale), 0.8 * scale),
                P.NAVY, flat=1, contour=False)


def paint_arm(painter, arm, ramp, scale):
    root, elbow, wrist = arm
    painter.add(capsule(root, elbow, 3.0 * scale, 2.6 * scale), ramp, relief=2.6 * scale, gain=4.2)
    painter.add(capsule(elbow, wrist, 2.6 * scale, 2.3 * scale), ramp, relief=2.4 * scale, gain=4.2)
    cuff = between(elbow, wrist, 0.62)
    painter.add(capsule(cuff, wrist, 2.4 * scale, 2.6 * scale), P.WHITE, relief=2.4 * scale, gain=3.4, ceiling=4.6)
    direction = (wrist[0] - elbow[0], wrist[1] - elbow[1])
    length = math.hypot(*direction) or 1.0
    fist = (wrist[0] + direction[0] / length * 1.8 * scale, wrist[1] + direction[1] / length * 1.8 * scale)
    painter.add(ellipse(fist[0], fist[1], 2.6 * scale, 2.5 * scale), P.WHITE, relief=2.6 * scale, gain=3.6, ceiling=4.8)
    ux, uy = direction[0] / length, direction[1] / length
    for offset in (-0.9, 0.9):
        start = (fist[0] - uy * offset * scale, fist[1] + ux * offset * scale)
        end = (start[0] + ux * 1.9 * scale, start[1] + uy * 1.9 * scale)
        painter.add(capsule(start, end, 0.35 * scale), P.WHITE, flat=1, contour=False, outline=False)
    painter.add(ellipse(root[0], root[1], 3.6 * scale, 3.4 * scale), ramp, relief=3.4 * scale, gain=4.8, ceiling=4.8)
    glint(painter, ramp, root[0] - 1.1 * scale, root[1] - 1.2 * scale, scale)


def paint_hero(painter, key, ramp, origin, scale=1.0, gem=P.MAGENTA, detail=1.0):
    joints = skeleton(key, scale)
    ox, oy = origin

    def moved(point):
        return (point[0] + ox, point[1] + oy)

    legs = [[moved(p) for p in leg] for leg in joints["legs"]]
    arms = [[moved(p) for p in arm] for arm in joints["arms"]]
    on_torso = joints["on_torso"]

    def torso_point(height, offset=0.0):
        return moved(on_torso(height, offset))

    for leg in legs:
        paint_leg(painter, leg, ramp, scale, True)

    pelvis = polygon([torso_point(0.0, -5.2), torso_point(0.0, 5.2), torso_point(5.0, 5.4), torso_point(5.0, -5.4)])
    pelvis |= ellipse(*torso_point(0.2), 5.6 * scale, 3.0 * scale)
    painter.add(pelvis, ramp, relief=3.0 * scale, gain=4.0)

    chest = polygon([
        torso_point(TORSO - 0.4, -SHOULDER - 0.4),
        torso_point(TORSO + 0.8, -3.4),
        torso_point(TORSO + 0.8, 3.4),
        torso_point(TORSO - 0.4, SHOULDER + 0.4),
        torso_point(TORSO - 3.4, SHOULDER - 1.2),
        torso_point(9.5, 6.0),
        torso_point(3.5, WAIST),
        torso_point(3.5, -WAIST),
        torso_point(9.5, -6.0),
        torso_point(TORSO - 3.4, -SHOULDER + 1.2),
    ])
    painter.add(chest, ramp, relief=5.0 * scale, gain=4.6, strength=3.0, ceiling=4.6)

    diamond = polygon([
        torso_point(15.4, 0.0), torso_point(11.0, 5.8), torso_point(5.6, 0.0), torso_point(11.0, -5.8),
    ])
    for sign in (-1, 1):
        strap = polygon([torso_point(14.0, sign * 1.0), torso_point(TORSO + 0.4, sign * 5.2), torso_point(TORSO - 1.8, sign * 7.0), torso_point(12.0, sign * 3.4)])
        painter.add(strap, P.WHITE, relief=1.6 * scale, gain=3.2, ceiling=4.6)
    painter.add(diamond, P.WHITE, relief=3.2 * scale, gain=4.0, ceiling=4.8)
    inner = polygon([
        torso_point(13.4, 0.0), torso_point(11.0, 3.0), torso_point(8.0, 0.0), torso_point(11.0, -3.0),
    ])
    painter.add(inner, ramp, relief=2.0 * scale, gain=3.0, contour=False, outline=False, ceiling=3.6)

    belt = polygon([torso_point(5.8, -WAIST - 0.6), torso_point(5.8, WAIST + 0.6), torso_point(2.8, WAIST + 0.8), torso_point(2.8, -WAIST - 0.8)])
    painter.add(belt, P.WHITE, relief=1.6 * scale, gain=3.2, ceiling=4.6)
    buckle = polygon([torso_point(5.6, -1.7), torso_point(5.6, 1.7), torso_point(3.0, 1.7), torso_point(3.0, -1.7)])
    painter.add(buckle, P.GOLD, relief=1.5 * scale, gain=3.6, ceiling=5.0)

    collar = ellipse(*torso_point(TORSO + 1.2), 4.8 * scale, 2.4 * scale)
    painter.add(collar, P.WHITE, relief=2.2 * scale, gain=3.4, ceiling=4.6)

    for arm in arms:
        paint_arm(painter, arm, ramp, scale)

    head = moved(joints["head"])
    paint_helmet(painter, KEYS[key], head[0], head[1], HEAD * scale, ramp, gem=gem)
    return {"head": head, "arms": arms, "legs": legs}


KEYS = {name: name for name in POSES}
