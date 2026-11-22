# Assets

Pixel art for the Rangers repository: the cover poster, the wide banner and one sprite per agent. The poster and the banner show the backdrop, the panels and the title; the heroes live only in the panels and in the sprites. Everything is drawn by code from shapes, lighting rules and palette ramps, with Python 3 standard library only. There is no image editor step and no external tool.

The art is fan-style and must be replaced by original art before the repository goes public. The checklist is in [before making the repository public](../docs/how-to/before-making-the-repo-public.md).

## Regenerate

```sh
python3 assets/build.py
```

The command rewrites the poster, the banner and the eleven sprites. Every image is drawn on a small native grid and scaled by a whole number with nearest neighbor, so no pixel is ever blurred.

| Image | Native grid | Export |
| --- | --- | --- |
| `banner/poster.png`, the cover at the top of the README | 320x480 | 4x, 1280x1920 |
| `banner/banner.png`, also the GitHub social preview | 320x160 | 4x, 1280x640 |
| `sprites/*.png`, one per agent | 48x64 | 4x, 192x256 |

Run the tests with `make check`, or only these:

```sh
python3 -I -m unittest discover -s assets/tests -t assets
```

## Where things live

| File | What it holds |
| --- | --- |
| `pixelart/geometry.py` | Polygons, ellipses, capsules and strokes turned into pixel sets. |
| `pixelart/painter.py` | Lights each part from the top left, with ordered dithering, contours, the outer outline and the two-ring rim. |
| `pixelart/palette.py` | The color ramps. Each ramp has six tones, from the outline tone to the specular tone. |
| `pixelart/helmets.py` | One helmet per hero, with its own crest, visor and cheek shape. Used in the poses and in the panels. |
| `pixelart/figure.py` | The body rig and the pose of each hero, as joint angles. |
| `pixelart/hero_sprites.py` | The six hero sprites, fitted to 48x64. |
| `pixelart/sprite_maps.py`, `pixelart/sprites.py`, `pixelart/shade.py` | The text maps and materials of the other five sprites. |
| `pixelart/backdrop.py`, `pixelart/noise.py` | The dark command center: mist, rings, tubes, walls, floor. |
| `pixelart/panels.py` | The six vertical close-up panels. |
| `pixelart/lettering.py`, `pixelart/emblem.py` | The bevelled title with its ribbon, and the gem crest. |
| `pixelart/stage.py` | The stage floor: glowing rings, panel reflections, colored glow pools, light beams and sparkles. |
| `pixelart/poster.py` | The poster and the banner, each as its own composition of the backdrop, the panels, the stage and the title. |

## Edit a hero

1. Open `pixelart/figure.py` and change the angles in `POSES`. An angle is measured from straight down, clockwise, in degrees. Each leg is thigh, shin and foot, and each arm is upper arm and forearm.
2. To change a helmet, edit the visor, crest or cheek points in `pixelart/helmets.py`. They are written in head units, so the same shape works at sprite size and in a panel.
3. Run `python3 assets/build.py` and open the PNGs. Look at the sprites at real size too, since GitHub shows them at half the export size.

To change a color, edit the ramp in `pixelart/palette.py`. `hue_ramp(base)` builds a ramp from one color: shadows drift toward blue and violet, highlights toward warm yellow.

## Dithering

`bayer` in `pixelart/painter.py` turns a smooth value into a pixel-art gradient. The transition between two tones is a narrow checkerboard band, not a blur. The mist, the light beams, the glows and the floor use it too.
