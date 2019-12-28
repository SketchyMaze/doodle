package doodle

import (
	"fmt"

	"git.kirsle.net/go/render"
	"git.kirsle.net/go/ui"
	"git.kirsle.net/apps/doodle/pkg/balance"
	"git.kirsle.net/apps/doodle/pkg/log"
)

// SetupPalette sets up the palette panel.
func (u *EditorUI) SetupPalette(d *Doodle) *ui.Window {
	window := ui.NewWindow("Palette")
	window.ConfigureTitle(balance.TitleConfig)
	window.TitleBar().Font = balance.TitleFont
	window.Configure(ui.Config{
		Background:  balance.WindowBackground,
		BorderColor: balance.WindowBorder,
	})

	// Doodad frame.
	{
		frame, err := u.setupDoodadFrame(d.Engine, window)
		if err != nil {
			d.Flash(err.Error())
		}

		// Even if there was an error (userdir.ListDoodads couldn't read the
		// config folder on disk or whatever) the Frame is still valid but
		// empty, which is still the intended behavior.
		u.DoodadTab = frame
		u.DoodadTab.Hide()
		window.Pack(u.DoodadTab, ui.Pack{
			Anchor: ui.N,
			Fill:   true,
		})
	}

	// Color Palette Frame.
	u.PaletteTab = u.setupPaletteFrame(window)
	window.Pack(u.PaletteTab, ui.Pack{
		Anchor: ui.N,
		Fill:   true,
	})

	return window
}

// setupPaletteFrame configures the Color Palette tab for Edit Mode.
// This is a subroutine of editor_ui.go#SetupPalette()
func (u *EditorUI) setupPaletteFrame(window *ui.Window) *ui.Frame {
	frame := ui.NewFrame("Palette Tab")
	frame.SetBackground(balance.WindowBackground)

	// Handler function for the radio buttons being clicked.
	onClick := func(p render.Point) {
		name := u.selectedSwatch
		swatch, ok := u.Canvas.Palette.Get(name)
		if !ok {
			log.Error("Palette onClick: couldn't get swatch named '%s' from palette", name)
			return
		}
		log.Info("Set swatch: %s", swatch)
		u.Canvas.SetSwatch(swatch)
	}

	// Draw the radio buttons for the palette.
	if u.Canvas != nil && u.Canvas.Palette != nil {
		for _, swatch := range u.Canvas.Palette.Swatches {
			swFrame := ui.NewFrame(fmt.Sprintf("Swatch(%s) Button Frame", swatch.Name))

			colorFrame := ui.NewFrame(fmt.Sprintf("Swatch(%s) Color Box", swatch.Name))
			colorFrame.Configure(ui.Config{
				Width:       16,
				Height:      16,
				Background:  swatch.Color,
				BorderSize:  1,
				BorderStyle: ui.BorderSunken,
			})
			swFrame.Pack(colorFrame, ui.Pack{
				Anchor: ui.W,
			})

			label := ui.NewLabel(ui.Label{
				Text: swatch.Name,
				Font: balance.StatusFont,
			})
			label.Font.Color = swatch.Color.Darken(128)
			swFrame.Pack(label, ui.Pack{
				Anchor: ui.W,
			})

			btn := ui.NewRadioButton("palette", &u.selectedSwatch, swatch.Name, swFrame)
			btn.Handle(ui.Click, onClick)
			u.Supervisor.Add(btn)

			btn.Compute(u.d.Engine)
			swFrame.Configure(ui.Config{
				Height: label.Size().H,

				// TODO: magic number, trying to left-align
				// the label by making the frame as wide as possible.
				Width: paletteWidth - 16,
			})

			frame.Pack(btn, ui.Pack{
				Anchor: ui.N,
				Fill:   true,
				PadY:   4,
			})
		}
	}

	return frame
}
