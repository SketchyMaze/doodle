# Changes

## v0.0.1-alpha - June 17 2018

* Add a debug overlay that shows FPS, coordinates, and useful info.
* Add FPS throttling to target 60 frames per second.
* Add `F12` for Screenshot key which saves the in-memory representation of
  the pixels you've drawn to disk as a PNG image.
* Smoothly connect dots between periods where the mouse button was held down
  but was moving too fast.

## v0.0.0-alpha

* Basic SDL canvas that draws pixels when you click and/or drag.
* The lines drawn aren't smooth, because the mouse cursor moves too fast.

### Screenshot Feature

Pressing `F12` takes a screenshot and saves it on disk as a PNG.

It does **NOT** read the SDL canvas data for this, though. It uses an
internal representation of the pixels you've been drawing, and writes that
to the PNG. This is important because that same pixel data will be used for
the custom level format.
