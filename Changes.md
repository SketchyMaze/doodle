# Changes

## v0.8.0 (September 4, 2021)

This release brings some new features, new doodads, and new levels.

New features:

* **Checkpoints** for gameplay will ease the pain of dying to fire
  pixels or Anvils by teleporting you back to the checkpoint instead
  of resetting the whole level.
* The **Doodad Properties** window while editing a doodad grants access
  to many features which were previously only available via the
  `doodad` tool, such as:
    * Edit metadata like the Title and Author of your doodad
    * Set the default hitbox of your doodad.
    * Attach, open, and delete the JavaScript for your doodad
    * Manage tags (key/value store) on your doodads: how you can
      communicate settings to the JavaScript which can receive the
      tags via `Self.GetTag("name")`
* Some **Generic Doodad Scripts** are built in. Using only the in-game
  tools, it is possible to create custom doodads which have some basic
  in-game logic and you don't need to write any code. The generic
  scripts include:
    * Generic Solid: the hitbox is solid
    * Generic Fire: its hitbox harms the player
    * Generic Anvil: harmless, deadly when falling
    * Generic Collectible Item: it goes in your inventory
* **All Characters are Playable!** Use the Link Tool to connect your
  Start Flag with another doodad on your level, and you will play
  **as** that doodad when the level starts. The Creature doodads are
  all intended to be fully functional; playing as buttons and doors
  leads to strange results.

New doodads have been added:

* The **Anvil** is a heavy metal object which is affected by gravity.
  It is harmless to collision, but if the anvil is in freefall, it
  will destroy every mobile doodad that it touches, and is deadly
  to the player character.
* The **Electric Trapdoor** is a trapdoor that opens and closes when
  powered by a button or switch. It is a horizontal version of the
  Electric Door.
* The **Thief** is a new character which will steal items from the
  player or other mobile doodads. The Thief is able to pick up items
  and unlock doors and he walks back and forth like the Azulians.
* The **Blue Azulian** is now selectable from the Doodads menu. It
  behaves like the Red Azulian but moves at half the speed. The
  Azulians can pick up items and open doors.
* The **Checkpoint Flag** will remember the player's spot in the level.
  Dying to fire pixels or Anvils no longer forces a restart of the
  level - you can resume from your last checkpoint, or the Start Flag
  by default.

New levels have been added:

* **Castle.level:** introduces the new Thief character. Castle-themed
  level showing off various new doodads.
* **Thief 1.level:** a level where you play as the Thief! You need to
  steal Small Keys from dozens of Azulians and even steal items back
  from another Thief who has already stolen some of the keys.

Some doodads have changed behavior:

* The **Bird** can no longer pick up items, unless controlled by
  the player character.
* The **Anvil** and **Box** will reset to their original locations
  if they receive a power signal from a linked button or switch.

The user interface has been improved:

* **Tabbed windows!** The Doodad Dropper window of the level editor
  and the Settings window use new, tabbed interfaces.
* **Doodad Categories:** the Doodad Dropper's tabs are divided into
  a few major categories.
    1. Objects: Anvil, Box, Crumbly Floor, and Flags
    2. Doors: Doors, Trapdoors, and Keys
    3. Gizmos: Buttons, Switches, Electric Doors, etc.
    4. Creatures: Bird, Azulians, Thief
    5. All: a classic view paging over all doodads (and doodads
       not fitting any of the above categories).

New functions are available in the JavaScript API for custom doodads:

* FailLevel(message string): global function that kills the player
  with a custom death message.
* SetCheckpoint(Point): set the player respawn location
* Self.MoveTo(Point(x, y int))
* Self.IsPlayer() bool
* Self.SetInventory(bool): turn on or off inventory. Keys and other
  items will now only give themselves to mobile doodads which have
  inventory.
* Self.HasInventory() bool
* Self.AddItem(filename string, quantity int) - zero quantity for
  permanent items like the colored keys.
* Self.RemoveItem(filename string, quantity int)
* Self.HasItem(filename string)
* Self.Inventory() map[string]int
* Self.Hitbox() - also see Self.Hitbox.IsEmpty()

The Events.OnLeave() callback now receives a CollideEvent argument,
like OnCollide, instead of the useless actor ID string. Notable
properties on the CollideEvent will be the .Actor which is leaving
and Settled=true.

Other miscellaneous changes:

* The **Link Tool** can now un-link two doodads by clicking on
  them again.
* Actor UUIDs in your levels will now be Type 1 UUIDs (time-based)
  instead of random. This will ensure each newly added doodad gets
  a larger ID than the previous one, so in cases of draw order
  conflicts or that sort of thing, the winner can be picked
  deterministically (most recently added renders on top).
* A **death barrier** will prevent Boy from falling forever on unbounded
  maps should he somehow fall off the level. The death barrier is a
  Y value 1,000 pixels below the lowest pixel on your map.
* Mobile doodads no longer "moonwalk" when they change directions.
* A new color is added to all default palettes: "hint" (pink) for
  writing hint notes.
* A maximum scroll speed on the "follow the player character" logic
  makes for cooler animations when the character teleports around.
* Levels and Doodads are now sorted on the Open menu.

## v0.7.2 (July 19 2021)

This release brings some new features and some new content.

New features:

* **Loading screens** have been added. In previous versions, 'chunks'
  of a level were rendered on-demand as they first scrolled onto the
  screen, but busy new chunks would cause gameplay to stutter. The
  loading screen is able to _pre-render_ the entire level up front
  to ensure gameplay is smooth(er).
