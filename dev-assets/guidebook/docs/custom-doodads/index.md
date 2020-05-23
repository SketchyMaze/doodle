# Creating Custom Doodads

Project: Doodle is designed to be modder friendly and provides tools to help
you create your own custom doodads to use in your levels.

You can draw the sprites for the doodad either in-game or using an external
image editor. Then, you can program their logic using JavaScript to make them
"do" stuff in-game and interact with the player and other doodads.

* Drawing your Doodad's Sprites
    * [In-Game](edit-in-game.md)
    * [In an External Program](edit-external.md)
* Program its Behavior
    * [JavaScript](scripts.md)

## doodad (Command Line Tool)

Your copy of the game should have shipped with a `doodad` command-line tool
bundled with it. On Windows it's called `doodad.exe` and should be in the same
folder as the game executable. On Mac OS, it is inside the .app bundle.

The `doodad` tool provides a command-line interface to create and inspect
doodad and level files from the game. You'll need to use this tool, at the very
least, to attach a JavaScript to your doodad to make it "do" stuff in-game.

You can create a doodad from PNG images on disk, attach or view the JavaScript
source on them, and view/edit metadata.

```bash
# (the $ represents the shell prompt in a command-line terminal)

# See metadata about a doodad file.
$ doodad show /path/to/custom.doodad

# Create a new doodad based on PNG images on disk.
$ doodad convert frame0.png frame1.png frame2.png output.doodad

# Add and view a custom script attached to the doodad.
$ doodad install-script index.js custom.doodad
$ doodad show --script custom.doodad
```

More info on the [`doodad` tool](../doodad-tool.md) here.
