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
		FontFilename: "DejaVuSans-Bold.ttf",
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

	// UIFont is the main font for UI labels.
	UIFont = render.Text{
		Size:    12,
		Padding: 4,
		Color:   render.Black,
	}

	// LabelFont is the font for strong labels in UI.
	LabelFont = render.Text{
		Size:         12,
		FontFilename: "DejaVuSans-Bold.ttf",
		Padding:      4,
		Color:        render.Black,
	}

	// Color for draggable doodad.
	DragColor = render.MustHexColor("#0099FF")

	// Link lines drawn between connected doodads.
	LinkLineColor        = render.Magenta
	LinkLighten          = 128
	LinkAnimSpeed uint64 = 30 // ticks

	PlayButtonFont = render.Text{
		FontFilename: "DejaVuSans-Bold.ttf",
		Size:         16,
		Padding:      4,
		Color:        render.RGBA(255, 255, 0, 255),
		Stroke:       render.RGBA(100, 100, 0, 255),
	}
)
