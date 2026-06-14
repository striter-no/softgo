# SoftGO

Project for implementing OpenGL-like software rendering for 3D graphics. It supports:

- Texture loading (animated ones too via GIF, see [loader](./loader/loader.go))
- Models loading (normals too, but only triangulated, only .obj)
- Double-terminal pixels (uses `▀` and background color for rectengular pixels, thx for [stg](https://github.com/striter-no/stg))
- X11 input for mouse AND keyboard (supports X-wayland via X11 terms)

## Requirements

SoftGO requires a GNU/Linux system with X11 support. Wayland sessions are also usable if XWayland-compatible terminals are employed.

You need Go 1.20 or newer to build and run the project.

The project uses a small set of Go dependencies for math, 3D utilities, terminal handling, and system/X11 interaction (including mathgl, go3d, stg, and a few golang.org/x/* packages).

A terminal emulator with proper X11 integration is required for correct input handling and rendering. Alacritty, Konsole, Kitty (X11 mode), or similar terminals are recommended.

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

## Project structure

The codebase is split into several modules:

- `api/` - core engine interfaces
  - camera system
  - shaders (vertex/fragment)
  - input handling (keyboard, mouse)
  - FPS counter
  - OBJ loader implementation
- `render/` - rasterization pipeline
  - triangle processing
  - clipping
  - FXAA
  - texture sampling
- `loader/` - asset loading layer
- `main/` - application entry point
- `assets/` - example models and textures

## License
Project is licensed under GNU GPL v3.0. Check the `LICENSE` file in the root directory