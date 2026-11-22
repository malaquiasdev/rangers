# Assets

Pixel art for the Rangers repository: the banner and one sprite per agent. Everything is generated from text maps and palette ramps with Python 3 standard library only. There is no image editor step and no external tool.

## Regenerate

```sh
python3 assets/build.py
```

The command rewrites `assets/banner/banner.png` (1280x640) and the eleven files in `assets/sprites/` (192x256). Every image is drawn on a small native grid and scaled by a whole number with nearest neighbor, so no pixel is ever blurred.

| Image | Native grid | Export |
| --- | --- | --- |
| Banner, also the GitHub social preview | 320x160 | 4x, 1280x640 |
| Sprites | 48x64 | 4x, 192x256 |

## Where things live

| File | What it holds |
| --- | --- |
| `pixelart/sprite_maps.py` | The 46x62 map of each sprite. One character per pixel, `.` is empty. |
| `pixelart/hero_maps.py` | The six smaller full-body poses used in the banner line-up. |
| `pixelart/title_map.py` | The hand-drawn glyphs of the title. |
| `pixelart/palette.py` | The color ramps. Each ramp has six tones, from the dark outline tone to the specular tone. |
| `pixelart/shade.py` | `render_map`: shading, dithering, interior and exterior outlines, rim light. |
| `pixelart/sprites.py` | The material of every map character, the draw order and the `LEGEND`. |
| `pixelart/title.py` | The bevelled title with gradient, extrusion, drop shadow and sparkles. |
| `pixelart/scene.py` | The layered banner: sky, moon, mountains, city, stage, reflections. |

## Edit a sprite

1. Open `pixelart/sprite_maps.py` and change characters in the map of the sprite. Every row must have the same length. The `LEGEND` in `pixelart/sprites.py` says what each character means.
2. Run `python3 assets/build.py`. The build fails with the row number if a row has the wrong width or if a character has no material.
3. Open the PNG in `assets/sprites/` and check it at its real size too, since GitHub shows it at half the export size.

You only draw the shapes. `render_map` lights each region from the top left: it picks one of the six tones of the ramp from the distance to the edge, draws a colored outline darker than the fill, adds a rim light on the shadow side and draws a one-pixel line where a part sits in front of another. The part in front is the one that comes later in the draw order list.

To change a color, edit the ramp in `pixelart/palette.py` or the `Material` of the character in `pixelart/sprites.py`. `hue_ramp(base)` builds a ramp from one color: shadows drift toward blue and violet, highlights toward warm yellow.

## Dithering

`bayer` and `banded` in `pixelart/shade.py` turn a smooth value into a pixel-art gradient. The transition between two tones is a narrow checkerboard band, not a blur. The sky, the moon, the light beams, the glow under each hero and the reflections use them.

## Originality

All shapes, helmets, emblems and logos here are original. No design from any show, game or console is traced, recolored or reused.
