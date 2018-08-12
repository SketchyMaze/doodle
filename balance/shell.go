package balance

import (
	"git.kirsle.net/apps/doodle/render"
)

// Shell related variables.
var (
	// TODO: why not renders transparent
	ShellFontFilename            = "./fonts/DejaVuSansMono.ttf"
	ShellBackgroundColor         = render.RGBA(0, 20, 40, 200)
	ShellForegroundColor         = render.RGBA(0, 153, 255, 255)
	ShellPromptColor             = render.White
	ShellPadding          int32  = 8
	ShellFontSize                = 16
	ShellCursorBlinkRate  uint64 = 20
	ShellHistoryLineCount        = 8

	// Ticks that a flashed message persists for.
	FlashTTL uint64 = 400
)
