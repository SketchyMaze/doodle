package balance

import (
	"git.kirsle.net/go/render"
	"git.kirsle.net/go/ui"
	"git.kirsle.net/go/ui/style"
)

// Theme and appearance variables.
var (
	// Title Screen Font
	TitleScreenFont = render.Text{
		Size:   46,
		Color:  render.Pink,
		Stroke: render.SkyBlue,
		Shadow: render.Black,
	}

	// Window and panel styles.
	TitleConfig = ui.Config{
		Background:   render.MustHexColor("#FF9900"),
		OutlineSize:  1,
		OutlineColor: render.Black,
	}
	TitleFont = render.Text{
		FontFilename: "DejaVuSans-Bold.ttf",
		Size:         9,
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
	MenuFontBold = render.Text{
		FontFilename: "DejaVuSans-Bold.ttf",
		Size:         12,
		PadX:         4,
	}

	// Modal backdrop color.
	ModalBackdrop = render.RGBA(1, 1, 1, 42)

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

	// SmallMonoFont for cramped spaces like the +/- buttons on Toolbar.
	SmallMonoFont = render.Text{
		Size:         14,
		PadX:         3,
		FontFilename: "DejaVuSansMono.ttf",
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

	// Doodad Dropper Window settings.
	DoodadButtonBackground = render.RGBA(255, 255, 200, 255)
	DoodadButtonSize       = 64
	DoodadDropperCols      = 6 // rows/columns of buttons
	DoodadDropperRows      = 3

	// Button styles, customized in init().
	ButtonPrimary = style.DefaultButton
)

func init() {
	// Customize button styles.
	ButtonPrimary.Background = render.RGBA(0, 60, 153, 255)
	ButtonPrimary.Foreground = render.RGBA(255, 255, 254, 255)
	ButtonPrimary.HoverBackground = render.RGBA(0, 153, 255, 255)
	ButtonPrimary.HoverForeground = ButtonPrimary.Foreground
}
