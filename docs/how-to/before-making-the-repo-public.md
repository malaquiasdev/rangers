# How to get the repository ready before making it public

This checklist is for the day the repository stops being private. Do each item, then change the visibility.

- [ ] Replace the fan-style art with original art. The poster, the banner and the eleven sprites are drawn in the style of an existing franchise. That is fine while the repository is private, but not for a public one.
  1. Redraw `assets/banner/poster.png`, `assets/banner/banner.png` and every file in `assets/sprites/` with original heroes, helmets, emblem and title. Keep the sizes in the [assets README](../../assets/README.md#regenerate), so the README and the social preview still fit.
  2. Change the helmets in `assets/pixelart/helmets.py`, the poses in `assets/pixelart/figure.py` and the gem in `assets/pixelart/emblem.py`, or replace the files with your own art.
  3. Run `python3 assets/build.py`, then `make check`, and open the poster and the banner to see them.
