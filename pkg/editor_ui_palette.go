package doodle

import (
	"git.kirsle.net/apps/doodle/lib/render"
	"git.kirsle.net/apps/doodle/lib/ui"
	"git.kirsle.net/apps/doodle/pkg/balance"
	"git.kirsle.net/apps/doodle/pkg/enum"
	"git.kirsle.net/apps/doodle/pkg/log"
	"git.kirsle.net/apps/doodle/pkg/uix"
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

	// Frame that holds the tab buttons in Level Edit mode.
	tabFrame := ui.NewFrame("Palette Tabs")
	for _, name := range []string{"Palette", "Doodads"} {
		if u.paletteTab == "" {
			u.paletteTab = name
		}

		tab := ui.NewRadioButton("Palette Tab", &u.paletteTab, name, ui.NewLabel(ui.Label{
			Text: name,
		}))
		tab.Handle(ui.Click, func(p render.Point) {
			if u.paletteTab == "Palette" {
				u.Canvas.Tool = uix.PencilTool
				u.PaletteTab.Show()
				u.DoodadTab.Hide()
			} else {
				u.Canvas.Tool = uix.ActorTool
				u.PaletteTab.Hide()
				u.DoodadTab.Show()
			}
			window.Compute(d.Engine)
		})
		u.Supervisor.Add(tab)
		tabFrame.Pack(tab, ui.Pack{
			Anchor: ui.W,
			Fill:   true,
			Expand: true,
		})
	}
	window.Pack(tabFrame, ui.Pack{
		Anchor: ui.N,
		Fill:   true,
		PadY:   4,
	})

	// Only show the tab frame in Level drawing mode!
	if u.Scene.DrawingType != enum.LevelDrawing {
		tabFrame.Hide()
	}

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
			label := ui.NewLabel(ui.Label{
				Text: swatch.Name,
				Font: balance.StatusFont,
			})
			label.Font.Color = swatch.Color.Darken(128)

			btn := ui.NewRadioButton("palette", &u.selectedSwatch, swatch.Name, label)
			btn.Handle(ui.Click, onClick)
			u.Supervisor.Add(btn)

			frame.Pack(btn, ui.Pack{
				Anchor: ui.N,
				Fill:   true,
				PadY:   4,
			})
		}
	}

	return frame
}
