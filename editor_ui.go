package doodle

import (
	"fmt"
	"path/filepath"
	"strconv"

	"git.kirsle.net/apps/doodle/balance"
	"git.kirsle.net/apps/doodle/doodads"
	"git.kirsle.net/apps/doodle/enum"
	"git.kirsle.net/apps/doodle/events"
	"git.kirsle.net/apps/doodle/level"
	"git.kirsle.net/apps/doodle/pkg/userdir"
	"git.kirsle.net/apps/doodle/render"
	"git.kirsle.net/apps/doodle/ui"
	"git.kirsle.net/apps/doodle/uix"
)

// Width of the panel frame.
var paletteWidth int32 = 150

// EditorUI manages the user interface for the Editor Scene.
type EditorUI struct {
	d     *Doodle
	Scene *EditorScene

	// Variables
	StatusBoxes        []*string
	StatusMouseText    string
	StatusPaletteText  string
	StatusFilenameText string
	StatusScrollText   string
	selectedSwatch     string       // name of selected swatch in palette
	cursor             render.Point // remember the cursor position in Loop

	// Widgets
	Supervisor *ui.Supervisor
	Canvas     *uix.Canvas
	Workspace  *ui.Frame
	MenuBar    *ui.Frame
	StatusBar  *ui.Frame

	// Palette window.
	Palette    *ui.Window
	PaletteTab *ui.Frame
	DoodadTab  *ui.Frame

	// Draggable Doodad canvas.
	DraggableActor *DraggableActor

	// Palette variables.
	paletteTab string // selected tab, Palette or Doodads
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
		StatusScrollText:   "Hello world",
	}

	// Bind the StatusBoxes arrays to the text variables.
	u.StatusBoxes = []*string{
		&u.StatusMouseText,
		&u.StatusPaletteText,
		&u.StatusFilenameText,
		&u.StatusScrollText,
	}

	u.Canvas = u.SetupCanvas(d)
	u.MenuBar = u.SetupMenuBar(d)
	u.StatusBar = u.SetupStatusBar(d)
	u.Palette = u.SetupPalette(d)
	u.Workspace = u.SetupWorkspace(d) // important that this is last!

	u.Resized(d)

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

// Resized handles the window being resized so we can recompute the widgets.
func (u *EditorUI) Resized(d *Doodle) {
	// Menu Bar frame.
	{
		u.MenuBar.Configure(ui.Config{
			Width:      int32(d.width),
			Background: render.Black,
		})
		u.MenuBar.Compute(d.Engine)
	}

	// Status Bar.
	{
		u.StatusBar.Configure(ui.Config{
			Width: int32(d.width),
		})
		u.StatusBar.MoveTo(render.Point{
			X: 0,
			Y: int32(d.height) - u.StatusBar.Size().H,
		})
		u.StatusBar.Compute(d.Engine)
	}

	// Palette panel.
	{
		u.Palette.Configure(ui.Config{
			Width:  paletteWidth,
			Height: int32(u.d.height) - u.StatusBar.Size().H,
		})
		u.Palette.MoveTo(render.NewPoint(
			int32(u.d.width)-u.Palette.BoxSize().W,
			u.MenuBar.BoxSize().H,
		))
		u.Palette.Compute(d.Engine)
	}

	// Position the workspace around with the other widgets.
	{
		frame := u.Workspace
		frame.MoveTo(render.NewPoint(
			0,
			u.MenuBar.Size().H,
		))
		frame.Resize(render.NewRect(
			int32(d.width)-u.Palette.Size().W,
			int32(d.height)-u.MenuBar.Size().H-u.StatusBar.Size().H,
		))
		frame.Compute(d.Engine)

		u.ExpandCanvas(d.Engine)
	}
}

// Loop to process events and update the UI.
func (u *EditorUI) Loop(ev *events.State) error {
	u.cursor = render.NewPoint(ev.CursorX.Now, ev.CursorY.Now)

	// Loop the UI and see whether we're told to stop event propagation.
	var stopPropagation bool
	if err := u.Supervisor.Loop(ev); err != nil {
		if err == ui.ErrStopPropagation {
			stopPropagation = true
		} else {
			return err
		}
	}

	// Update status bar labels.
	{
		debugWorldIndex = u.Canvas.WorldIndexAt(u.cursor)
		u.StatusMouseText = fmt.Sprintf("Rel:(%d,%d)  Abs:(%s)",
			ev.CursorX.Now,
			ev.CursorY.Now,
			debugWorldIndex,
		)
		u.StatusPaletteText = fmt.Sprintf("%s Tool",
			u.Canvas.Tool,
		)
		u.StatusScrollText = fmt.Sprintf("Scroll: %s   Viewport: %s",
			u.Canvas.Scroll,
			u.Canvas.Viewport(),
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
			filepath.Base(filename),
			fileType,
		)
	}

	// Recompute widgets.
	u.MenuBar.Compute(u.d.Engine)
	u.StatusBar.Compute(u.d.Engine)
	u.Palette.Compute(u.d.Engine)

	// Only forward events to the Canvas if the UI hasn't stopped them.
	if !stopPropagation {
		u.Canvas.Loop(ev)
	}
	return nil
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

	// Are we dragging a Doodad canvas?
	if u.Supervisor.IsDragging() {
		if actor := u.DraggableActor; actor != nil {
			var size = actor.canvas.Size()
			actor.canvas.Present(u.d.Engine, render.NewPoint(
				u.cursor.X-(size.W/2),
				u.cursor.Y-(size.H/2),
			))
		}
	}
}

