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

# Keybindings

Global Keybindings:

```
Escape
  Close the developer console if open, without running any commands.
  Exit the program otherwise.

Enter
  Open and close the developer console, and run commands while the console
  is open.
```

In Play Mode:

```
Cursor Keys
  Move the player around.
```

In Edit Mode:

```
F12
  Take a screenshot (generate a PNG based on level data)
```

## Developer Console

Press `Enter` at any time to open the developer console.

Commands supported:

```
new
  Create a new map in Edit Mode.

save [filename.json]
  Save the current map in Edit Mode. The filename is required if the map has
  not been saved yet.

edit [filename.json]
  Open a map in Edit Mode.

play [filename.json]
  Open a map in Play Mode.

echo <text>
  Flash a message to the console.

clear
  Clear the console output history.

exit
quit
  Close the developer console.
```

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
* [x] Abstract away SDL logic into a small corner so it can be replaced with
  OpenGL or something later on.
* [x] Implement a command line shell in-game to ease development before a user
  interface is created.
  * [x] Add support for the shell to pop itself open and ask the user for
    input prompts.

## Alpha Platformer

* [x] Inflate the pixel history from the map file into a full lookup grid
  of `(X,Y)` coordinates. This will be useful for collision detection.
* [x] Create a dummy player character sprite, probably just a
  `render.Circle()`. In **Play Mode** run collision checks and gravity on
  the player sprite.
  * [x] Create the concept of the Doodad and make the player character
    implement one.
* [x] Get basic movement and collision working. With a cleanup this can
  make a workable **ALPHA RELEASE**
  * [x] Ability to move laterally along the ground.
  * [x] Ability to walk up reasonable size slopes but be stopped when
    running against a steeper wall.
  * [x] Basic gravity

## UI Overhaul

* [x] Create a user interface toolkit which will be TREMENDOUSLY helpful
  for the rest of this program.
  * [x] Labels
  * [x] Buttons (text only is OK)
      * [x] Buttons wrap their Label and dynamically compute their size based
        on how wide the label will render, plus padding and border.
      * [x] Border colors and widths and paddings are all configurable.
      * [x] Buttons should interact with the cursor and be hoverable and
        clickable.
  * [x] UI Manager that will keep track of buttons to know when the mouse
    is interacting with them.
  * [x] Frames
      * Like Buttons, can have border (raised, sunken or solid), padding and
        background color.
      * [x] Should be able to size themselves dynamically based on child widgets.
  * [x] Windows (fixed, non-draggable is OK)
      * [x] Title bar with label
      * [x] Window body implements a Frame that contains child widgets.
      * [x] Window can resize itself dynamically based on the Frame.
* [x] Create a "Main Menu" scene with buttons to enter a new Edit Mode,
  play an existing map from disk, etc.
* [x] Add user interface Frames or Windows to the Edit Mode.
  * [x] A toolbar of buttons (New, Save, Open, Play) can be drawn at the top
    before the UI toolkit gains a proper MenuBar widget.
  * [x] Expand the Palette support in levels for solid vs. transparent, fire,
    etc. with UI toolbar to choose palettes.
* [x] Add a "Canvas" widget that will hold level drawing data and abstract it
  out such that the Canvas can have a constrained width and height, position,
  and "scrollable" viewport area that determines rendered pixels. Will be VERY
  useful for Doodads and working on levels smaller than your window size (like
  doodads).

Lesser important UI features that can come at any later time:

* [ ] MenuBar widget with drop-down menu support.
* [x] Checkbox and Radiobox widgets.
* [ ] Text Entry widgets (in the meantime use the Developer Shell to prompt for
  text input questions)

## Doodad Editor

* [x] The Edit Mode should support creating drawings for Doodads.
  * [x] It should know whether you're drawing a Map or a Doodad as some
    behaviors may need to be different between the two.
  * [ ] Compress the coordinates down toward `(0,0)` when saving a Doodad,
    by finding the toppest, leftest point and making that `(0,0)` and adjusting
    the rest accordingly. This will help trim down Doodads into the smallest
    possible space for easy collision detection.
  * [ ] Add a UX to edit multiple frames for a Doodad.
  * [ ] Edit Mode should be able to fully save the drawings and frames, and an
    external CLI tool can install the JavaScript into them.
  * [ ] Possible UX to toggle Doodad options, like its collision rules and
    whether the Doodad is continued to be "mobile" (i.e. doors and buttons won't
    move, but items and enemies may be able to; and non-mobile Doodads don't
    need to collision check against level geometry).
* [ ] Edit Mode should have a Doodad Palette (Frame or Window) to drag
  Doodads into the map.
* [ ] ???

# Building

Fedora dependencies:

```bash
$ sudo dnf install SDL2-devel SDL2_ttf-devel
```

## Fonts

The `fonts/` folder is git-ignored. The app currently uses font files here
named:

* `DejaVuSans.ttf` for sans-serif font.
* `DejaVuSans-Bold.ttf` for bold sans-serif font.
* `DejaVuSansMono.ttf` for monospace font.

These are the open source **DejaVu Sans [Mono]** fonts, so copy them in from
your `/usr/share/fonts/dejavu` folder or provide alternative fonts.

```bash
mkdir fonts
cp /usr/share/fonts/dejavu/{DejaVuSans.ttf,DejaVuSans-Bold.ttf,DejaVuSansMono.ttf} fonts/
```
