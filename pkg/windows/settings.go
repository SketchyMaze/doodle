package windows

import (
	"fmt"

	"git.kirsle.net/apps/doodle/pkg/balance"
	"git.kirsle.net/apps/doodle/pkg/branding"
	"git.kirsle.net/apps/doodle/pkg/native"
	"git.kirsle.net/go/render"
	"git.kirsle.net/go/ui"
)

// Settings window.
type Settings struct {
	// Settings passed in by doodle
	Supervisor *ui.Supervisor
	Engine     render.Engine
}

// NewSettingsWindow initializes the window.
func NewSettingsWindow(cfg Settings) *ui.Window {
	window := ui.NewWindow("Game Settings")
	window.SetButtons(ui.CloseButton)
	window.Configure(ui.Config{
		Width:      400,
		Height:     170,
		Background: render.Grey,
	})

	// {
	// 	row := ui.NewFrame("Theme Frame")
	// 	label := ui.NewLabel(ui.Label{
	// 		Text: "Theme:",
	// 		Font: balance.MenuFont,
	// 	})
	// }

	text := ui.NewLabel(ui.Label{
		Text: fmt.Sprintf("%s is a drawing-based maze game.\n\n"+
			"Copyright © %s.\nAll rights reserved.\n\n"+
			"Version %s",
			branding.AppName,
			branding.Copyright,
			branding.Version,
		),
	})
	window.Pack(text, ui.Pack{
		Side:    ui.N,
		Padding: 8,
	})

	frame := ui.NewFrame("Button frame")
	buttons := []struct {
		label string
		f     func()
	}{
		{"Website", func() {
			native.OpenURL(branding.Website)
		}},
		{"Open Source Licenses", func() {
			// TODO: open file
			native.OpenURL("./Open Source Licenses.md")
		}},
	}
	for _, button := range buttons {
		button := button

		btn := ui.NewButton(button.label, ui.NewLabel(ui.Label{
			Text: button.label,
			Font: balance.MenuFont,
		}))

		btn.Handle(ui.Click, func(ed ui.EventData) error {
			button.f()
			return nil
		})

		btn.Compute(cfg.Engine)
		cfg.Supervisor.Add(btn)

		frame.Pack(btn, ui.Pack{
			Side:   ui.W,
			PadX:   4,
			Expand: true,
			Fill:   true,
		})
	}
	window.Pack(frame, ui.Pack{
		Side:    ui.N,
		Padding: 8,
	})

	return window
}
