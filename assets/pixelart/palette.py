import colorsys

from .canvas import rgb

SHADOW_HUE = 262
LIGHT_HUE = 48


def toward(hue, target, amount):
    delta = ((target - hue + 540) % 360) - 180
    return (hue + delta * amount) % 360


def hue_ramp(base, shadow_hue=SHADOW_HUE, light_hue=LIGHT_HUE, shift=0.34):
    red, green, blue = (channel / 255 for channel in rgb(base))
    hue, sat, val = colorsys.rgb_to_hsv(red, green, blue)
    hue *= 360
    steps = [
        (-1.0, 0.92, 0.30),
        (-0.62, 1.04, 0.52),
        (-0.30, 1.06, 0.76),
        (0.0, 1.0, 1.0),
        (0.38, 0.80, 1.12),
        (0.80, 0.34, 1.20),
    ]
    ramp = []
    for tilt, sat_scale, val_scale in steps:
        target = shadow_hue if tilt < 0 else light_hue
        tinted = toward(hue, target, abs(tilt) * shift)
        value = min(1.0, val * val_scale)
        saturation = min(1.0, sat * sat_scale)
        r, g, b = colorsys.hsv_to_rgb(tinted / 360, saturation, value)
        ramp.append((round(r * 255), round(g * 255), round(b * 255)))
    return ramp


RED = hue_ramp("#d92f3f")
BLACK = hue_ramp("#4a4a66", shift=0.25)
PINK = hue_ramp("#f0609e")
BLUE = hue_ramp("#2f7cf0")
YELLOW = hue_ramp("#f5c425", shadow_hue=300)
GREEN = hue_ramp("#2fbf5f")
GOLD = hue_ramp("#e8b230", shadow_hue=330)
STEEL = hue_ramp("#8a97b8", light_hue=190)
NAVY = hue_ramp("#27304f", shift=0.2)
CYAN = hue_ramp("#38d6e8", light_hue=120)
PURPLE = hue_ramp("#8a3fc8", shadow_hue=250)
MAGENTA = hue_ramp("#d63fb0")
CRIMSON = hue_ramp("#a8203a")
OBSIDIAN = hue_ramp("#2a2438", shift=0.2)
CLAY = hue_ramp("#a08468", shadow_hue=285)
SKIN = hue_ramp("#e8a878", shadow_hue=320)
CREAM = hue_ramp("#f0e4c8", shadow_hue=280)
VISOR = hue_ramp("#1b5f86", shadow_hue=270, shift=0.3)
HOLO = hue_ramp("#9be6f0", shadow_hue=250, light_hue=70, shift=0.3)
SILVER = hue_ramp("#dfe6f4", shadow_hue=255, light_hue=60, shift=0.28)
GLASS = [(4, 6, 18), (10, 16, 40), (22, 40, 84), (44, 92, 150), (120, 184, 224), (255, 255, 255)]
WHITE = [(52, 58, 108), (118, 130, 188), (196, 208, 238), (238, 243, 255), (250, 252, 255), (255, 255, 255)]
MIST = [(5, 14, 20), (10, 36, 40), (22, 76, 70), (54, 128, 100), (112, 186, 128), (190, 232, 160)]
MINT = [(10, 30, 38), (26, 62, 72), (60, 112, 118), (116, 168, 156), (180, 216, 184), (226, 246, 214)]
NIGHT = [(3, 4, 12), (7, 9, 24), (12, 16, 40), (20, 28, 62), (32, 46, 92), (54, 78, 140)]
