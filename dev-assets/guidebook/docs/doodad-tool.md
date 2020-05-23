# Doodad Tool

The game ships with a command-line program called `doodad` which assists in
creating and managing custom doodads and levels.

The `doodad` tool can show and set details on .doodad and .level files used by
the game, create new doodads from PNG images and attach custom JavaScript source
to program behavior of doodads.

## Where to Find It

The `doodad` tool should be in the same place as the game executable.

On Windows, the program is called `doodad.exe` and comes in the zip file next
to the game executable, `doodle.exe`.

On Linux, it will typically be at `/opt/project-doodle/doodad`.

On Mac OS, it is found inside the .app bundle.

## Usage

Run `doodad --help` to get usage information.

The program includes several sub-commands, such as `doodad convert`. Type a
subcommand and `--help` to get help on that command, for example:

```bash
doodad convert --help
```

# Examples

Here are some common scenarios and use cases for the doodad tool.

## Show

```bash
# Usage:
doodad show [doodad or level filename]
```

Shows metadata and details about a level or doodad file.

Examples:

```bash
$ doodad show button.doodad
===== Doodad: button.doodad =====
Headers:
  File version: 1
  Game version: 0.0.10-alpha
  Doodad title: Button
        Author: Noah
        Locked: true
        Hidden: false
   Script size: 473 bytes

Palette:
  - Swatch name: Color<#000000+ff>
    Attributes:  solid
    Color:       #000000
  - Swatch name: Color<#666666+ff>
    Attributes:  none
    Color:       #666666
  - Swatch name: Color<#999999+ff>
    Attributes:  fire
    Color:       #999999

Layer 0: button1
Chunks:
  Pixels Per Chunk: 37^2
  Number Generated: 1
  Coordinate Range: (0,0) ... (36,36)
  World Dimensions: 36x36
  Use -chunks or -verbose to serialize Chunks

Layer 1: button2
Chunks:
  Pixels Per Chunk: 37^2
  Number Generated: 1
  Coordinate Range: (0,0) ... (36,36)
  World Dimensions: 36x36
  Use -chunks or -verbose to serialize Chunks
```

## Convert

```bash
# Usage:
doodad convert [options] <input files.png> <output file.doodad>
```

### Creating a Doodad from PNG images

Suppose you have PNG images named "frame0.png" through "frame3.png" and want
to create a doodad from those images. This will convert them to the doodad
file "custom.doodad":

```bash
# Convert PNG images into a doodad.
doodad convert frame0.png frame1.png frame2.png frame3.png custom.doodad

# The same, but also attach custom tags with the doodad.
doodad convert --tag color=blue frame{0,1,2,3}.png custom.doodad
```

### Convert a level to a PNG image

```bash
doodad convert my.level output.png
```

### Create a level from a PNG image

```bash
doodad convert level.png output.level
```
