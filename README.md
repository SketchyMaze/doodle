# Sketchy Maze

> **Homepage:** https://www.sketchymaze.com

Sketchy Maze is a drawing-based maze game.

The game is themed around hand-drawn, side-scrolling platformer
type mazes. You can draw your own levels using freehand and basic drawing tools,
color in some fire or water, and drag in pre-made "Doodads" like buttons, keys
and doors to add some interaction to your level.

This is a _very_ early pre-release version of the game. Expect bugs and slowness
but get a general gist of what the game is about.

This alpha release of the game comes with some example levels built-in for
playing or editing and a handful of built-in Doodads.

See the **Guidebook** included with this game for good user-facing
documentation or online at https://www.sketchymaze.com/guidebook

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

(Eventually) all these features will support custom content in the
game:

* Users can create **Custom Doodads** to extend the game with a scripting
  language like JavaScript. The sprites and animations are edited in-game
  in Edit Mode, but the scripting is done in your text editor.
* In **Edit Mode** you can drag custom doodads into your maps.
* To **Share** your maps, you can choose to **bundle** the custom
  doodads inside your map file itself, so that other players can play
  the map without needing to install the doodads separately.
* If you receive a map with custom doodads, you can **install** the doodads
  into your copy of the game and use them in your own maps.

# Keybindings

Global Keybindings:

```
Escape
  Close the developer console if open, without running any commands.
  Exit the program otherwise.

Enter
  Open and close the developer console, and run commands while the
  console is open.

F3
  Toggle the Debug Overlay.

F4
  Toggle debug collision hitboxes.
```

In Play Mode:

```
Cursor Keys
  Move the player around.
"E" Key
  Edit the map you're currently playing if you came from Edit Mode.
```

In Edit Mode:

```
Cursor Keys
  Scroll the view of the map around.
"P" Key
  Playtest the current map you're working on.
"F" Key
  Switch to the Pencil (Freehand) Tool
"L" Key
  Switch to the Line Tool
"R" Key
  Switch to the Rectangle Tool
Ctrl-Z
  Undo
Ctrl-Y
  Redo
```

# Gamepad Controls

The game supports Xbox and Nintendo style game controllers. The button
bindings are not yet customizable, except to choose between the
"X Style" or "N Style" for A/B and X/Y button mappings.

Gamepad controls very depending on two modes the game can be in:

## Mouse Mode

The Gamepad emulates a mouse cursor in this mode.

* The left analog stick moves a cursor around the screen.
* The right analog stick scrolls the level (title screen and editor)
* A or X button simulates a Left-click
* B or Y button simulates a Right-click
* L1 (left shoulder) emulates a Middle-click
* L2 (left trigger) closes the top-most window in the editor mode
  (like the Backspace key).

## Gameplay Mode

When playing a level, the controls are as follows:

* The left analog stick and the D-Pad will move the player character.
* A or X button to "Use" objects such as Warp Doors.
* B or Y button to "Jump"
* R1 (right shoulder) toggles between Mouse Mode and Gameplay Mode.

You can use the R1 button to access Mouse Mode to interact with the
menus or click on the "Edit Level" button.

Note: characters with antigravity (such as the Bird) can move in all
four directions but characters with gravity only move left and right
and have the dedicated "Jump" button. This differs from regular
keyboard controls where the Up arrow is to Jump.

# Built-In Doodads

A brief introduction to the built-in doodads available so far:

- **Characters**
  - Blue Azulian: this is used as the player character for now. If
    dragged into a level, it doesn't do anything but is affected
    by gravity.
  - Red Azulian: an example mobile mob for now. It walks back and
    forth, changing directions only when it comes across an
    obstacle and can't proceed any further.
- **Doors and Keys**
  - Colored Doors: these act like solid barriers until the player or
    another doodad collects the matching colored key.
  - Colored Keys: these are collectable items that allow the player or
    another doodad to open the door of matching color. Note that
    inventories are doodad-specific, so other mobs besides the
    player can steal a key before the player gets it! (For now)
  - Electric Door: this mechanical door stays closed and only
    opens when it receives a power signal from a linked button.
  - Trapdoor: this door allows one-way access, but once it's closed
    behind you, you can't pass through it from that side!
- **Buttons**
  - Button: while pressed by a player or other doodad, the button
    emits a power signal to any doodad it is linked to. Link a button
    to an electric door, and the door will open while the button is
    pressed and close when the button is released.
  - Sticky Button: this button will stay pressed once activated and
    will not release unless it receives a power signal from another
    linked doodad. For example, one Button that links to a Sticky
    Button will release the sticky button if pressed.
- **Switches**
  - Switch: when touched by the player or other doodad, the switch will
    toggle its state from "OFF" to "ON" or vice versa. Link it to an
    Electric Door to open/close the door. Link switches _to each other_ as
    well as to a door, and all switches will stay in sync with their ON/OFF
    state when any switch is pressed.
- **Crumbly Floor**
  - This rocky floor will break and fall away after being stepped on.
- **Two State Blocks**
  - Blue and orange blocks that will toggle between solid and pass-thru
    whenever the corresponding ON/OFF block is hit.
