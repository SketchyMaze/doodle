package doodle

import (
	"fmt"

	"git.kirsle.net/apps/doodle/pkg/balance"
	"git.kirsle.net/apps/doodle/pkg/log"
	"git.kirsle.net/go/ui"
)

// Width of the panel frame.
var paletteWidth = 50

// SetupPalette sets up the palette panel.
func (u *EditorUI) SetupPalette(d *Doodle) *ui.Window {
	window := ui.NewWindow("Palette")
	window.ConfigureTitle(balance.TitleConfig)
	_, label := window.TitleBar()
	label.Font = balance.TitleFont
	window.Configure(ui.Config{
		Background:  balance.WindowBackground,
		BorderColor: balance.WindowBorder,
	})

	// Color Palette Frame.
	u.PaletteTab = u.setupPaletteFrame(window)
	window.Pack(u.PaletteTab, ui.Pack{
		Side: ui.N,
		Fill: true,
	})

	return window
}

// setupPaletteFrame configures the Color Palette tab for Edit Mode.
// This is a subroutine of editor_ui.go#SetupPalette()
func (u *EditorUI) setupPaletteFrame(window *ui.Window) *ui.Frame {
	frame := ui.NewFrame("Palette Tab")
	frame.SetBackground(balance.WindowBackground)

	var (
		packAlign  = ui.N
		packConfig = ui.Pack{
			Side: packAlign,
			Fill: true,
			PadY: 4,
		}
		tooltipEdge = ui.Left
		buttonSize  = 32
	)
	if balance.HorizontalToolbars {
		packAlign = ui.W
		packConfig = ui.Pack{
			Side: packAlign,
			Fill: true,
			PadX: 2,
		}
		tooltipEdge = ui.Top
		buttonSize = 24
	}

	// Handler function for the radio buttons being clicked.
	onClick := func(ed ui.EventData) error {
		name := u.selectedSwatch
		swatch, ok := u.Canvas.Palette.Get(name)
		if !ok {
			log.Error("Palette onClick: couldn't get swatch named '%s' from palette", name)
			return nil
		}
		log.Info("Set swatch: %s", swatch)
		u.Canvas.SetSwatch(swatch)
		return nil
	}

	// Draw the radio buttons for the palette.
	if u.Canvas != nil && u.Canvas.Palette != nil {
		for _, swatch := range u.Canvas.Palette.Swatches {
			swFrame := ui.NewFrame(fmt.Sprintf("Swatch(%s) Button Frame", swatch.Name))
			swFrame.Configure(ui.Config{
				Width:      buttonSize,
				Height:     buttonSize,
				Background: swatch.Color,
			})

			btn := ui.NewRadioButton("palette", &u.selectedSwatch, swatch.Name, swFrame)
			btn.Handle(ui.Click, onClick)
			u.Supervisor.Add(btn)

			// Add a tooltip showing the swatch attributes.
			ui.NewTooltip(btn, ui.Tooltip{
				Text: fmt.Sprintf("Name: %s\nAttributes: %s", swatch.Name, swatch.Attributes()),
				Edge: tooltipEdge,
			})

			btn.Compute(u.d.Engine)

			frame.Pack(btn, packConfig)
		}
	}

	// Draw the Edit Palette button.
	btn := ui.NewButton("Edit Palette", ui.NewLabel(ui.Label{
		Text: "Edit",
		Font: balance.MenuFont,
	}))
	btn.Handle(ui.Click, func(ed ui.EventData) error {
		u.OpenPaletteWindow()
		return nil
	})
	u.Supervisor.Add(btn)

	btn.Compute(u.d.Engine)
	frame.Pack(btn, packConfig)

	return frame
}
