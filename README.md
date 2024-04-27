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

# Developer Documentation

In case you are reading this from the game's open source repository, take a
look in the `docs/` folder or the `*.md` files in the root of the repository.

Some to start with:

* [Building](Building.md) the game (tl;dr. run bootstrap.py)
* [Tour of the Code](docs/Tour%20of%20the%20Code.md)

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

The doodle engine for _Sketchy Maze_ is released as open source software under
the terms of the GNU General Public License. The assets to the game, including
its default doodads and levels, are licensed separately from the doodle engine.
Any third party fork of the doodle engine MUST NOT include any official artwork
from Sketchy Maze.

    Doodle Engine for Sketchy Maze
    Copyright (C) 2022  Noah Petherbridge

    This program is free software: you can redistribute it and/or modify
    it under the terms of the GNU General Public License as published by
    the Free Software Foundation, either version 3 of the License, or
    (at your option) any later version.

    This program is distributed in the hope that it will be useful,
    but WITHOUT ANY WARRANTY; without even the implied warranty of
    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
    GNU General Public License for more details.

    You should have received a copy of the GNU General Public License
    along with this program.  If not, see <http://www.gnu.org/licenses/>.