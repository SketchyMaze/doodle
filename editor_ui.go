package doodle

import (
	"fmt"
	"strconv"

	"git.kirsle.net/apps/doodle/balance"
	"git.kirsle.net/apps/doodle/enum"
	"git.kirsle.net/apps/doodle/events"
	"git.kirsle.net/apps/doodle/level"
	"git.kirsle.net/apps/doodle/render"
	"git.kirsle.net/apps/doodle/ui"
	"git.kirsle.net/apps/doodle/uix"
)

// EditorUI manages the user interface for the Editor Scene.
type EditorUI struct {
	d     *Doodle
	Scene *EditorScene

	// Variables
	StatusMouseText    string
	StatusPaletteText  string
	StatusFilenameText string
	selectedSwatch     string // name of selected swatch in palette

	// Widgets
	Supervisor *ui.Supervisor
	Canvas     *uix.Canvas
	Workspace  *ui.Frame
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

	u.Canvas = u.SetupCanvas(d)
	u.MenuBar = u.SetupMenuBar(d)
	u.StatusBar = u.SetupStatusBar(d)
	u.Palette = u.SetupPalette(d)
	u.Workspace = u.SetupWorkspace(d) // important that this is last!

	// Position the Canvas inside the frame.
	u.Workspace.Pack(u.Canvas, ui.Pack{
		Anchor: ui.N,
	})
	u.Workspace.Compute(d.Engine)
	u.ExpandCanvas(d.Engine)

	// Select the first swatch of the palette.
	if u.Canvas.Palette != nil && u.Canvas.Palette.ActiveSwatch != nil {
		u.selectedSwatch = u.Canvas.Palette.ActiveSwatch.Name
	}
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
		u.Canvas.Palette.ActiveSwatch,
	)

	// Statusbar filename label.
	filename := "untitled.map"
	fileType := "Level"
	if u.Scene.filename != "" {
		filename = u.Scene.filename
	}
	if u.Scene.DrawingType == enum.DoodadDrawing {
		fileType = "Doodad"
	}
	u.StatusFilenameText = fmt.Sprintf("Filename: %s (%s)",
		filename,
		fileType,
	)

	u.MenuBar.Compute(u.d.Engine)
	u.StatusBar.Compute(u.d.Engine)
	u.Palette.Compute(u.d.Engine)
	u.Canvas.Loop(ev)
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
	u.Workspace.Present(e, u.Workspace.Point())
}

// SetupWorkspace configures the main Workspace frame that takes up the full
// window apart from toolbars. The Workspace has a single child element, the
// Canvas, so it can easily full-screen it or center it for Doodad editing.
func (u *EditorUI) SetupWorkspace(d *Doodle) *ui.Frame {
	frame := ui.NewFrame("Workspace")

	// Position and size the frame around the other main widgets.
	frame.MoveTo(render.NewPoint(
		0,
		u.MenuBar.Size().H,
	))
	frame.Resize(render.NewRect(
		d.width-u.Palette.Size().W,
		d.height-u.MenuBar.Size().H-u.StatusBar.Size().H,
	))
	frame.Compute(d.Engine)

	return frame
}

// SetupCanvas configures the main drawing canvas in the editor.
func (u *EditorUI) SetupCanvas(d *Doodle) *uix.Canvas {
	drawing := uix.NewCanvas(balance.ChunkSize, true)
	drawing.Palette = level.DefaultPalette()
	if len(drawing.Palette.Swatches) > 0 {
		drawing.SetSwatch(drawing.Palette.Swatches[0])
	}
	return drawing
}

// ExpandCanvas manually expands the Canvas to fill the frame, to work around
// UI packing bugs. Ideally I would use `Expand: true` when packing the Canvas
// in its frame, but that would artificially expand the Canvas also when it
// _wanted_ to be smaller, as in Doodad Editing Mode.
func (u *EditorUI) ExpandCanvas(e render.Engine) {
	u.Canvas.Resize(u.Workspace.Size())
	u.Workspace.Compute(e)
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
				d.Prompt("Doodad size [100]>", func(answer string) {
					size := balance.DoodadSize
					if answer != "" {
						i, err := strconv.Atoi(answer)
						if err != nil {
							d.Flash("Error: Doodad size must be a number.")
							return
						}
						size = i
					}
					d.NewDoodad(size)
				})
			},
		},
		menuButton{
			Text: "Save",
			Click: func(render.Point) {
				var saveFunc func(filename string)

				switch u.Scene.DrawingType {
				case enum.LevelDrawing:
					saveFunc = func(filename string) {
						if err := u.Scene.SaveLevel(filename); err != nil {
							d.Flash("Error: %s", err)
						} else {
							d.Flash("Saved level: %s", filename)
						}
					}
				case enum.DoodadDrawing:
					saveFunc = func(filename string) {
						if err := u.Scene.SaveDoodad(filename); err != nil {
							d.Flash("Error: %s", err)
						} else {
							d.Flash("Saved doodad: %s", filename)
						}
					}
				default:
					d.Flash("Error: Scene.DrawingType is not a valid type")
				}

				if u.Scene.filename != "" {
					saveFunc(u.Scene.filename)
				} else {
					d.Prompt("Save filename>", func(answer string) {
						if answer != "" {
							saveFunc(answer)
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
						u.d.EditDrawing("./maps/" + answer) // TODO: maps path
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
		w.Handle(ui.MouseUp, btn.Click)
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

			window.Pack(btn, ui.Pack{
				Anchor: ui.N,
				Fill:   true,
				PadY:   4,
			})
		}
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
		Anchor: ui.E,
		PadX:   1,
	})

	// TODO: right-aligned labels clip out of bounds
	extraLabel := ui.NewLabel(ui.Label{
		Text: "blah",
		Font: balance.StatusFont,
	})
	extraLabel.Configure(ui.Config{
		Background:  render.Grey,
		BorderStyle: ui.BorderSunken,
		BorderColor: render.Grey,
		BorderSize:  1,
	})
	extraLabel.Compute(d.Engine)
	frame.Pack(extraLabel, ui.Pack{
		Anchor: ui.E,
	})

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
