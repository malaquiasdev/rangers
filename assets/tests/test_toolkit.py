import unittest

from pixelart import palette as P
from pixelart.canvas import Canvas
from pixelart.geometry import bounds, capsule, ellipse, polygon, shift, symmetric
from pixelart.painter import Painter, depth_map


class GeometryTest(unittest.TestCase):
    def test_a_rectangle_polygon_fills_exactly_its_pixels(self):
        pixels = polygon([(2, 3), (7, 3), (7, 6), (2, 6)])
        self.assertEqual(len(pixels), 5 * 3)
        self.assertEqual(bounds(pixels), (2, 3, 6, 5))

    def test_a_symmetric_polygon_mirrors_around_its_axis(self):
        pixels = polygon(symmetric([(0.0, -4), (3, -1), (2, 3), (0.0, 4)]))
        self.assertEqual({(-x - 1, y) for x, y in pixels}, pixels)

    def test_a_capsule_covers_both_end_points(self):
        pixels = capsule((4.5, 4.5), (20.5, 4.5), 2.5)
        self.assertIn((4, 4), pixels)
        self.assertIn((20, 4), pixels)
        self.assertNotIn((4, 12), pixels)

    def test_shifting_keeps_the_shape(self):
        disc = ellipse(10, 10, 4, 4)
        self.assertEqual(shift(shift(disc, 3, 2), -3, -2), disc)


class PainterTest(unittest.TestCase):
    def draw_disc(self):
        painter = Painter()
        painter.add(ellipse(20, 20, 10, 10), P.RED, relief=8.0)
        canvas = Canvas(40, 40)
        painter.draw(canvas)
        return canvas

    def test_depth_grows_toward_the_middle_of_a_shape(self):
        depth = depth_map(ellipse(20, 20, 10, 10))
        self.assertLess(depth[(10, 20)], depth[(15, 20)])
        self.assertLess(depth[(15, 20)], depth[(20, 20)])

    def test_light_comes_from_the_top_left(self):
        canvas = self.draw_disc()
        lit = sum(canvas.get(14, 14))
        shaded = sum(canvas.get(26, 26))
        self.assertGreater(lit, shaded)

    def test_a_dark_outline_surrounds_the_silhouette(self):
        canvas = self.draw_disc()
        self.assertEqual(canvas.get(9, 20), P.RED[0])
        self.assertIsNone(canvas.get(7, 20))

    def test_a_rim_adds_two_rings_outside_the_outline(self):
        painter = Painter()
        painter.add(ellipse(20, 20, 8, 8), P.BLUE, relief=6.0)
        canvas = Canvas(40, 40)
        painter.draw(canvas, rim=((255, 255, 255), (10, 200, 255)))
        self.assertEqual(canvas.get(11, 20), P.BLUE[0])
        self.assertEqual(canvas.get(10, 20), (255, 255, 255))
        self.assertEqual(canvas.get(9, 20), (10, 200, 255))

    def test_a_part_in_front_gets_no_outline_when_it_is_a_decal(self):
        painter = Painter()
        painter.add(ellipse(20, 20, 10, 10), P.GREEN, relief=8.0)
        painter.add(ellipse(20, 20, 3, 3), P.WHITE, flat=5, contour=False, outline=False)
        canvas = Canvas(40, 40)
        painter.draw(canvas)
        self.assertNotIn(P.GREEN[0], [canvas.get(x, 20) for x in range(14, 26)])


if __name__ == "__main__":
    unittest.main()
