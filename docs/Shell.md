# Developer Console

The backtick key (<code>`</code>) toggles open the developer console where commands and cheats may be entered.

The source of truth is in `pkg/shell.go`.

The `help` command provides in-game documentation for its commands.

There is an `eval` or `$` command which allows running arbitrary JavaScript statements in the shell. Several useful game structs are exposed, with all of their Go-exported properties available.

* `d` is the root Doodle object for the game engine.
* `d.Scene` points to the current Scene (e.g. PlayScene or EditScene). Its properties will vary depending on the current scene.
* Several Go type constructors are available:
	* `RGBA(r, g, b, a uint8) render.Color`
	* `Point(x, y int) render.Point`
	* `Vector(x, y float64) physics.Vector`
	* `Rect(width, height int) render.Rect`
* `Tree(ui.Widget)` will print a UI widget tree to the in-game console.
	* Example: `$ Tree(d.Scene.UI.MenuBar)`

## Commands

The source of truth is in `pkg/commands.go`.

* `help` shows the list of commands, `help <command>` to get info about a command.
* `clear` resets the console and erases previous text.
* `echo <text>` to add text to the in-game flashes (blue text).
* `error <text>` to add text to the in-game error flashes (orange text).
* `alert <text>` to trigger an in-game alert box message.
* `confirm <text>` to trigger an Ok/Cancel message (both buttons dismiss the dialog).
* `new` switches to the New Drawing scene.
* `save <filename>` to save your drawing in Edit Mode.
* `edit <filename>` to open the Editor to the level or doodad. The file extension (like `.doodad`) is optional. If the same name exists as both a level and a doodad, the level is assumed first with no file extension is given.
* `play <filename>` to play a level.
* `close` exits to the Main Menu.
* `titlescreen <filename>` exits to the Main Menu but with the specified level playing as the background.
* `exit`, `quit` exits the game.
* `reload` tears down and restarts the current scene.
* `guitest` opens an internal GUI test debugging scene.
* `eval` or `$` to execute a JavaScript statement.
* `repl` to enter an interactive JavaScript REPL (every command is a statement).
* `boolprop <name> [true, false, flip]` to toggle some [Bool Props](#bool-props).
* `extract-bindata <path>` extracts game assets embedded in the binary to a folder on disk. This will extract built-in doodads, levels, wallpapers and other assets from the game binary.
* `throw <text>` to test the JavaScript exception catcher.
	* `throw2` stress tests the exception catcher with a large amount of text.
	* `throw3` tests a fairly standard realistic exception that a doodad script might throw.
* `flush-textures` triggers the render engine to release all memory for stored textures. The game uses Go SDL2 and this will free all texture pointers from the C side. If those textures were still needed, the game will rebuild them next tick from cached image.Image objects stored in native Go (and subject to its garbage collector when they truly have gone out of scope).

## Cheats

The source of truth is in `pkg/balance/cheats.go`.

Many of these cheats are available graphically in-game without needing to enter their commands. The Cheats Menu can be opened from the game's Settings -> Experimental -> "Open Cheats Window" button. On the Misc tab of the Cheats Menu, you can "Enable cheats menu" to make this window easily accessible from the "Help" menu during gameplay.

A ⭐ emoji next to a cheat below means it is available in the Cheats Menu in-game.

General cheats:

* `unleash the beast` - disable frame rate throttling.
* `master key` - unlock all levels.

Play Mode cheats:

* `don't edit and drive` - enable editing while playing a level. Experimental!
* `scroll scroll scroll your boat` - enable scrolling the level with arrow keys while playing a level.
* ⭐ `import antigravity` - disables gravity for the player character and allows free movement in all directions with the arrow keys. Enter the cheat again to restore gravity to normal.
  * Note: under antigravity, hold down the Shift key to lower the player speed to only one pixel per tick.
* ⭐ `ghost mode` - toggles NoClip and antigravity for the player character, enabling them to move through solid obstacles.
* ⭐ `show all actors` - makes all hidden actors visible (e.g. makes technical doodads visible).
* ⭐ `give all keys` - get all 4 colored keys and +99 small keys.
* ⭐ `give all gems` - get all 4 gems.
* ⭐ `drop all items` - clears your inventory.
* ⭐ `god mode` - toggle invincibility.
* ⭐ `warp whistle` - skip the current level.
* `tesla` - broadcasts a power toggle signal to ALL doodads in the level (e.g. opening all electric doors and activating all devices). Enter the cheat again to toggle off/on signals.

Change the default player character for all levels (except ones that mandate a custom character):

* ⭐ `pinocchio`: Boy (default)
* ⭐ `play as thief`: Thief
* ⭐ `the cell`: Blue Azulian
* `super azulian`: Red Azulian
* `hyper azulian`: White Azulian
* ⭐ `fly like a bird`: Red bird
* `bluebird`: Blue bird
* ⭐ `megaton weight`: Anvil

UI test screens:

* `test load screen`
* `test wait screen`

## Bool Props

Some boolean switches can be toggled in the command shell.

Usage: `boolProp <name> <value>`

The value is truthy if its first character is the letter T or the number 1.
All other values are false. Examples: True, true, T, t, 1.

`boolProp list` will show a list of available props.

* `Debug` or `D`: toggle debug mode within the app.
* `DebugOverlay` or `DO`: toggle the debug text overlay (F3 key)
* `DebugCollision` or `DC`: toggle collision hitbox lines (F4 key)

Some more boolprops:

* `eager-render`: DEPRECATED - controlled whether levels would eagerly load/render all chunks.
* `horizontal-toolbars`: make Editor Mode toolbars horizontal instead of vertical. Requires a scene `reload` to take effect.
* `pretty-json`: DEPRECATED - saved the legacy v1 JSON versions of levels/doodads with pretty-printed indentation.
* `show-hidden-doodads`: shows hidden doodads in the Level Editor, such as Boy under the Creatures category. Note: requires a scene `reload` to take effect.
* `write-lock-override`: enables editing a level or doodad which had the write lock flag set (which includes all the game's built-in doodads).

## Interesting Tricks

### Custom Player Doodad

Change the default player doodad for all levels to be any arbitrary doodad in your game:

```
$ d.SetPlayerCharacter("example-mario.doodad")
```

The `.doodad` suffix is optional.

Built-in doodads can be set this way as well.

```
$ d.SetPlayerCharacter("bird-red")
$ d.SetPlayerCharacter("anvil")
$ d.SetPlayerCharacter("azu-blu")
```

Many of their filenames can be found in pkg/balance/cheats.go.

> boy, thief, azu-blu, azu-red, bird-red, bird-blue, crusher, snake, anvil

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

### Spawn Cheats Window

Usable on screens which FindLikelySupervisor() can find one for (Main Menu, Play, Editor).

```
$ d.MakeCheatsWindow( d.FindLikelySupervisor()[0] )
```

### Scene Loaders

| Command                   | Description                                                                            |
| ------------------------- | -------------------------------------------------------------------------------------- |
| `$ d.GotoLoadMenu()`      | DEPRECATED: Old file open screen for levels/doodads.                                   |
| `$ d.GotoPlayMenu()`      | DEPRECATED: Old file open screen to play local levels.                                 |
| `$ d.GotoNewMenu()`       | New Level screen.                                                                      |
| `$ d.GotoNewDoodadMenu()` | New Level screen with Doodad tab selected.                                             |
| `$ d.GotoSettingsMenu()`  | Should show a blank scene with Settings dialog, but crashes the game instead.          |
| `$ d.NewDoodad(128, 128)` | Go immediately into Edit Mode for a new doodad with the given width,height configured. |
| `$ d.NewMap()`            | Go immediately into Edit Mode for a new level with default settings.                   |

### Window Size Jank

```
$ d.SetWindowSize(800, 600)
```

This won't change the actual window size on screen, but changes the game's internal idea of how big the window should be. May require a scene reload for any effects to be visible. Run the `reload` command after changing the window size.

If you set the window size to significantly smaller than the true window size, then the Editor/Play scenes will draw their UI smaller in the top-left corner of the window. Scrolling behavior during gameplay may be buggy. You can see the chunk loading/unloading 'out of bounds' action this way.

To reset, just resize the actual game window and the game should update to match.

### Title Screen (MainScene)

Commands that work only from the title screen:

* `$ d.Scene.PauseLazyScroll = true` to stop the automatic scroll/pan of the background level.
* `$ d.Scene.LabelVersion().Hide()` to hide the version number.
* `$ d.Scene.LabelHint().Hide()` to hide the "Hint: press the Arrow Keys" label.
* `$ d.Scene.ButtonFrame().Hide()` to hide the action buttons.
* `$ d.Scene.MakePhotogenic(true)` pauses lazy scroll and hides UI elements except the title and version number, for photogenic screenshots.

### Play Mode (PlayScene)

During Play Mode:

* `$ d.Scene.BeatLevel()` to clear the current level. Note: any script command entered during gameplay marks your session as cheated so it won't log a high score.
* `$ d.Scene.DieByFire("name")` fails the level with "Watch out for name!"
* `$ d.Scene.ResetTimer()` resets the level timer to zero. Note: your session will be flagged as cheated by running any of these commands!
* `$ d.Scene.RestartLevel()`
* `$ d.Scene.RetryCheckpoint()`
* `$ d.Scene.SetCheckpoint( Point(5, 5) )` to move your respawn point.
* `$ d.Scene.ShowEndLevelModal(false, "title", "message")` to fail the level with a custom title and message.
* `$ d.Scene.ShowEndLevelModal(true, "title", "message")` to win the level with a custom title and message.