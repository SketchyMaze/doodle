package doodle

import (
	"fmt"

	"git.kirsle.net/apps/doodle/balance"
	"git.kirsle.net/apps/doodle/events"
	"git.kirsle.net/apps/doodle/render"
	"git.kirsle.net/apps/doodle/ui"
)

// EditorUI manages the user interface for the Editor Scene.
type EditorUI struct {
	d     *Doodle
	Scene *EditorScene

	// Variables
	StatusMouseText    string
	StatusPaletteText  string
	StatusFilenameText string

	// Widgets
	Supervisor *ui.Supervisor
	Palette    *ui.Window
	StatusBar  *ui.Frame
}

// NewEditorUI initializes the Editor UI.
func NewEditorUI(d *Doodle, s *EditorScene) *EditorUI {
	u := &EditorUI{
		d:                  d,
		Scene:              s,
		Supervisor:         ui.NewSupervisor(),
		StatusMouseText:    "Cursor: (waiting)",
		StatusPaletteText:  "Swatch: <none>",
		StatusFilenameText: "Filename: <none>",
	}
	u.StatusBar = u.SetupStatusBar(d)
	u.Palette = u.SetupPalette(d)
	return u
}

// Loop to process events and update the UI.
func (u *EditorUI) Loop(ev *events.State) {
	u.Supervisor.Loop(ev)

	u.StatusMouseText = fmt.Sprintf("Mouse: (%d,%d)",
		ev.CursorX.Now,
		ev.CursorY.Now,
	)
	u.StatusPaletteText = fmt.Sprintf("Swatch: %s",
		u.Scene.Swatch,
	)

	// Statusbar filename label.
	filename := "untitled.map"
	if u.Scene.filename != "" {
		filename = u.Scene.filename
	}
	u.StatusFilenameText = fmt.Sprintf("Filename: %s",
		filename,
	)

	u.StatusBar.Compute(u.d.Engine)
	u.Palette.Compute(u.d.Engine)
}

// Present the UI to the screen.
func (u *EditorUI) Present(e render.Engine) {
	u.Palette.Present(e, u.Palette.Point())
	u.StatusBar.Present(e, u.StatusBar.Point())
}

// SetupPalette sets up the palette panel.
func (u *EditorUI) SetupPalette(d *Doodle) *ui.Window {
	window := ui.NewWindow("Palette")
	window.Configure(ui.Config{
		Width:  150,
		Height: u.d.height - u.StatusBar.Size().H,
	})
	window.MoveTo(render.NewPoint(
		u.d.width-window.BoxSize().W,
		0,
	))

	// Handler function for the radio buttons being clicked.
	onClick := func(p render.Point) {
		name := u.Scene.Palette.ActiveSwatch
		swatch, ok := u.Scene.Palette.Get(name)
		if !ok {
			log.Error("Palette onClick: couldn't get swatch named '%s' from palette", name)
			return
		}
		u.Scene.Swatch = swatch
	}

	// Draw the radio buttons for the palette.
	for _, swatch := range u.Scene.Palette.Swatches {
		label := ui.NewLabel(ui.Label{
			Text: swatch.Name,
			Font: balance.StatusFont,
		})
		label.Font.Color = swatch.Color.Darken(40)

		btn := ui.NewRadioButton("palette", &u.Scene.Palette.ActiveSwatch, swatch.Name, label)
		btn.Handle("MouseUp", onClick)
		u.Supervisor.Add(btn)

		window.Pack(btn, ui.Pack{
			Anchor: ui.N,
			Fill:   true,
		})
	}

	return window
}

// SetupStatusBar sets up the status bar widget along the bottom of the window.
func (u *EditorUI) SetupStatusBar(d *Doodle) *ui.Frame {
	frame := ui.NewFrame("Status Bar")
	frame.Configure(ui.Config{
		BorderStyle: ui.BorderRaised,
		Background:  render.Grey,
		BorderSize:  2,
		Width:       d.width,
	})

	style := ui.Config{
		Background:  render.Grey,
		BorderStyle: ui.BorderSunken,
		BorderColor: render.Grey,
		BorderSize:  1,
	}

	cursorLabel := ui.NewLabel(ui.Label{
		TextVariable: &u.StatusMouseText,
		Font:         balance.StatusFont,
	})
	cursorLabel.Configure(style)
	cursorLabel.Compute(d.Engine)
	frame.Pack(cursorLabel, ui.Pack{
		Anchor: ui.W,
	})

	paletteLabel := ui.NewLabel(ui.Label{
		TextVariable: &u.StatusPaletteText,
		Font:         balance.StatusFont,
	})
	paletteLabel.Configure(style)
	paletteLabel.Compute(d.Engine)
	frame.Pack(paletteLabel, ui.Pack{
		Anchor: ui.W,
	})

	filenameLabel := ui.NewLabel(ui.Label{
		TextVariable: &u.StatusFilenameText,
		Font:         balance.StatusFont,
	})
	filenameLabel.Configure(style)
	filenameLabel.Compute(d.Engine)
	frame.Pack(filenameLabel, ui.Pack{
		Anchor: ui.W,
	})

	// TODO: right-aligned labels clip out of bounds
	// extraLabel := ui.NewLabel(ui.Label{
	// 	Text: "blah",
	// 	Font: balance.StatusFont,
	// })
	// extraLabel.Configure(ui.Config{
	// 	Background:  render.Grey,
	// 	BorderStyle: ui.BorderSunken,
	// 	BorderColor: render.Grey,
	// 	BorderSize:  1,
	// })
	// extraLabel.Compute(d.Engine)
	// frame.Pack(extraLabel, ui.Pack{
	// 	Anchor: ui.E,
	// })

	frame.Resize(render.Rect{
		W: d.width,
		H: cursorLabel.BoxSize().H + frame.BoxThickness(1),
	})
	frame.Compute(d.Engine)
	frame.MoveTo(render.Point{
		X: 0,
		Y: d.height - frame.Size().H,
	})

	return frame
}
