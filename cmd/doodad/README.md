# doodad.exe

The `doodad` tool is a command line interface for interacting with Levels and Doodad files, collectively referred to as "Doodle drawings" or just "drawings" for short. It provides many useful features for custom content creators that are not available in the game's user interface.

- [doodad.exe](#doodadexe)
- [Quick Start](#quick-start)
  - [Creating a custom doodad](#creating-a-custom-doodad)
  - [Creating a custom Level Pack](#creating-a-custom-level-pack)
- [Usage](#usage)
- [Features in Depth](#features-in-depth)
  - [$ `doodad convert`: to and from image files](#-doodad-convert-to-and-from-image-files)
  - [$ `doodad show`: Get information about a level or doodad](#-doodad-show-get-information-about-a-level-or-doodad)
  - [Editing Level or Doodad Properties](#editing-level-or-doodad-properties)
- [Where to Find It](#where-to-find-it)

# Quick Start

## Creating a custom doodad

This is an example how to fully create a custom doodad using PNG images for their sprites. The game's built-in doodads are all built in this way:

```bash
# Create the initial doodad from PNG images making up its layers.
% doodad convert --title "Bird (red)" --author "Myself" \
    left-1.png left-2.png right-1.png right-2.png \
    bird-red.doodad

# Attach my custom JavaScript to program the doodad.
% doodad install-script bird.js bird-red.doodad

# Note: `doodad show --script` can get the script back out.

# Set tags and options on my doodad
% doodad edit-doodad --tag "category=creatures" \
    --tag "color=blue" \
    --option "No A.I.=bool" \
    bird-red.doodad

# See the Guidebook for more information!
```

## Creating a custom Level Pack

The doodad tool is currently the best way to create a custom Level Pack for the game. Level Packs will appear in the Story Mode selector, and you can install local levelpacks by putting them in that folder of your Profile Directory.

```bash
# First Quest
doodad levelpack create -t "First Quest" -d "The first story mode campaign." \
	-a "$AUTHOR" --doodads none --free 1 \
	"levelpacks/builtin-100-FirstQuest.levelpack" \
	"levels/Castle.level" \
	"levels/Boat.level" \
	"levels/Jungle.level" \
	"levels/Thief 1.level" \
	"levels/Desert-1of2.level" \
	"levels/Desert-2of2.level" \
	"levels/Shapeshifter.level"
```

Some useful options you can set on your levelpack:

* Title: defaults to your first level's title.
* Description: for the Story Mode picker.
* Free levels: if you want progressive unlocking of levels, specify how many levels are unlocked to start with (at least 1). Otherwise all levels are unlocked.
* Doodads: by default any custom doodad will be bundled with the levelpack.
  * Options for `--doodads` are `none`, `custom` (default), and `all`
  * Use `none` if your levelpack uses _only_ built-in doodads.

# Usage

See `doodad --help` for documentation of the available commands and their options; a recent example of which is included here.

```
NAME:
   doodad - command line interface for Doodle

USAGE:
   doodad [global options] command [command options]

VERSION:
   v0.14.1 (open source) build N/A. Built on 2024-05-24T19:23:33-07:00

COMMANDS:
   convert         convert between images and Doodle drawing files
   edit-doodad     update metadata for a Doodad file
   edit-level      update metadata for a Level file
   install-script  install the JavaScript source to a doodad
   levelpack       create and manage .levelpack archives
   resave          load and re-save a level or doodad file to migrate to newer file format versions
   show            show information about a level or doodad file
   help, h         Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --debug        enable debug level logging (default: false)
   --help, -h     show help
   --version, -v  print the version
```

The `--help` or `-h` flag can be given to subcommands too, which often have their own commands and flags. For example, `doodad convert -h`

# Features in Depth

## $ `doodad convert`: to and from image files

This command can convert between image files (BMP or PNG) and Doodle drawings (levels and doodads) in both directions.

For custom doodad authors this means you can draw your sprites in a much more capable image editing tool, and create a doodad from PNG images. The game's built-in doodads are all created in this way.

This command can also be used to "export" a drawing as a PNG image: when run against a Level file, it may export a massive PNG image containing the entire level! Basically the "Giant Screenshot" function from the level editor.

It can also create a Level from a screenshot, generating the Palette from the distinct colors found (for you to assign properties to in the editor, for e.g. solid geometry).

Examples:

```bash
# Export a full screenshot of your level
$ doodad convert mymap.level screenshot.png

# Create a new level based from a PNG image.
$ doodad convert scanned-drawing.png new-level.level

# Create a doodad from a series of PNG images as its layers.
$ doodad convert -t "My Button" up.png down.png my-button.doodad

# Create a new doodad based from a BMP image, and in this image the chroma
# color (transparent) is #FF00FF instead of white as default.
$ doodad convert --key '#FF00FF' button.png button.doodad
```

Supported image types to convert to or from:

* PNG (8-bit or 24-bit, with transparent pixels or chroma)
* BMP (bitmap image with chroma key)

The chrome key defaults to white (`#FFFFFF`), so pixels of that color are treated as transparent and ignored. For PNG images, if a pixel is fully transparent (alpha channel 0%) it will also be skipped. So the easiest format to use are PNG images with transparent pixels.

When converting an image into a Level or Doodad, the unique colors identified in the drawing are extracted to create the **Palette**. You will need to later edit the palette to assign meaning to the colors (giving them names or level properties). In particular, when importing a Level you'd want to mark which colors are solid ground.

## $ `doodad show`: Get information about a level or doodad

The `doodad show` command can return metadata and debug the contents of a Level or Doodad file.

```
% doodad show Example.level
===== Level: Example.level =====
Headers:
   File format: zipfile
  File version: 1
  Game version: 0.14.0
    Level UUID: d136fef3-1a3a-453c-b616-05aca8dd6840
   Level title: Lesson 1: Controls
        Author: Noah P
      Password: 
        Locked: false

Game Rules:
  Difficulty: Normal (0)
    Survival: false

Palette:
  - Swatch name: solid
    Attributes:  solid
    Color:       #777777
  - Swatch name: decoration
    Attributes:  none
    Color:       #ff66ff
  - Swatch name: fire
    Attributes:  fire,water
    Color:       #ff0000
  - Swatch name: semisolid
    Attributes:  semi-solid
    Color:       #cccccc

Level Settings:
  Page type: Bounded
   Max size: 2550x3300
  Wallpaper: notebook.png

Attached Files:
  assets/screenshots/large.png: 63377 bytes
  assets/screenshots/medium.png: 25807 bytes
  assets/screenshots/small.png: 6837 bytes
  assets/screenshots/tiny.png: 7414 bytes

Actors:
  Level contains 16 actors
  Use -actors or -verbose to serialize Actors

Chunks:
  Pixels Per Chunk: 128^2
  Number Generated: 102
  Coordinate Range: (0,0) ... (1919,2047)
  World Dimensions: 1919x2047
  Use -chunks or -verbose to serialize Chunks
```

Or for a doodad:

```
% doodad show example.doodad
===== Doodad: example.doodad =====
Headers:
   File format: zipfile
  File version: 1
  Game version: 0.14.1
  Doodad title: Fire Region
        Author: Noah
    Dimensions: Rect<0,0,128,128>
        Hitbox: Rect<0,0,128,128>
        Locked: true
        Hidden: false
   Script size: 378 bytes

Tags:
  category: technical

Options:
   str name = fire

Palette:
  - Swatch name: Color<#8a2b2b+ff>
    Attributes:  solid
    Color:       #8a2b2b

Layer 0: fire-128
Chunks:
  Pixels Per Chunk: 128^2
  Number Generated: 1
  Coordinate Range: (0,0) ... (127,127)
  World Dimensions: 127x127
  Use -chunks or -verbose to serialize Chunks
```

## Editing Level or Doodad Properties

The `edit-doodad` and `edit-level` subcommands allow setting properties on your custom files programmatically.

Properties you can set for both file types include:

* Metadata like the Title and Author name
* Lock your drawing from being edited in-game

For Doodads, you can set their Tags and Options, Hitbox, etc. - most of the useful settings are supported, as the game's built-in doodads use this program!

# Where to Find It

The `doodad` tool ships with the official releases of the game, and may be found in one of the following places:

* **Windows:** doodad.exe should be in the same place as sketchymaze.exe (e.g. in the ZIP file)
* **Mac OS:** it is available inside the "Sketchy Maze.app" bundle, in the "Contents/MacOS" folder next to the `sketchymaze` binary.

    Invoke it from a terminal like:

    ```bash
    alias doodad="/Applications/Sketchy Maze.app/Contents/MacOS/doodad"
    doodad -h
    ```
* **Linux:** the doodad binary should be in the same place as the sketchymaze program itself:
    * In the .tar.gz file
    * In /opt/sketchymaze if installed by an .rpm or .deb package
    * AppImage: `./SketchyMaze.AppImage doodad` will invoke the doodad command.
    * Flatpak: `flatpak run com.sketchymaze.Doodle doodad` will invoke the doodad command. Invoke it from a terminal like:

        ```bash
        alias doodad="flatpak run com.sketchymaze.Doodle doodad"
        doodad -h
        ```