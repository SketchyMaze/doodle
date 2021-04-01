# Changes

## v0.5.0-alpha (Mar 31 2021)

Project: Doodle is renamed to Sketchy Maze in this release.

New Features:

* **New Tutorial Levels:** the bundled levels demonstrate the built-in doodads
  and how they interact and shows off several game features.
* **Level Editor:** you can now set the Title and Author of the level you're
  editing by using the Level->Page Settings window.
* The **Inventory HUD** in Play Mode now shows a small number indicator for items
  which have quantity, such as the Small Key. Colored Keys do not have quantity
  and don't show a number: those are permanent collectibles.
* **Fire Pixels:** when the player character dies by touching a "Fire" pixel
  during gameplay, the death message uses the **name** of the color instead
  of calling it "fire." For example, if you name a color "spikes" and give
  it the Fire attribute, it will say "Watch out for spikes!" if the player
  dies by touching it.
* New cheat code: `give all keys` gives all four colored keys and 99x Small Keys
  to the player character. The `drop all items` cheat clears your inventory.

New Doodads:

* **Warp Doors** allow the player character to fast travel to another location
  on the map. Drag two Warp Doors into your level and use the Link Tool to
  connect them together. Doors without an exit link will be "locked" and don't
  open.
* **Small Key Doors** are locked doors which consume the Small Keys when unlocked,
  unlike the colored doors where the key is multi-use. The player character can
  hold many small keys at once and only unlock as many doors as he has keys.

Updated Doodads:

* **Several doodads** were increased in size to better match the player character:
  Colored Locked Doors, Trapdoors, the Crumbly Floor and Electric Door, and the
  blue and orange Boolean State Blocks.
* **Colored Doors** now have a visual locked vs. unlocked state: while locked, a
  golden padlock hangs from the door, which goes away after it's been unlocked.
* **Switches** now interact differently with Electric Doors: the door will _always_
  toggle its current state regardless of the 'power' setting of the Switch.
* **Buttons** which are linked to a **Sticky Button** will press and stay down
  if the linked Sticky Button is pressed. Or in other words, the Sticky Button
  makes all linked Buttons act sticky too and stay pressed while the Sticky
  Button is pressed. If the Sticky Button is released later (e.g. by receiving
  power from a Switch) it releases its linked Buttons as well.

## v0.4.0-alpha (Nov 21 2020)

This update brings improvements to the editor; you can now fully draw all the
graphics for a custom doodad using the in-app tools!

A key difference between a Level and a Doodad is that doodads have **Layers**
where they keep multiple frames of animation or state. The editor tool now
supports working with these extra layers.

Other new features:

* The **Guidebook** has been updated with tons of good info and screenshots of
  the game's features. Press `F1` in-game to open the guidebook, or check it
  online at https://www.sketchymaze.com/guidebook/
* **Layer Selection Window for Doodads:** when you're editing a Doodad drawing,
  a "Lyr." button is added to the toolbar to access the Layers window.
* **Global UI popup modals:** when you're about to close a level with unsaved
  changes, you get an Ok/Cancel prompt confirming if you want to do that.
  Hitting the Escape key will ask you before just exiting the program. Alert
  boxes are supported too, and an `alert` command added to the developer console.

## v0.3.0-alpha (Sept 19 2020)

This update introduces the player character to the game. He doesn't have a name;
the game just refers to him as Boy.

His sprite size is bigger (33x54) than the 32x32 size of the placeholder player
character from before. So, the three example levels shipped with previous versions
of Project: Doodle no longer work. This release comes with two replacement levels
which are better decorated, anyway.

Other new features:

* **Palette Editor:** you can add and modify the colors on your level
  palette! Pick any colors you want, give a name to each swatch, and mark
  whether they behave as Solid, Water or Fire. Without any of these
  properties, colors are decorational-only by default.
* **Doodad Window:** when selecting the Actor or Link Tools in the editor,
  the Doodads window pops up. Doodads can be dragged from this window onto your
  level, instead of the Palette toolbar turning into a Doodad palette.
* The Palette toolbar on the Editor is now thinner and only shows colors.
  Mouse-over tooltips show the name and properties of each swatch.
* Added a --window option to the Doodle program to set the default window size.
  Options can be a resolution (e.g. 1024x768) or a special keyword
  "desktop", "mobile", "landscape" or "maximized"
  - Default size is desktop: 1024x768
  - Mobile and landscape mimic a smartphone at 375x812 resolution.

## v0.2.0-alpha (June 7 2020)

This release brings Sound Effects and Menus to the game.

New features:

* Added some User Documentation to ship with the game which teaches you how to
  create your own custom Doodads and program them with JavaScript. More
  documentation to come with time.
* Sound effects! Several doodads have a first pass at sound effects using some
  free sounds I found online. More doodads still need sounds and the existing
  sounds are by no means final. Buttons, switches, doors and keys have sound
  effects so far.
* The game now has a Menu Bar with pull-down menus in the Editor Mode instead
  of just a top panel with New/Save/Open buttons.

## v0.1.0-alpha (Apr 13 2020)

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

## v0.0.10-alpha (July 18 2019)

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

## v0.0.9-alpha (July 9 2019)

First alpha release.