- **Start and Exit Flags**
  - The "Go" flag lets you pick a spawn point for the player character.
  - The "Exit" flag marks the level goal.

# Developer Console

Press `Enter` at any time to open the developer console. The console
provides commands and advanced functionality, and is also where cheat
codes can be entered.

Commands supported:

```
close
  Exit to the game's title screen.

new
  Show the "New Level" screen to start editing a new map.

save [filename]
  Save the current map in Edit Mode. The filename is required
  if the map has not been saved yet.

edit [filename]
  Open a map or doodad in Edit Mode.

play [filename]
  Open a map in Play Mode.

echo <text>
  Flash a message to the console.

alert <text>
  Test an alert box modal with a custom message.

clear
  Clear the console output history.

exit
quit
  Close the developer console.

boolProp <property> <true/false>
  Toggle certain boolean settings in the game. Most of these
  are debugging related. `boolProp list` shows the available
  props.

eval <expression>
$ <expression>
  Execute a line of JavaScript code in the console. Several
  of the game's core data types are available here; `d` is
  the master game struct; d.Scene is the pointer to the
  current scene. d.Scene.UI.Canvas may point to the level edit
  canvas in Editor Mode. Object.keys() can enumerate public
  functions and variables.

repl
  Enters an interactive JavaScript shell, where the console
  stays open and pre-fills a $ prompt for subsequent commands.
```

The JavaScript console is a feature for advanced users and was
used while developing the game. Cool things you can do with it
may be documented elsewhere.

## Cheat Codes

The following cheats can be entered into the developer console.

Play Mode:

* `import antigravity`
  - This disables the effects of gravity for the player
    character. Arrow keys can freely move the player in any direction.
* `ghost mode`
  - This disables collision detection for the player character
    so that you can pass through walls and solid doodads. Combine with
    antigravity or else you'll fall to the bottom of the map!
* `give all keys`
  - Adds all four colored keys to the player's inventory.
* `drop all items`
  - Clears the player's inventory of all items.

Experimental:

* `unleash the beast`
  - Removes the 60 FPS frame rate lock, allowing the game to run as quickly
    as your hardware permits.
* `don't edit and drive`
  - Allows editing the level _while_ you're playing it: you can click and drag
    new pixels with the freehand pencil tool.
* `scroll scroll scroll your boat`
  - Enables Editor Mode scrolling (with the arrow keys) while playing a level.
    The player character must always remain on screen though so you can't
    scroll too far away.

Unsupported shell commands (here be dragons):

* `reload`: reloads the current 'scene' within the game engine, using the
  existing scene's data. If playing a level this will start the level over.
  If editing a level this will reload the editor, but your recent unsaved
  changes _should_ be left intact.
* `guitest`: loads the GUI Test scene within the game. This was where I
  was testing UI widgets early on; not well maintained; the `close`
  command can get you out of it.

## Environment Variables

To enable certain debug features or customize some aspects of the game,
run it with environment variables like the following:

```bash
# Draw a semi-transparent yellow background over all level chunks
$ DEBUG_CHUNK_COLOR=FFFF0066 ./doodle

# Set a window size for the application
# (equivalent to: doodle --window 1024x768)
$ DOODLE_W=1024 DOODLE_H=768 ./doodle

# Turn on lots of fun debug features.
$ DEBUG_CANVAS_LABEL=1 DEBUG_CHUNK_COLOR=FFFF00AA \
  DEBUG_CANVAS_BORDER=FF0 ./doodle
```

Supported variables include:

* `DOODLE_W` and `DOODLE_H` set the width and height of the application
  window. Equivalent to the `--window` command-line option.
* `D_SCROLL_SPEED` (int): tune the canvas scrolling speed. Default might
  be around 8 or so.
* `D_DOODAD_SIZE` (int): default size for newly created doodads
* `D_SHELL_BG` (color): set the background color of the developer console
* `D_SHELL_FG` (color): text color for the developer console
* `D_SHELL_PC` (color): color for the shell prompt text
* `D_SHELL_LN` (int): set the number of lines of output history the
  console will show. This dictates how 'tall' it rises from the bottom
  of the screen. Large values will cover the entire screen with console
  whenever the shell is open.
* `D_SHELL_FS` (int): set the font size for the developer shell. Default
  is about 16. This also affects the size of "flashed" text that appears
  at the bottom of the screen.
* `DEBUG_CHUNK_COLOR` (color): set a background color over each chunk
  of drawing (level or doodad). A solid color will completely block out
  the wallpaper; semitransparent is best.
* `DEBUG_CANVAS_BORDER` (color): the game will draw an insert colored
  border around every "Canvas" widget (drawing) on the screen. The level
  itself is a Canvas and every individual Doodad or actor in the level is
  its own Canvas.
* `DEBUG_CANVAS_LABEL` (bool): draws a text label over every Canvas
  widget on the screen, showing its name or Actor ID and some properties,
  such as Level Position (LP) and World Position (WP) of actors within
  a level. LP is their placement in the level file and WP is their
  actual position now (in case it moves).

# Author

Copyright (C) 2021 Noah Petherbridge. All rights reserved.
