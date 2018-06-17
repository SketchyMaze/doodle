# Doodle

Doodle is a drawing-based maze game written in Go.

# Features

(Eventually), the high-level, user-facing features for the game are:

* **Draw your own levels** freehand and then play them like a 2D platformer
  game.
* In **Adventure Mode** you can play through a series of official example
  levels that ship with the game.
* In **Edit Mode** you can draw a map freehand-style and lay down physical
  geometry, and mark which lines are solid or which ones behave like fire.
* Drag and drop **Doodads** like buttons, doors and keys into your level and
  link them together so that buttons open doors and levers activate devices.
* In **Play Mode** you can play your level as a 2D platformer game where you
  collect keys, watch out for enemies, and solve puzzles to get to the exit.
* Easily **Share** your custom maps with friends.

## Mod-friendly

* Users can create **Custom Doodads** to extend the game with a scripting
  language like JavaScript. The sprites and animations are edited in-game
  in Edit Mode, but the scripting is done in your text editor.
* In **Edit Mode** you can drag custom doodads into your maps.
* To **Share** your maps, you can choose to **bundle** the custom
  doodads inside your map file itself, so that other players can play
  the map without needing to install the doodads separately.
* If you receive a map with custom doodads, you can **install** the doodads
  into your copy of the game and use them in your own maps.

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
* [x] Create a custom map file format (protobufs maybe) and "screenshot" the
  canvas into this custom file format.
* [x] Make the program able to read this file format and reproduce the same
  pixels on the canvas.

## Platformer

* [ ] Inflate the pixel history from the map file into a full lookup grid
  of `(X,Y)` coordinates. This will be useful for collision detection.
* [ ] Create a dummy player character sprite, probably just a
  `render.Circle()`. In **Play Mode** run collision checks and gravity on
  the player sprite.
* [ ] Get basic movement and collision working. With a cleanup this can
  make a workable **ALPHA RELEASE**
* [ ] Wrap a Qt GUI around the SDL window to make the Edit Mode easier to
  work with, with toolbars to select brushes and doodads and junk.
* [ ] Work on support for solid vs. transparent, fire, etc. geometry.
* [ ] ???

# Building

Fedora dependencies:

```bash
$ sudo dnf install SDL2-devel SDL2_ttf-devel
```
