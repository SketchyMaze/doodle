# Doodle

# Milestones

As a rough idea of the milestones needed for this game to work:

## SDL Paint Program

* [x] Create a basic SDL window that you can click on to color pixels.
  * [x] Connect the pixels while the mouse is down to cover gaps.
* [x] Implement a "screenshot" button that translates the canvas to a PNG
  image on disk.
  * `F12` key to take a screenshot of your drawing.
  * It reproduces a PNG image using its in-memory knowledge of the pixels you
    have drawn, *not* by reading the SDL canvas. This will be important for
    making the custom level format later.
  * The PNG I draw looks slightly different to what you see on the SDL canvas;
    maybe difference between `Renderer.DrawLine()` and my own algorithm or
    the anti-aliasing.
* [ ] Create a custom map file format (protobufs maybe) and "screenshot" the
  canvas into this custom file format.
* [ ] Make the program able to read this file format and reproduce the same
  pixels on the canvas.

## Platformer

* [ ] Start implementing a platformer that uses the custom map format for its
  rendering and collision detection.
* [ ] ???

# Building

Fedora dependencies:

```bash
$ sudo dnf install SDL2-devel SDL2_ttf-devel
```
