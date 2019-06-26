# Project: Doodle (Working Title)

Project: Doodle is a drawing-based maze game written in Go.

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
```

# Built-In Doodads

A brief introduction to the built-in doodads available so far:

- **Characters**
  - Blue Azulian: this is used as the play character for now. If
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

## Developer Console

Press `Enter` at any time to open the developer console.

Commands supported:

```
new
  Show the "New Level" screen to start editing a new map.

save [filename]
  Save the current map in Edit Mode. The filename is required if the map has
  not been saved yet.

edit [filename]
  Open a map or doodad in Edit Mode.

play [filename]
  Open a map in Play Mode.

echo <text>
  Flash a message to the console.

clear
  Clear the console output history.

exit
quit
  Close the developer console.
```

# Building

Fedora dependencies:

```bash
$ sudo dnf install SDL2-devel SDL2_ttf-devel
```

# Author

Copyright (C) 2019 Noah Petherbridge. All rights reserved.
