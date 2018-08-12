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
	MenuBar    *ui.Frame
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
	u.MenuBar = u.SetupMenuBar(d)
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

	u.MenuBar.Compute(u.d.Engine)
	u.StatusBar.Compute(u.d.Engine)
	u.Palette.Compute(u.d.Engine)
}

// Present the UI to the screen.
func (u *EditorUI) Present(e render.Engine) {
	// TODO: if I don't Compute() the palette window, then, whenever the dev console
	// is open the window will blank out its contents leaving only the outermost Frame.
	// The title bar and borders are gone. But other UI widgets don't do this.
	// FIXME: Scene interface should have a separate ComputeUI() from Loop()?
	u.Palette.Compute(u.d.Engine)

	u.Palette.Present(e, u.Palette.Point())
	u.MenuBar.Present(e, u.MenuBar.Point())
	u.StatusBar.Present(e, u.StatusBar.Point())
}

// SetupMenuBar sets up the menu bar.
func (u *EditorUI) SetupMenuBar(d *Doodle) *ui.Frame {
	frame := ui.NewFrame("MenuBar")
	frame.Configure(ui.Config{
		Width:      d.width,
		Background: render.Black,
	})

	type menuButton struct {
		Text  string
		Click func(render.Point)
	}
	buttons := []menuButton{
		menuButton{
			Text: "New Level",
			Click: func(render.Point) {
				d.NewMap()
			},
		},
		menuButton{
			Text: "New Doodad",
			Click: func(render.Point) {
				d.NewMap()
			},
		},
		menuButton{
			Text: "Save",
			Click: func(render.Point) {
				if u.Scene.filename != "" {
					u.Scene.SaveLevel(u.Scene.filename)
					d.Flash("Saved: %s", u.Scene.filename)
				} else {
					d.Prompt("Save filename>", func(answer string) {
						if answer != "" {
							u.Scene.SaveLevel("./maps/" + answer) // TODO: maps path
							d.Flash("Saved: %s", answer)
						}
					})
				}
			},
		},
		menuButton{
			Text: "Save as...",
			Click: func(render.Point) {
				d.Prompt("Save as filename>", func(answer string) {
					if answer != "" {
						u.Scene.SaveLevel("./maps/" + answer) // TODO: maps path
						d.Flash("Saved: %s", answer)
					}
				})
			},
		},
		menuButton{
			Text: "Load",
			Click: func(render.Point) {
				d.Prompt("Open filename>", func(answer string) {
					if answer != "" {
						u.d.EditLevel("./maps/" + answer) // TODO: maps path
					}
				})
			},
		},
	}

	for _, btn := range buttons {
		w := ui.NewButton(btn.Text, ui.NewLabel(ui.Label{
			Text: btn.Text,
			Font: balance.MenuFont,
		}))
		w.Configure(ui.Config{
			BorderSize:  1,
			OutlineSize: 0,
		})
		w.Handle("MouseUp", btn.Click)
		u.Supervisor.Add(w)
		frame.Pack(w, ui.Pack{
			Anchor: ui.W,
			PadX:   1,
		})
	}

	frame.Compute(d.Engine)
	return frame
}

// SetupPalette sets up the palette panel.
func (u *EditorUI) SetupPalette(d *Doodle) *ui.Window {
	log.Error("SetupPalette Window")
	window := ui.NewWindow("Palette")
	window.ConfigureTitle(balance.TitleConfig)
	window.TitleBar().Font = balance.TitleFont
	window.Configure(ui.Config{
		Width:       150,
		Height:      u.d.height - u.StatusBar.Size().H,
		Background:  balance.WindowBackground,
		BorderColor: balance.WindowBorder,
	})
	window.MoveTo(render.NewPoint(
		u.d.width-window.BoxSize().W,
		u.MenuBar.BoxSize().H,
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
			PadY:   4,
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
		PadX:   1,
	})

	paletteLabel := ui.NewLabel(ui.Label{
		TextVariable: &u.StatusPaletteText,
		Font:         balance.StatusFont,
	})
	paletteLabel.Configure(style)
	paletteLabel.Compute(d.Engine)
	frame.Pack(paletteLabel, ui.Pack{
		Anchor: ui.W,
		PadX:   1,
	})

	filenameLabel := ui.NewLabel(ui.Label{
		TextVariable: &u.StatusFilenameText,
		Font:         balance.StatusFont,
	})
	filenameLabel.Configure(style)
	filenameLabel.Compute(d.Engine)
	frame.Pack(filenameLabel, ui.Pack{
		Anchor: ui.W,
		PadX:   1,
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
