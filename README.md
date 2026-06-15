# Soft Engine

Simple 3D Engine that utilises SoftGO for terminal 3D rendering and adds user-friendly API for creating scenes with lighting, multiple objects and textures

## Usage

Firstly install `libXfixes-devel` for HID on X11, than:

```sh
go run ./examples/main.go
```

## How to run on Wayland

```sh
env WAYLAND_DISPLAY= alacritty # or any other terminal that supports X11

# run example in new window
```
