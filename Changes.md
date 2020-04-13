# Changes

## v0.1.0-alpha

New doodads:

* Start Flag: drag this into your level to set where the player character will
  spawn. There should only be one per level.
* Crumbly Floor: a rocky floor that breaks and falls away after a couple
  seconds when the player (or other mobile doodad) walks onto it.
* State Blocks: blue and orange blocks that toggle between solid and passable
  when the corresponding ON/OFF button is touched.

New features:

* An inventory overlay now appears in Play Mode when the player character picks
  up one of the colored keys.
* While editing a level, you can click the new "Options" button in the top menu
  to open the level settings window (like the one you see when creating a new
  level): to change the wallpaper image or the page type.

Other changes:

* Added better platforming physics to the player character: acceleration and
  friction when walking.
* The colored Locked Door doodads have been re-designed to be shown in a
  side-view perspective and have an open and closed state in either direction.
* Tooltips added to various buttons in the Editor to show names of doodads and
  functions of various buttons.

## v0.0.10-alpha

New features:

* Added the **Eraser Tool** and support for **brush sizes**. Now you can clean
  up your mistakes and draw shapes with thicker lines!
* Added the **Ellipse Tool** for drawing elliptical shapes.
* Added a third example to the game's built-in levels.

Bug fixes:

* The Undo command now restores the original color of a pixel instead of just
  deleting it. Only works for thin lines so far.
* Improved collision detection algorithm to prevent players from clipping
  through a solid doodad, regardless of speed. This change is invisible this
  build, but opens the door to improvements in the 2D platforming physics and
  making the player character move and fall faster.
* Fix mobile non-player doodads from sometimes being able to clip through a
  solid doodad. For example, a Red Azulian could sometimes walk through a locked
  door without interacting with it.
* Sometimes hitting Undo would leave a broken "chunk" in your level, if the
  Undo operation deleted all pixels in that chunk. The broken chunk would show
  as a solid black square (non-solid) in the level. This has been fixed: empty
  chunks are now culled when the last pixel is deleted and existing level files
  will be repaired on next save.

## v0.0.9-alpha

First alpha release.
