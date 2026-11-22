from . import palette as P
from .shade import Material, banded, render_map
from .sprite_maps import SPRITE_MAPS

RIM = (150, 200, 255)
WHITE = [(255, 255, 255)] * 6


def under(ramp):
    return [ramp[0], ramp[0], ramp[1], ramp[2], ramp[3], ramp[4]]


def shine(flat=5):
    return Material(P.CYAN, flat=flat, outline=False)


def common():
    return {
        "t": Material(P.GOLD),
        "e": shine(),
        "+": shine(),
        "*": Material(WHITE, flat=3, outline=False),
        "G": Material(P.OBSIDIAN),
        "g": Material(P.STEEL),
        "n": Material(P.NAVY),
        "v": Material(P.VISOR, lit=0.0, shade=0.0),
    }


RANGER_ORDER = ["u", "c", "n", "U", "r", "g", "G", "S", "m", "k", "q", "p", "h", "v", "t", "e", "*", "+"]


def ranger(main):
    materials = common()
    materials.update(
        {
            "u": Material(under(main)),
            "U": Material(under(main)),
            "c": Material(main),
            "r": Material(main),
            "p": Material(main),
            "h": Material(main),
            "S": Material(P.GOLD),
            "m": Material(P.STEEL),
            "k": Material(P.CRIMSON),
            "q": Material(P.CREAM),
        }
    )
    return materials


RANGER_RAMPS = {
    "jason": P.RED,
    "zack": P.BLACK,
    "kimberly": P.PINK,
    "billy": P.BLUE,
    "trini": P.YELLOW,
    "tommy": P.GREEN,
}

CAST_ORDER = ["z", "y", "w", "u", "c", "n", "U", "r", "g", "G", "T", "d", "F", "f", "k", "q", "p", "h", "v", "x", "t", "E", "e", "*", "+"]


def zordon():
    materials = common()
    materials.update(
        {
            "z": Material(P.NAVY, flat=1, outline=False),
            "E": Material(P.GOLD, flat=5, outline=False),
            "y": Material(P.CYAN, lit=0.5, shade=0.5),
            "w": Material(P.CYAN, flat=4, outline=False),
            "T": Material(P.GOLD),
            "d": Material(P.SILVER),
            "F": Material(P.SILVER, flat=1),
            "f": Material(P.HOLO),
        }
    )
    return materials


def alpha():
    materials = common()
    materials.update(
        {
            "e": Material(P.YELLOW, flat=5, outline=False),
            "c": Material(P.STEEL),
            "U": Material(P.STEEL),
            "u": Material(P.NAVY),
            "v": Material(P.NAVY, lit=0.5, shade=0.5),
            "g": Material(P.GOLD),
            "G": Material(P.NAVY),
        }
    )
    return materials


def rita():
    materials = common()
    materials.update(
        {
            "c": Material(P.PURPLE),
            "U": Material(P.PURPLE),
            "r": Material(P.MAGENTA),
            "p": Material(P.MAGENTA),
            "h": Material(P.MAGENTA),
            "f": Material(P.SKIN),
        }
    )
    return materials


def zedd():
    materials = common()
    materials.update(
        {
            "k": Material(P.OBSIDIAN),
            "q": Material(P.CRIMSON),
            "c": Material(P.CRIMSON),
            "h": Material(P.OBSIDIAN),
            "U": Material(P.OBSIDIAN),
            "u": Material(P.NAVY),
            "r": Material(P.CRIMSON),
            "p": Material(P.OBSIDIAN),
            "e": Material(P.YELLOW, flat=5, outline=False),
            "g": Material(P.CRIMSON),
            "v": Material(P.OBSIDIAN, lit=0.0, shade=0.0),
        }
    )
    return materials


def putties():
    materials = common()
    materials.update(
        {
            "c": Material(P.CLAY),
            "U": Material(P.CLAY),
            "u": Material(P.CLAY),
            "x": Material(P.CLAY, flat=1),
        }
    )
    return materials


GLOW_STOPS = [(10, 16, 52), (16, 30, 84), (22, 52, 120), (34, 92, 156), (60, 150, 196)]


def tube_glow(canvas, rows, center, radius):
    for y, row in enumerate(rows):
        for x, char in enumerate(row):
            if char != "z":
                continue
            distance = ((x - center[0]) ** 2 + ((y - center[1]) * 0.8) ** 2) ** 0.5
            level = (1 - min(1.0, distance / radius)) * (len(GLOW_STOPS) - 1)
            canvas.put(x + 1, y + 1, GLOW_STOPS[min(len(GLOW_STOPS) - 1, banded(level + 0.3, x, y, 3.0))])


def build_sprite(name):
    rows = SPRITE_MAPS[name]
    if name in RANGER_RAMPS:
        return render_map(rows, ranger(RANGER_RAMPS[name]), RANGER_ORDER, rim=RIM)
    makers = {"zordon": zordon, "alpha-5": alpha, "rita": rita, "lord-zedd": zedd, "putties": putties}
    canvas = render_map(rows, makers[name](), CAST_ORDER, rim=RIM)
    if name == "zordon":
        tube_glow(canvas, rows, (23, 26), 24)
    return canvas


def build_all():
    return {name: build_sprite(name) for name in SPRITE_MAPS}


LEGEND = {
    ".": "transparent",
    "u": "undersuit, a darker step of the suit ramp",
    "U": "undersuit of an arm, drawn over the chest",
    "c": "chest plate",
    "n": "belt",
    "r": "bracers and greaves",
    "p": "pauldrons, knee plates and crests",
    "h": "helmet shell",
    "v": "glossy visor",
    "g": "gloves",
    "G": "boots",
    "t": "gold trim",
    "e": "emissive glow",
    "*": "specular highlight",
    "+": "cool glint",
    "S": "shield rim",
    "m": "shield face",
    "k": "book cover",
    "q": "book pages",
}
