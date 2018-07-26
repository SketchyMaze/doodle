# Debug Notes

## Entering Debug Mode

Command line argument:

```bash
% doodle -debug

# running the dev build uses debug mode by default
% make run
```

In the developer console:

```dos
> boolProp Debug true
> boolProp D true
```

## Debug Options

The `boolProp` command can also be used to toggle on and off different
debug options while the game is running.

```
DebugOverlay
DO
  Toggles the main debug text overlay with FPS counter.

DebugCollision
DC
  Toggles the collision detection bounding box lines.
```

## JavaScript Shell

The developer console can parse JavaScript commands for more access to the
game's internal objects.

The following global variables are available to the shell:

* `d` is the master Doodle struct.
* `log` is the master logger object for logging messages to the terminal.
* `RGBA()` is the `render.RGBA()` function for creating a Color value.
* `Point(x, y)` to create a `render.Point`
