# balance

Constants and settings for the Doodle app.

## Environment Variables

Some runtime settings can be configured in the environment. Here they are
with their default values.

Most colors work with alpha channels; just provide an 8 hex character code,
like `#FF00FF99` for 153 ($99) on the alpha channel.

* Application Windw Size (ints):
  * `DOODLE_W=1024`
  * `DOODLE_H=768`
* Shell settings:
  * `D_SHELL_BG=#001428C8`: shell background color.
  * `D_SHELL_FG=#0099FF`: shell text color.
  * `D_SHELL_PC=#FFFFFF`: shell prompt color.
  * `D_SHELL_LN=8`: shell history line count (how tall the shell is in lines)
  * `D_SHELL_FS=16`: font size for both the shell and on-screen flashed
    messages.
* Debug Colors and Hitboxes (default invisible=off):
  * `DEBUG_CHUNK_COLOR=#FFFFFF`: background color when caching a
    chunk to bitmap. Helps visualize where the chunks and caching
    are happening.
  * `DEBUG_CANVAS_BORDER`: draw a border color around every uix.Canvas
    widget. This effectively draws the bounds of every Doodad drawn on top
    of a level or inside a button and the bounds of the level space itself.
* Tuning constants (may not be available in production builds):
  * `D_SCROLL_SPEED=8`: Canvas scroll speed when using the keyboard arrows
    in the Editor Mode, in pixels per tick.
  * `D_DOODAD_SIZE=100`: Default size when creating a new Doodad.
