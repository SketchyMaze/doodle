package balance

import (
	"git.kirsle.net/apps/doodle/lib/render"
	"git.kirsle.net/apps/doodle/lib/ui"
)

// Theme and appearance variables.
var (
	// Window and panel styles.
	TitleConfig = ui.Config{
		Background:   render.MustHexColor("#FF9900"),
		OutlineSize:  1,
		OutlineColor: render.Black,
	}
	TitleFont = render.Text{
		FontFilename: "./fonts/DejaVuSans-Bold.ttf",
		Size:         12,
		Padding:      4,
		Color:        render.White,
		Stroke:       render.Red,
	}
	WindowBackground = render.MustHexColor("#cdb689")
	WindowBorder     = render.Grey

	// Menu bar styles.
	MenuBackground = render.Black
	MenuFont       = render.Text{
		Size: 12,
		PadX: 4,
	}

	// StatusFont is the font for the status bar.
	StatusFont = render.Text{
		Size:    12,
		Padding: 4,
		Color:   render.Black,
	}

	// Color for draggable doodad.
	DragColor = render.MustHexColor("#0099FF")
)
