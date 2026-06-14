# SoftGO

Project for implementing OpenGL-like software rendering for 3D graphics. It supports:

- Texture loading (animated ones too via GIF, see [loader](./loader/loader.go))
- Models loading (normals too, but only triangulated, only .obj)
- Double-terminal pixels (uses `▀` and background color for rectengular pixels, thx for [stg](https://github.com/striter-no/stg))
- X11 input for mouse AND keyboard (supports X-wayland via X11 terms)

## How to run an example

Firstly you need to install `libXfixes-devel` package on your system, than:

```sh
go run ./example/main.go # Ctrl + C for exit
```

- WASD - for movement
- E/Q - for UP/DOWN movement

## How to run on Wayland

Use terminal, that supports X11 (for example - alacritty) and set `WAYLAND_DISPLAY` to null:

```sh
env WAYLAND_DISPLAY= alacritty

# and than run example
```
