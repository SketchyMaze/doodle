# Drawing Doodads in an External Program

Doodad sprites can be drawn using any image editor and saved as .png files
(with transparency). You can then create a doodad file from your series of
PNG images.

Most of the built-in doodads that ship with the game were created in this way.

## Create a Doodad from Images

Save each frame of your doodad as a separate PNG image and then use the `doodad`
command-line tool to convert them to a `.doodad` file.

```bash
# Usage:
doodad convert [options] <inputs.png> <output.doodad>

# Example:
doodad convert door-closed.png door-open.png door.doodad
```
