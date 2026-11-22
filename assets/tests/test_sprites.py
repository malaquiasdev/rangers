import itertools
import struct
import unittest
from pathlib import Path

from pixelart import hero_sprites, palette

SPRITES = Path(__file__).resolve().parent.parent / "sprites"
HEROES = ["red", "black", "pink", "blue", "yellow", "green"]
FILES = ["jason", "zack", "kimberly", "billy", "trini", "tommy", "zordon", "alpha-5", "rita", "lord-zedd", "putties"]


def outline_tones():
    return {value[0] for value in vars(palette).values() if isinstance(value, list) and len(value) == 6}


def png_size(path):
    header = path.read_bytes()[:24]
    return struct.unpack(">II", header[16:24])


class HeroSpriteTest(unittest.TestCase):
    def setUp(self):
        self.sprites = {key: hero_sprites.build_hero_sprite(key) for key in HEROES}

    def test_every_hero_sprite_is_48_by_64(self):
        for key, canvas in self.sprites.items():
            self.assertEqual((canvas.width, canvas.height), (48, 64), key)

    def test_the_corners_are_transparent(self):
        for key, canvas in self.sprites.items():
            for x, y in ((0, 0), (47, 0), (0, 63), (47, 63)):
                self.assertIsNone(canvas.get(x, y), key)

    def test_only_outline_pixels_touch_the_canvas_border(self):
        outlines = outline_tones()
        for key, canvas in self.sprites.items():
            edge = [canvas.get(x, 0) for x in range(48)] + [canvas.get(x, 63) for x in range(48)]
            edge += [canvas.get(0, y) for y in range(64)] + [canvas.get(47, y) for y in range(64)]
            self.assertTrue(all(pixel is None or pixel in outlines for pixel in edge), key)

    def test_the_feet_rest_on_the_same_baseline(self):
        bottoms = set()
        for canvas in self.sprites.values():
            rows = [y for y in range(64) if any(canvas.get(x, y) is not None for x in range(48))]
            bottoms.add(max(rows))
        self.assertEqual(bottoms, {63})

    def test_every_silhouette_is_different(self):
        masks = {key: [pixel is not None for pixel in canvas.pixels] for key, canvas in self.sprites.items()}
        for first, second in itertools.combinations(HEROES, 2):
            self.assertNotEqual(masks[first], masks[second], (first, second))

    def test_each_hero_wears_a_visor_a_white_chest_diamond_and_white_gloves(self):
        for key, canvas in self.sprites.items():
            colors = [pixel for pixel in canvas.pixels if pixel is not None]
            whites = [color for color in colors if min(color) > 225]
            self.assertGreater(len(whites), 40, key)
            visor = [color for color in colors if color[2] > color[0] + 30 and sum(color) < 260]
            self.assertGreater(len(visor), 12, key)


class ExportedSpriteTest(unittest.TestCase):
    def test_every_exported_sprite_is_192_by_256(self):
        for name in FILES:
            self.assertEqual(png_size(SPRITES / f"{name}.png"), (192, 256), name)


if __name__ == "__main__":
    unittest.main()
