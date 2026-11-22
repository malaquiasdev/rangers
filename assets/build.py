from pathlib import Path

from pixelart import hero_sprites, sprites
from pixelart.sprite_maps import SPRITE_MAPS

HERE = Path(__file__).resolve().parent
SCALE = 4
SPRITE_SIZE = (48, 64)
HERO_FILES = {
    "red": "jason",
    "black": "zack",
    "pink": "kimberly",
    "blue": "billy",
    "yellow": "trini",
    "green": "tommy",
}


def save(canvas, size, relative):
    if (canvas.width, canvas.height) != size:
        raise ValueError(f"{relative} is {canvas.width}x{canvas.height}, expected {size[0]}x{size[1]}")
    canvas.save(HERE / relative, SCALE)
    print(f"{relative} {canvas.width * SCALE}x{canvas.height * SCALE}")


def build_sprites():
    for key, name in HERO_FILES.items():
        save(hero_sprites.build_hero_sprite(key), SPRITE_SIZE, f"sprites/{name}.png")
    for name in SPRITE_MAPS:
        save(sprites.build_sprite(name), SPRITE_SIZE, f"sprites/{name}.png")


def main():
    build_sprites()


if __name__ == "__main__":
    main()
