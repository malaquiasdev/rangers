import math


def lattice(seed, ix, iy):
    value = (ix * 374761393 + iy * 668265263 + seed * 2147483647) & 0xFFFFFFFF
    value = ((value ^ (value >> 13)) * 1274126177) & 0xFFFFFFFF
    value ^= value >> 16
    return (value & 0xFFFF) / 65535.0


def smooth(t):
    return t * t * (3 - 2 * t)


def value_noise(seed, x, y):
    ix, iy = math.floor(x), math.floor(y)
    fx, fy = smooth(x - ix), smooth(y - iy)
    top = lattice(seed, ix, iy) * (1 - fx) + lattice(seed, ix + 1, iy) * fx
    bottom = lattice(seed, ix, iy + 1) * (1 - fx) + lattice(seed, ix + 1, iy + 1) * fx
    return top * (1 - fy) + bottom * fy


def fbm(seed, x, y, octaves=4):
    total, amplitude, frequency, weight = 0.0, 1.0, 1.0, 0.0
    for octave in range(octaves):
        total += value_noise(seed + octave * 17, x * frequency, y * frequency) * amplitude
        weight += amplitude
        amplitude *= 0.5
        frequency *= 2.0
    return total / weight
