from pathlib import Path

from pixelart import scene, sprites
from pixelart.sprite_maps import SPRITE_MAPS

HERE = Path(__file__).resolve().parent
SPRITE_SCALE = 4
BANNER_SCALE = 4
SPRITE_SIZE = (48, 64)
BANNER_SIZE = (320, 160)


def build_sprites():
    for name in SPRITE_MAPS:
        canvas = sprites.build_sprite(name)
        if (canvas.width, canvas.height) != SPRITE_SIZE:
            raise ValueError(f"{name} is {canvas.width}x{canvas.height}, expected {SPRITE_SIZE[0]}x{SPRITE_SIZE[1]}")
        canvas.save(HERE / "sprites" / f"{name}.png", SPRITE_SCALE)
        print(f"sprites/{name}.png {canvas.width * SPRITE_SCALE}x{canvas.height * SPRITE_SCALE}")


def build_banner():
    canvas = scene.build_banner()
    if (canvas.width, canvas.height) != BANNER_SIZE:
        raise ValueError(f"banner is {canvas.width}x{canvas.height}, expected {BANNER_SIZE[0]}x{BANNER_SIZE[1]}")
    canvas.save(HERE / "banner" / "banner.png", BANNER_SCALE)
    print(f"banner/banner.png {canvas.width * BANNER_SCALE}x{canvas.height * BANNER_SCALE}")


def main():
    build_sprites()
    build_banner()


if __name__ == "__main__":
    main()
