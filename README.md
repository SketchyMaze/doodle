# Doodle

# Milestones

As a rough idea of the milestones needed for this game to work:

## SDL Paint Program

* [x] Create a basic SDL window that you can click on to color pixels.
  * [ ] Connect the pixels while the mouse is down to cover gaps.
* [ ] Implement a "screenshot" button that translates the canvas to a PNG
  image on disk.
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
