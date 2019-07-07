# Ideas

## Table of Contents

* [Major Milestones](#major-milestones)
* [Release Modes](#release-modes)
* [File Formats](#file-formats)
* [Text Console](#text-console)
* [Doodads](#doodads)

# Major Milestones

The major milestones of the game are roughly:

* [x] Prototype: make a simple SDL painting program that does nothing special.
* [x] Simple Platformer: be able to toggle between "edit mode" and "play mode"
  and control a character who can walk around your level and bump into the
  solid geometry you've drawn (no objects yet, just the basics here).
* [ ] Add Doodads (buttons, doors, the player character themself, enemies, ...)
  * Share a lot in common with map drawings, in that they're hand-drawn, will
    share a similar file format.
  * Available doodads can be dragged/dropped into maps.
  * The player character should be a Doodad under the hood to keep it from
    becoming too special (read: easier to make the game multiplayer in the
    future by putting a "networked user" in control of a doodad instead of
    the keyboard/mouse).
* [ ] **Version 1:** Single Player Campaign and Editor. This is the minimum
  feature set for a first public release of the game. Required features:
  * The game should ship with a single-player "campaign mode" of pre-made maps
    that link to one another in sequence. i.e. 100 levels that the player can
    play through in a certain order.
  * It must include the level editor feature so players can create and share
    their own maps.
  * Dev tools may be clunky to use at this stage; i.e. players creating custom
    Doodads will need to use external tools outside the game (i.e. code editors
    to program the JavaScript logic of the doodad), but everything should be
    available and possible for modders to extend the game with custom features.
  * Game should have a good mixture of doodads and features: doors, buttons,
    switches, etc. and make a usable single player experience.
  * World sizes might be limited in dimension.
* [ ] **Version 2:** Multiplayer Collaborative World Builder. This is a
  "pie in the sky" long-term vision for the game, to make it multiplayer,
  hopefully addicting, and possibly slightly Minecraft-like. Some ideas:
  * Players can self-host their own multiplayer servers to draw worlds with
    friends.
  * A new server would initialize as a blank white level with maybe a single
    platform (a line) for players to spawn on.
  * Gameplay is a mixture of players drawing the world and playing on it.
    * i.e.: one player could be drawing himself a castle and, as he's drawing,
      another player could be walking on the lines being laid down, etc.
  * World size should be infinite.
    * Version 1 can have limited world sizes as it will probably be easier,
      but this should be opened up eventually.
  * Besides creative mode, other game modes should be explored eventually...
    * Automatically-spawning enemy doodads that you have to fight?
    * Procedurally generated default maps? Having a blank white canvas is
      sorta like Superflat worlds in Minecraft, whereas normal Minecraft worlds
      come with randomly generated terrain to start from.
    * Find a way to incorporate drawing into a survival mode game? i.e. instead
      of a "Creative Mode" style, "unlimited ink to draw as much as you want,"
      have some natural limiter where players have to spend time in Play Mode
      to be able to change the map.

# Edit Mode Features

* [ ] A menu bar along the top of the screen with normal drop-down menus.
  UI toolkit widgets needed:
  * [ ] Menu: a pop-up menu, like one summoned by a right-click action.
  * [ ] MenuButton: a Button widget that opens a Menu when clicked.
  * [ ] MenuBar: a Frame widget that automatically finds the top of your
    window and makes it easy to fill it with MenuButtons.
* [x] A status bar that shows your cursor position and other useful details.
* [x] A palette window that shows you your current palette as a series of
  radio buttons, and you can toggle between the palette choices.
  * Palettes are saved with the level file and the list is dynamic.
  * Colors are not tied to behaviors. Each "Swatch" on the palette has its own
    color and a set of boolean flags for `solid`, `fire` and `water` behaviors.
  * [ ] User interface to edit (add/remove) swatches from the palette.
* [x] A Toolbox window with radio buttons to select between various drawing tools.
  * [x] Pencil (the default) draws single pixels on the level.
  * [x] Rectangle would draw a rectangular outline.
  * [x] Line would draw a line from point to point.
* [ ] A way to adjust brush properties:
  * [ ] Brush size, shape (round or square).
* [ ] Tools to toggle "layers" of visibility into your level:
  * A drop-down menu with options like "Show all solid", "Show all fire",
    "Show all decoration", etc.
  * When layers are applied, adjust the way the pixels in the Grid are drawn on
    screen. Any pixel that doesn't fit the layers requested should draw in a
    muted color like light grey, and the layers requested should show in a
    vibrant color.
  * Use case: a full-color level can have _many_ solid pixels (grass, dirt,
    walls) rendering in all kinds of colors. But you want to see how the collision
    burden will be for the level and you can toggle on "Show all solid pixels"
    and every solid pixel becomes bright red and all the non-solid pixels turn a
    muted grey. This way I can easily see the difference between the colors and
    behaviors of the pixels in my level.

For creating Doodads in particular:

* [x] Make a way to enter Edit Mode in either "Level Mode" or "Doodad Mode",
  i.e. by a "New Level" or "New Doodad" button.
* [ ] Create a "frame manager" window to see and page between the frames of the
  drawing.
* [x] Ability to work on canvases with constrained size (including smaller than
  your window). This will use a Canvas widget in the UI toolkit as an abstraction
  layer. Small canvases will be useful for drawing doodads of a fixed size.

# Release Modes

## Shareware/Demo Version

This would be a free version with some limitations. Early public alpha releases
would be built with this release mode.

* Optional expiration date after which the game WILL NOT run.
* Can play the built-in maps and create your own custom maps.
* No support for Custom Doodads. The game will have the code to read Doodads from
  disk dummied out/not compiled in, and any third-party map that embeds or
  references custom Doodads will not be allowed to run.
* Custom maps created in a demo version will have some feature limitations:
  * Infinite map sizes not allowed, only bounded ones with a fixed size.
  * No custom wallpaper images, only built-in ones.
  * No custom palette for new maps, only the default standard palette.
  * No features for drawing doodad graphics (multiple frames, etc.)

As an end user, it means basically:

* You are limited to built-in doodads but you can make (and share) and play
  other users' custom maps that only use the built-in doodads.

## Release Version

TBD.

Probably mostly DRM free. Will want some sort of account server early-on though.

# File Formats

* The file formats should eventually have a **Protocol Buffers** binary
  representation before we go live. JSON support shall remain, but the
  production application will not _write_ JSON files, only read them.
  (This way we can ship drawings in the git repo as text files).
* The app will support reading three types of files:
  * `.canvas` files are the lowest common denominator, raw drawing data. It
    contains a Palette and a pixel grid and nothing more.
  * `.map` files are level maps. They include a Canvas along with level
    metadata, Doodad array, attached files, etc.
  * `.doodad` files are for doodads. They include a Canvas along with
    metadata, embedded JavaScript, attached files, etc.
* JSON versions will have `.json.<ext>` file suffixes, like `.json.canvas`
  or `.json.map`
* The **production** app will be only be able to read the binary format of
  the files. The JSON reading code is for dev builds only.
* Shareware/Demo builds will have even more restrictions on supported file
  types. For example it won't be built with the code that allows it to
  read _or_ write a Doodad from disk, as it will be limited only to built-in
  Doodads and won't support authoring custom ones.

## Common Drawing Files

* A common base format should be shared between Levels and Doodads. You should
  be able to use the Editor mode and draw a map *or* draw a doodad like a
  button. The drawing data should be a common structure between Level and
  Doodad files.
* The drawing is separated between a **Palette** and the **Pixels**
  themselves. The Pixels reference the Palette values and their X,Y
  coordinate.
* The _color_ and the _behavior_ of the palette are decoupled.
  * In the base game, all the solid lines you draw may be black and red
    lines are fire, but these aren't hard and fast rules. You could hack a
    custom map file that makes black lines fire and red lines water if
    you wanted.
  * The Palette in the map file stores the attributes and colors of each
    distinct type of pixel used in the map. Here it says "color 0 is
    black and is solid", "color 1 is red and is fire and is not solid",
    etc.
  * A mod tool could be written to produce a full-color pixel art level
    that still behaves and follows the normal rules of the Doodle game
    with regards to geometry and collisions.
* Ideas for pixel attributes:
  * Brush: what shape brush to draw the pixel with.
  * Solid: can't collide with other solid pixels.
  * Fire: applies fire damage to doodads that intersect with it.
  * Water: If a doodad passes through a blue pixel, they toggle their
    underwater physics. This way pools can be entered from ANY side (top,
    bottom, sides) and the physics should toggle on and off.
  * Slippery: when a doodad is standing on a slippery pixel, do some extra
    checks to find a slope and slide the doodad down it. Makes the pixels
    act like ice.
* Standard palette:
  * The base game's map editor will tend toward hand-drawn style, at least
    at first.
  * Black lines are solid.
  * Dashed black lines are slippery.
  * Red lines are fire.
  * Blue lines are water.
  * Light grey lines are decoration (non solid, background geometry)
  * May make it possible to choose arbitrary colors separately from the
    type of pixel. A palette manager UX would be great.

## Level Files

* In the level file, store the `pixelHistory` as the definitive source
  of pixels rather than the grid of pixels. Let the grid be populated when
  the level is being inflated. The grid should have `json:"-"` so it doesn't
  serialize to the JSON.
  *  This makes it possible to animate levels as they load -- by
     fast-tracing the original lines that the mapper drew, watching them draw
     the map before you play it.
  * Makes the file _slightly_ lighter weight because a lot of lines will have
    delta positions in the pixelHistory so we don't need to store the middle
    pixels.
* It should have space to store copies of any custom Doodads that the user
  wants to export with the level file itself, for easy sharing.
* It should have space to store a custom background image.

## Wallpaper Images

* Levels can pick a "wallpaper image" to go behind their pixels. One example of
  a wallpaper would be a sheet of standard ruled notebook paper.
* The texture file will be a square (rectangular maybe ok) with four quadrants
  from which the textures will be extracted. For example if the overall image
  size was 100x100 pixels, it will be divided into the four 50x50 quadrants.
  1. `Corner`: Top left corner is the top left edge of the "page" the level is on
  2. `Top`: Top right corner is the repeated "top of page" texture.
  3. `Left`: Bottom left corner is the repeated "left of page" texture.
  4. `Repeat`: Bottom right corner is the repeated background texture that extends
     infinitely in all directions.
* The Repeat texture is used all the time, and the other three are used when the
  level type has boundaries (on the top and left edges in particular) to draw
  decorative borders instead of the Repeat texture.
* Levels will be able to choose a "page type" which controls how the wallpaper
  will be drawn and how the level boundaries may be constrained. There will be
  four options:
  1. **Unbounded:** The map can freely grow in any direction, including into the
     negative X/Y coordinates. The map author will not run up against a boundary
     as the level grows in any direction.
  2. **No Negative Space:** The map coordinates can not dip below `(0,0)`, the
     origin at the top-left edge of the map. The map can grow infinitely in the
     positive X and Y directions (to the right and down) but is constrained on
     the left and right edges. The game engine will stop scrolling the map when
     the top or left edges are reached, and those edges will behave like a solid
     wall.
  3. **Bounded:** The map has a fixed width and height and is bounded on all
     four edges.
  4. **Bordered:** same as Bounded but with a different wallpaper behavior.
     The bottom and right edges are covered with mirror images of the top and
     left edges.
* The page types will have their own behaviors with how wallpapers are drawn:
  * **Unbounded:** only the `BR` texture from the wallpaper is used, repeated
    infinitely in the X and Y directions. The top-left, top, and left edge
    textures are not used.
  * **No Negative Space:** the `TL` texture is drawn at coordinate `(0,0)`.
    To its right, the `TR` texture is repeated forever in the X direction, and
    along the left edge of the page, the `BL` texture is repeated in the Y
    direction. The remaining whitespace on the page repeats the `BR` texture
    infinitely.
  * **Bounded:** same as No Negative Space.
  * **Bounded, Mirrored Wallpaper:** same as No Negative Space, but all of the
    _other_ corners and edges are textured too, with mirror images of the Top,
    Top Left, and Left textures. This would look silly on the "ruled notebook"
    texture, but could be useful to emborder the level with a fancy texture.
* The game will come with a few built-in textures for levels to refer to by
  name. These textures don't need to be distributed with the map files themselves,
  as every copy of the game should include these (or a sensible fallback would
  be used).
* The map author can also attach their own custom texture that will be included
  inside the map file.

### Default Wallpapers

**notebook**: standard ruled notebook paper with a red line alone the Left
dge and a blank margin along the Top, with a Corner and the blue lines
aking up the Repeat in all directions.

![notebook.png](../assets/wallpapers/notebook.png)

**graph**: graph paper made up of a grid of light grey or blue lines.

**dots**: graph paper made of dots at the intersections but not the lines in
between.

**legal**: yellow lined notebook paper (legal pad).

**placemat**: a placemat texture with a wavy outline that emborders the map
on all four sides. To be used with the Bordered level type.

# Text Console

* Create a rudimentary dev console for entering text commands in-game. It
  will be helpful until we get a proper UI developed.
  * The `~` key would open the console.
  * Draw the console on the bottom of the screen. Show maybe 6 lines of
    output history (a `[]string` slice) and the command prompt on the
    bottom.
* Ideas for console commands:
  * `save <filename.json>` to save the drawing to disk.
  * `open <filename.json>`
  * `clear` to clear the drawing.
* Make the console scriptable so it can be used as a prompt, in the mean
  time before we get a UI.
  * Example: the key binding `Ctrl-S` would be used to save the current
    drawing, and we want to ask the user for a file name. There is no UI
    toolkit yet to draw a popup window or anything.
  * It could be like `console.Prompt("Filename:")` and it would force open
    the text console (if it wasn't already open) and the command prompt would
    have that question... and have a callback command to run, like
    `save <filename.json>` using their answer.

# Doodads

Doodads will be the draggable, droppable, scriptable assets that make the
mazes interactive.

* They'll need to store multiple frames, for animations or varying states.
  Example: door opening, button being pressed, switch toggled on or off.
* They'll need a scripting engine to make them interactive. Authoring the
  scripts can be done externally of the game itself.
* The built-in doodads should be scripted the same way as custom doodads,
  dogfooding the system.
* Custom doodads will be allowed to bundle with a level file for easy
  shipping.
  * Installing new doodads from a level file could be possible too.
* Doodads within a level file all have a unique ID, probably just an
  integer. Could be just their array index even.

Some ideas for doodad attributes:

* Name (string)
* Frames (drawings, like levels)

Doodad instances in level files would have these attributes:

* ID (int)
* X,Y coordinates
* Target (optional int; doodad ID):
  * For buttons and switches and things. The target would be another
    doodad that can be interacted with.
  * Self-contained doodads, like trapdoors, won't have a Target.
* Powered (bool)
  * Default `false` and most things won't care.
  * A Button would be default `false` until pressed, then it's `true`
  * A Switch is `true` if On or `false` if Off
  * A Door is `true` if Open and `false` if Closed
  * So when a switch is turned on and it opens a door by pushing a `true`
    state to the door... this is the underlying system.

## Scripting

* Probably use Otto for a pure Go JavaScript runtime, to avoid a whole world
  of hurt.
* Be able to register basic event callbacks like:
  * On load (to initialize any state if needed)
  * On visible (for when we support scrolling levels)
  * On collision with another doodad or the player character
  * On interaction (player hits a "Use" button, as if to toggle a switch)
* Doodads should be able to pass each other messages by ID.
  * Example: a Button should be able to tell a Door to open because the
    button has been pressed by another doodad or the player character.

Some ideas for API features that should be available to scripts:

* Change the direction and strength of gravity (i.e. Antigravity Boots).
* Teleport the player doodad to an absolute or relative coordinate.
* Summon additional doodads at some coordinate.
* Add and remove items from the player's inventory.

## Ideas for Doodads

Some specific ideas for doodads that should be in the maze game, and what
sorts of scripting features they might need:

* Items (class)
  * A class of doodad that is "picked up" when touched by the player
    character and placed into their inventory.
  * Scriptable hooks can still apply, callback ideas:
    * On enter inventory
    * On leave inventory
  * Example: Gravity Boots could be scripted to invert the global gravity
    when the item enters your inventory until you drop the boots.
  * Some attribute ideas:
    * Undroppable: player can't remove the item from their inventory.
  * Item ideas to start with:
    * Keys to open doors (these would just be mere collectables)
    * Antigravity Boots (scripted to mess with gravity)
* Buttons
  * At least 2 frames: pressed and not pressed.
  * Needs to associate with a door or something that works with buttons.
  * On collision with a doodad or player character: send a notification to
    its associated Door that it should open. (`Powered: true`)
  * When collision ends, button and its door become unpowered.
* Sticky Buttons
  * Buttons that only become `true` once. They stick "On" when activated
    for the first time.
  * Once pressed they can't be unpressed. However, there's nothing stopping
    a switch from targeting a sticky button, so when the switch is turned off
    the sticky button turns off too.
* Switches
  * Like a button. On=`true` and Off=`false`
  * 2 frames for the On and Off position.
  * On "use" by the player, toggle the switch and notify the door of the new
    boolean value.
    * It would invert the value of the target, not just make it match the
      value of the switch. i.e. if the switch is `false` and the door is
      already open (`true`), making the switch `true` closes the door.
* Powered Doors
  * Can only be opened when powered.
  * 2 frames of animation: open and closed.
  * A switch or button must target the door as a way to open/close it.
* Locked Doors
  * Requires a key item to be in the player's inventory.
  * On collision with the player: if they have the key, the door toggles to
    its `true` powered state (open) and stays open.
  * The door takes the key from the player's inventory when opened.
* Trapdoors
  * One-way doors that close behind you.
  * Can be placed horizontally: a doodad falling from above should cause
    the door to swing open (provided it's a downward-only door) and fall
    through.
  * Can be placed vertically and acts as a one-way door.
  * Needs several frames of animation.
