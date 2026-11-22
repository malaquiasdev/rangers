import struct
import unittest
from pathlib import Path

from pixelart import poster

BANNER = Path(__file__).resolve().parent.parent / "banner"
PANEL_X = [12 + index * 49 for index in range(6)]


def png_size(path):
    header = path.read_bytes()[:24]
    return struct.unpack(">II", header[16:24])


def dominant(canvas, x0, y0, x1, y1):
    counts = {}
    for y in range(y0, y1):
        for x in range(x0, x1):
            name = family(canvas.get(x, y))
            counts[name] = counts.get(name, 0) + 1
    counts.pop("other", None)
    return max(counts, key=counts.get)


def family(color):
    red, green, blue = color
    if max(color) < 120 and max(color) - min(color) < 50:
        return "black"
    if red > 180 and green > 150 and blue < 110:
        return "yellow"
    if red > 170 and green < 90 and blue < 110:
        return "red"
    if red > 200 and 60 < green < 130 and blue > 110:
        return "pink"
    if green > red + 40 and green > blue + 40:
        return "green"
    if blue > red + 70:
        return "blue"
    return "other"


class PosterTest(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.canvas = poster.build_poster()

    def test_the_poster_is_a_native_320_by_480_and_fully_painted(self):
        self.assertEqual((self.canvas.width, self.canvas.height), (320, 480))
        self.assertTrue(all(pixel is not None for pixel in self.canvas.pixels))

    def test_each_panel_shows_the_suit_color_of_its_hero_in_the_listed_order(self):
        expected = ["blue", "pink", "red", "green", "yellow", "black"]
        for index, name in enumerate(expected):
            color = self.canvas.get(PANEL_X[index] + 3, 330)
            self.assertEqual(family(color), name, (name, color))

    def test_the_title_band_is_gold_with_a_dark_red_extrusion(self):
        title = [self.canvas.get(x, y) for x in range(60, 260) for y in range(90, 130)]
        gold = [color for color in title if family(color) == "yellow"]
        self.assertGreater(len(gold), 800)

    def test_the_gem_glows_in_the_middle_above_the_title(self):
        cyan = [self.canvas.get(x, y) for x in range(145, 176) for y in range(30, 95)]
        self.assertGreater(len([color for color in cyan if color[2] > 180 and color[0] < 120]), 150)

    def test_each_hero_has_a_bright_rim_around_the_figure(self):
        rows = [self.canvas.get(x, 440) for x in range(0, 320)]
        whites = [color for color in rows if min(color) > 240]
        self.assertGreater(len(whites), 10)


class BannerTest(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.canvas = poster.build_banner()

    def test_the_banner_is_a_native_320_by_160_and_fully_painted(self):
        self.assertEqual((self.canvas.width, self.canvas.height), (320, 160))
        self.assertTrue(all(pixel is not None for pixel in self.canvas.pixels))

    def test_the_six_panels_and_heroes_keep_the_poster_order(self):
        expected = ["blue", "pink", "red", "green", "yellow", "black"]
        for index, name in enumerate(expected):
            left = 14 + index * 49
            self.assertEqual(dominant(self.canvas, left + 8, 74, left + 38, 84), name)


class ExportedSceneTest(unittest.TestCase):
    def test_the_exports_are_4x_of_the_native_grid(self):
        self.assertEqual(png_size(BANNER / "poster.png"), (1280, 1920))
        self.assertEqual(png_size(BANNER / "banner.png"), (1280, 640))


if __name__ == "__main__":
    unittest.main()
