# doodad.exe

The doodad tool is a command line interface for interacting with Levels and
Doodad files, collectively referred to as "Doodle drawings" or just "drawings"
for short.

# Commands

## doodad convert

Convert between standard image files (bitmap or PNG) and Doodle drawings
(levels or doodads).

This command can be used to "export" a Doodle drawing as a PNG (when run against
a Level file, it may export a massive PNG image containing the entire level).
It may also "import" a new Doodle drawing from an image on disk.

Example:

```bash
# Export a full screenshot of your level
$ doodad convert mymap.level screenshot.png

# Create a new level based from a PNG image.
$ doodad convert scanned-drawing.png new-level.level

# Create a new doodad based from a BMP image, and in this image the chroma
# color (transparent) is #FF00FF instead of white as default.
$ doodad convert --key '#FF00FF' button.png button.doodad
```

Supported image types:

* PNG (8-bit or 24-bit, with transparent pixels or chroma key)
* BMP (bitmap image with chroma key)

The chrome key defaults to white (`#FFFFFF`), so pixels of that color are
treated as transparent and ignored. For PNG images, if a pixel is fully
transparent (alpha channel 0%) it will also be skipped.

When converting an image into a drawing, the unique colors identified in the
drawing are extracted into the palette. You will need to later edit the palette
to assign meaning to the colors.
