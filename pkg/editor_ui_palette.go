package doodle

import (
	"git.kirsle.net/apps/doodle/lib/render"
	"git.kirsle.net/apps/doodle/lib/ui"
	"git.kirsle.net/apps/doodle/pkg/balance"
	"git.kirsle.net/apps/doodle/pkg/log"
)

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
			label.Font.Color = swatch.Color.Darken(40)

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
