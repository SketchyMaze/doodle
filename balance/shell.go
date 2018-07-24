package balance

import "git.kirsle.net/apps/doodle/render"

// Shell related variables.
var (
	// TODO: why not renders transparent
	ShellBackgroundColor         = render.Color{0, 10, 20, 128}
	ShellForegroundColor         = render.White
	ShellPadding          int32  = 8
	ShellFontSize                = 16
	ShellCursorBlinkRate  uint64 = 20
	ShellHistoryLineCount        = 8

	// Ticks that a flashed message persists for.
	FlashTTL uint64 = 400
)
