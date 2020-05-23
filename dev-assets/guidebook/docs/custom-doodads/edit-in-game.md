# Drawing a Doodad In-Game

Project: Doodle has some **limited** support to draw your doodad sprites
in-game. Currently you can only draw one frame (image) for the doodad
and save it to disk.

To start a new doodad, open the game and enter the level editor.

Select the "New Doodad" button at the top of the screen to start drawing a
new doodad. Choose the size (square) of its sprite when prompted.

Doodads saved in-game go in your user config directory for the game. On Linux,
this is at ~/.config/doodle.

If you want to create a doodad with multiple frames (to animate it or have
varying states that change the doodad's appearance in the level), the
`doodad` tool is recommended. See
[drawing images in an external program](edit-external.md).

## Future Planned Features

Creating doodads in-game is intended to be a fully supported feature. The
following features are planned to be supported:

* Support editing multiple frames instead of only the first frame.
* Implement some features only available on the `doodad` tool using in-game
  UI, such as attaching JavaScripts to the doodad.
