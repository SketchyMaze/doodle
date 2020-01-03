# Developer Console

## Cheats

* `unleash the beast` - disable frame rate throttling.
* `don't edit and drive` - enable editing while playing a level.
* `scroll scroll scroll your boat` - enable scrolling the level with arrow keys
  while playing a level.
* `import antigravity` - during Play Mode, disables gravity for the player
  character and allows free movement in all directions with the arrow keys.
  Enter the cheat again to restore gravity to normal.
  * Note: under antigravity, hold down the Shift key to lower the player
    speed to only one pixel per tick.
* `ghost mode` - during Play Mode, toggles noclip for the player character.

## Bool Props

Some boolean switches can be toggled in the command shell.

Usage: `boolProp <name> <value>`

The value is truthy if its first character is the letter T or the number 1.
All other values are false. Examples: True, true, T, t, 1.

* `Debug` or `D`: toggle debug mode within the app.
* `DebugOverlay` or `DO`: toggle the debug text overlay.
* `DebugCollision` or `DC`: toggle collision hitbox lines.

## Interesting Tricks

### Editable Map While Playing

In Play Mode run the command:

| Command                                    | Effect                                                         |
|--------------------------------------------|----------------------------------------------------------------|
| `$ d.Scene.Drawing().Editable = true`      | Can click and drag new pixels onto the level while playing it. |
| `$ d.Scene.Drawing().Scrollable = true`    | Arrow keys scroll the map, like in editor mode.                |
| `$ d.Scene.Drawing().NoLimitScroll = true` | Allow map to scroll beyond bounded limits.                     |

The equivalent Canvas in the Edit Mode is at `d.Scene.UI.Canvas`

### Edit Out-of-Bounds in Editor Mode

In Edit Mode run the command:

`$ d.Scene.UI.Canvas.NoLimitScroll = true`

and you can scroll the map freely outside of the normal scroll boundaries. For
example, to see/edit pixels outside the top-left edges of bounded levels.