* **Compression for Levels and Doodads:** levels and doodads are now compressed
  with Gzip for an average 88% smaller file size on disk. The example level,
  "Tutorial 2.level" shrank from 2.2 MB to 414 KB with compression. The game
  can still read old (uncompressed) files, but will compress them on save.

Some new content:

* **New Levels:** two desert-themed levels were added. One has you
  platforming and climbing the outside of a pyramid; the next is
  inside the pyramid and has puzzles involving movable boxes.
* **New Doodad:** Box is a pushable crate that is affected by
  gravity. It is taller than Boy and can be used to gain some extra
  height and reach higher platforms.
* **New Palette:** To the "Colored Pencil" palette was added a new
  default color: Sandstone (solid).
* **New Pattern:** Perlin Noise was added as one of the available
  brush patterns. Sandstone uses this pattern by default.

Some miscellaneous changes:

* **Slowly scroll in editor:** holding down the Shift key while scrolling the
  level editor will scroll your drawing _very_ slowly.
* **'Doodads' Hotkey:** in the Level Editor, the `Q` key will open
  the Doodads window instead of the `D` key, as the `D` key is now
  part of WASD for scrolling the drawing.

## v0.7.1 (July 11 2021)

Fixes a bug on the Windows version:

* Built-in wallpapers other than the default Notebook were failing to
  load in the Windows release of v0.7.0

## v0.7.0 (June 20 2021)

This is the first release of the game where the "free version" drifts meaningfully
away from the "full version". Free versions of the game will show the label
"(shareware)" next to the game version numbers and will not support embedding
doodads inside of level files -- for creating them or playing them.
Check the website for how you can register the full version of the game.

This release brings several improvements to the game:

* **Brush Patterns** for your level palette. Instead of your colors drawing on as
  plain, solid pixels, a color swatch can _sample_ with a Pattern to create a
  textured appearance when plotted on your level. Several patterns are built in
  including Noise, Marker, Ink, and others. The idea is that your brush strokes can
  look as though they were drawn in pencil graphite or similar.
* **Title Screen:** the demo level shown on the title screen will leisurely scroll
  around the page. The arrow keys may still manually scroll the level any direction.
* **Attach Doodads to Level Files:** this is the first release that supports _truly_
  portable custom levels! By attaching your custom doodads _with_ your custom level
  file, it will "just play" on someone else's computer, and they don't need to copy
  all your custom doodads for it to work! But, free versions of the game will not
  get to enjoy this feature.
* **Settings UI**: a "Settings" button on the home screen (or the Edit->Settings
  menu in the editor) will open a settings window. Check it out!
* **Horizontal Toolbars option:** if enabled in the Settings window, the toolbar
  and palette in the Editor will be horizontal instead of vertical, along the top
  and bottom of the screen. This may be better optimized for smartphone-sized
  screens like the Pinephone. If the program is started with `-w mobile` the first
  time, it will use horizontal toolbars by default.

Some small bits of polish in the game's user interface:

* Some buttons are more colorful! The "Ok" button in alert boxes is blue and pressing
  Enter will select the blue button.
* When opening a drawing to play or edit, a blue **Browse...** button is
  added so you can more easily find downloaded custom levels and play them.
* In the Level Editor, the "Level ->  **Attached Files**" menu will let you see
  and manage files attached to your level, such as its custom wallpaper image or
  any custom doodads that were published with the level.
* The keyboard shortcut to open the developer console is now the tilde/grave key
  (`) instead of Enter.

Bugs fixed:

* The WASD keys to move the player character (as an alternative to the arrow keys)
  now works more reliably. Previously, they were affected by key-repeat so Boy would
  do a quick hop followed by a longer one when pressing W to jump. Also, his
  animation would not update correctly when moving via the WASD keys. Both bugs
  are fixed in this release.
* Shortcut keys advertised in the menu, such as Ctrl-N and Ctrl-S, now actually work.

## v0.6.0-alpha (June 6 2021)

This release brings less jank and some new features.

The new features:

* **Choice of Default Palette for New Levels:** when creating a new level, a
  "Palette:" option appears which allows you to set the default colors to start
  your level with. The options include:
  * Default: the classic default 4 colors (black, grey, red, blue).
  * Colored Pencil: a set with more earthy tones for outdoorsy levels
    (grass, dirt, stone, fire, water)
  * Blueprint: the classic Blueprint wallpaper theme, a bright version of Default
    for dark level backgrounds.
* **Custom Wallpapers:** unhappy with the default, paper-themed level background
  images? You can now use your own! They attach to your level data for easy
  transport when sharing your level with others.
* **More Wallpapers:** a couple of new default wallpapers are added: Graph paper
  and Dotted paper. They are both on light white paper and offer a light
  alternative to Blueprint.

Some bugs fixed:

* **Collision fixes:** you should be able to walk up gentle slopes to the left
  without jumping, as easily as you could to the right.
* **Debugging:** the F4 key to show collision hitboxes around all doodads in
  Play Mode now functions again, and draws boxes around _all_ doodads, not just the 
  player character.
* **Hitboxes are tighter:** a doodad's declared hitbox size (from their JavaScript)
  is used when a doodad collides against level geometry or other doodads. Meaning:
  Boy, who has a narrow body but a square sprite size, now collides closer with
  objects on his right side.
* **Physics are tweaked:** Boy now moves and accelerates slightly faster.

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
