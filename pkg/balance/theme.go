package balance

import (
	"git.kirsle.net/go/render"
	"git.kirsle.net/go/ui"
	"git.kirsle.net/go/ui/style"
)

// Theme and appearance variables.
var (
	// Sprite filenames.
	WindowIcon = "assets/icons/96.png"
	GoldCoin   = "assets/sprites/gold.png"
	SilverCoin = "assets/sprites/silver.png"
	LockIcon   = "assets/sprites/padlock.png"
	CursorIcon = "assets/sprites/pointer.png"

	// Title Screen Font
	TitleScreenFont = render.Text{
		Size:   46,
		Color:  render.Pink,
		Stroke: render.SkyBlue,
		Shadow: render.Black,
	}
	TitleScreenSubtitleFont = render.Text{
		FontFilename: SansSerifFont,
		Size:         18,
		Color:        render.SkyBlue,
		Shadow:       render.SkyBlue.Darken(128),
		// Color:        render.RGBA(255, 153, 0, 255),
		// Shadow:       render.RGBA(200, 80, 0, 255),
	}
	TitleScreenVersionFont = render.Text{
		Size:   14,
		Color:  render.Grey,
		Shadow: render.Black,
	}

	// Loading Screen fonts.
	LoadScreenFont = render.Text{
		Size:   46,
		Color:  render.Pink,
		Stroke: render.SkyBlue,
		Shadow: render.Black,
	}
	LoadScreenSecondaryFont = render.Text{
		FontFilename: SansSerifFont,
		Size:         18,
		Color:        render.SkyBlue,
		Shadow:       render.SkyBlue.Darken(128),
		// Color:        render.RGBA(255, 153, 0, 255),
		// Shadow:       render.RGBA(200, 80, 0, 255),
	}

	// Play Mode Touch UI Hints Font
	TouchHintsFont = render.Text{
		FontFilename: SansSerifFont,
		Size:         14,
		Color:        render.SkyBlue,
		Shadow:       render.SkyBlue.Darken(128),
		Padding:      8,
		PadY:         12,
	}

	// Window and panel styles.
	TitleConfig = ui.Config{
		Background:   render.MustHexColor("#FF9900"),
		OutlineSize:  1,
		OutlineColor: render.Black,
	}
	TitleFont = render.Text{
		FontFilename: SansBoldFont,
		Size:         9,
		Padding:      4,
		Color:        render.White,
		Stroke:       render.Red,
	}
	WindowBackground = render.MustHexColor("#cdb689")
	WindowBorder     = render.Grey

	// Developer Shell and Flashed Messages styles.
	FlashStrokeDarken = 60
	FlashShadowDarken = 120
	FlashFont         = func(text string) render.Text {
		return render.Text{
			Text:   text,
			Size:   18,
			Color:  render.SkyBlue,
			Stroke: render.SkyBlue.Darken(FlashStrokeDarken),
			Shadow: render.SkyBlue.Darken(FlashShadowDarken),
		}
	}
	FlashErrorColor = render.MustHexColor("#FF9900")

	// Menu bar styles.
	MenuBackground = render.Black
	MenuFont       = render.Text{
		Size: 12,
		PadX: 4,
	}
	MenuFontBold = render.Text{
		FontFilename: SansBoldFont,
		Size:         12,
		PadX:         4,
	}

	// TabFrame styles.
	TabFont = render.Text{
		Size: 12,
		PadX: 8,
		PadY: 4,
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
		FontFilename: SansBoldFont,
		Padding:      4,
		Color:        render.Black,
	}

	// A New Record! Label (Gold/Perfect and Silver/Normal)
	NewRecordPerfectFont = LabelFont.Update(render.Text{
		Color:  render.Yellow,
		Stroke: render.Orange,
	})
	NewRecordFont = LabelFont.Update(render.Text{
		Color:  render.White,
		Stroke: render.Grey,
	})

	LargeLabelFont = render.Text{
		Size:         18,
		FontFilename: SansBoldFont,
		Padding:      4,
		Color:        render.Black,
	}

	// SmallMonoFont for cramped spaces like the +/- buttons on Toolbar.
	SmallMonoFont = render.Text{
		Size:         14,
		PadX:         3,
		FontFilename: MonospaceFont,
		Color:        render.Black,
	}

	// CodeLiteralFont for rendering <code>-like text.
	CodeLiteralFont = render.Text{
		Size:         11,
		PadX:         3,
		FontFilename: MonospaceFont,
		Color:        render.Magenta,
	}

	// Small font
	SmallFont = render.Text{
		Size:    10,
		Padding: 2,
		Color:   render.Black,
	}

	// Color for draggable doodad.
	DragColor = render.MustHexColor("#0099FF")

	// Link lines drawn between connected doodads.
	LinkLineColor        = render.Magenta
	LinkLighten          = 128
	LinkAnimSpeed uint64 = 30 // ticks

	PlayButtonFont = render.Text{
		FontFilename: SansBoldFont,
		Size:         16,
		Padding:      4,
		Color:        render.RGBA(255, 255, 0, 255),
		Stroke:       render.RGBA(100, 100, 0, 255),
	}

	// In-game level timer font.
	TimerFont = render.Text{
		FontFilename: MonospaceFont,
		Size:         16,
		Color:        render.Cyan,
		Stroke:       render.DarkCyan,
	}

	// Doodad Dropper Window settings.
	DoodadButtonBackground = render.RGBA(255, 255, 200, 255)
	DoodadButtonSize       = 64
	DoodadDropperCols      = 6 // rows/columns of buttons
	DoodadDropperRows      = 3

	// Button styles, customized in init().
	ButtonPrimary  = style.DefaultButton
	ButtonDanger   = style.DefaultButton
	ButtonBabyBlue = style.DefaultButton
	ButtonPink     = style.DefaultButton
	ButtonLightRed = style.DefaultButton

	DefaultCrosshairColor = render.RGBA(0, 153, 255, 255)
)

// Customize the various button styles.
func init() {
	// Primary: white on rich blue color
	ButtonPrimary.Background = render.RGBA(0, 60, 153, 255)
	ButtonPrimary.Foreground = render.RGBA(255, 255, 254, 255)
	ButtonPrimary.HoverBackground = render.RGBA(0, 153, 255, 255)
	ButtonPrimary.HoverForeground = ButtonPrimary.Foreground

	// Danger: white on red
	ButtonDanger.Background = render.RGBA(153, 30, 30, 255)
	ButtonDanger.Foreground = render.RGBA(255, 255, 254, 255)
	ButtonDanger.HoverBackground = render.RGBA(255, 30, 30, 255)
	ButtonDanger.HoverForeground = ButtonPrimary.Foreground

	ButtonBabyBlue.Background = render.RGBA(40, 200, 255, 255)
	ButtonBabyBlue.Foreground = render.Black
	ButtonBabyBlue.HoverBackground = render.RGBA(0, 220, 255, 255)
	ButtonBabyBlue.HoverForeground = render.Black

	ButtonPink.Background = render.RGBA(255, 153, 255, 255)
	ButtonPink.HoverBackground = render.RGBA(255, 220, 255, 255)

	ButtonLightRed.Background = render.RGBA(255, 90, 90, 255)
	ButtonLightRed.HoverBackground = render.RGBA(255, 128, 128, 255)
}