// SetupWorkspace configures the main Workspace frame that takes up the full
// window apart from toolbars. The Workspace has a single child element, the
// Canvas, so it can easily full-screen it or center it for Doodad editing.
func (u *EditorUI) SetupWorkspace(d *Doodle) *ui.Frame {
	frame := ui.NewFrame("Workspace")
	return frame
}

// SetupCanvas configures the main drawing canvas in the editor.
func (u *EditorUI) SetupCanvas(d *Doodle) *uix.Canvas {
	drawing := uix.NewCanvas(balance.ChunkSize, true)
	drawing.Name = "edit-canvas"
	drawing.Palette = level.DefaultPalette()
	drawing.SetBackground(render.White)
	if len(drawing.Palette.Swatches) > 0 {
		drawing.SetSwatch(drawing.Palette.Swatches[0])
	}

	// Handle the Canvas deleting our actors in edit mode.
	drawing.OnDeleteActors = func(actors []*level.Actor) {
		if u.Scene.Level != nil {
			for _, actor := range actors {
				u.Scene.Level.Actors.Remove(actor)
			}
			drawing.InstallActors(u.Scene.Level.Actors)
		}
	}

	// A drag event initiated inside the Canvas. This happens in the ActorTool
	// mode when you click an existing Doodad and it "pops" out of the canvas
	// and onto the cursor to be repositioned.
	drawing.OnDragStart = func(filename string) {
		doodad, err := doodads.LoadJSON(userdir.DoodadPath(filename))
		if err != nil {
			log.Error("drawing.OnDragStart: %s", err.Error())
		}
		u.startDragActor(doodad)
	}

	// Set up the drop handler for draggable doodads.
	// NOTE: The drag event begins at editor_ui_doodad.go when configuring the
	// Doodad Palette buttons.
	drawing.Handle(ui.Drop, func(e render.Point) {
		log.Info("Drawing canvas has received a drop!")
		var P = ui.AbsolutePosition(drawing)

		// Was it an actor from the Doodad Palette?
		if actor := u.DraggableActor; actor != nil {
			log.Info("Actor is a %s", actor.doodad.Filename)
			if u.Scene.Level == nil {
				u.d.Flash("Can't drop doodads onto doodad drawings!")
				return
			}

			size := actor.canvas.Size()
			u.Scene.Level.Actors.Add(&level.Actor{
				// Uncenter the drawing from the cursor.
				Point: render.Point{
					X: (u.cursor.X - drawing.Scroll.X - (size.W / 2)) - P.X,
					Y: (u.cursor.Y - drawing.Scroll.Y - (size.H / 2)) - P.Y,
				},
				Filename: actor.doodad.Filename,
			})

			drawing.InstallActors(u.Scene.Level.Actors)
		}
	})
	u.Supervisor.Add(drawing)
	return drawing
}

// ExpandCanvas manually expands the Canvas to fill the frame, to work around
// UI packing bugs. Ideally I would use `Expand: true` when packing the Canvas
// in its frame, but that would artificially expand the Canvas also when it
// _wanted_ to be smaller, as in Doodad Editing Mode.
func (u *EditorUI) ExpandCanvas(e render.Engine) {
	if u.Scene.DrawingType == enum.LevelDrawing {
		u.Canvas.Resize(u.Workspace.Size())
	} else {
		// Size is managed externally.
	}
	u.Workspace.Compute(e)
}

// SetupMenuBar sets up the menu bar.
func (u *EditorUI) SetupMenuBar(d *Doodle) *ui.Frame {
	frame := ui.NewFrame("MenuBar")

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

// SetupStatusBar sets up the status bar widget along the bottom of the window.
func (u *EditorUI) SetupStatusBar(d *Doodle) *ui.Frame {
	frame := ui.NewFrame("Status Bar")
	frame.Configure(ui.Config{
		BorderStyle: ui.BorderRaised,
		Background:  render.Grey,
		BorderSize:  2,
	})

	style := ui.Config{
		Background:  render.Grey,
		BorderStyle: ui.BorderSunken,
		BorderColor: render.Grey,
		BorderSize:  1,
	}

	var labelHeight int32
	for _, variable := range u.StatusBoxes {
		label := ui.NewLabel(ui.Label{
			TextVariable: variable,
			Font:         balance.StatusFont,
		})
		label.Configure(style)
		label.Compute(d.Engine)
		frame.Pack(label, ui.Pack{
			Anchor: ui.W,
			PadX:   1,
		})

		if labelHeight == 0 {
			labelHeight = label.BoxSize().H
		}
	}

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

	// Set the initial good frame size to have the height secured,
	// so when resizing the application window we can just adjust for width.
	frame.Resize(render.Rect{
		W: int32(d.width),
		H: labelHeight + frame.BoxThickness(1),
	})
	frame.Compute(d.Engine)

	return frame
}
