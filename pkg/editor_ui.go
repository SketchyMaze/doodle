package doodle

import (
	"fmt"
	"path/filepath"
	"strconv"

	"git.kirsle.net/apps/doodle/lib/events"
	"git.kirsle.net/apps/doodle/lib/render"
	"git.kirsle.net/apps/doodle/lib/ui"
	"git.kirsle.net/apps/doodle/pkg/balance"
	"git.kirsle.net/apps/doodle/pkg/branding"
	"git.kirsle.net/apps/doodle/pkg/doodads"
	"git.kirsle.net/apps/doodle/pkg/drawtool"
	"git.kirsle.net/apps/doodle/pkg/enum"
	"git.kirsle.net/apps/doodle/pkg/level"
	"git.kirsle.net/apps/doodle/pkg/log"
	"git.kirsle.net/apps/doodle/pkg/uix"
)

// Width of the panel frame.
var paletteWidth int32 = 160

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
	ToolBar    *ui.Frame
	PlayButton *ui.Button

	// Palette window.
	Palette    *ui.Window
	PaletteTab *ui.Frame
	DoodadTab  *ui.Frame

	// Doodad Palette window variables.
	doodadSkip       int
	doodadRows       []*ui.Frame
	doodadPager      *ui.Frame
	doodadButtonSize int32
	doodadScroller   *ui.Frame

	// ToolBar window.
	activeTool string

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

	// Default tool in the toolbox.
	u.activeTool = drawtool.PencilTool.String()

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
	u.ToolBar = u.SetupToolbar(d)
	u.Workspace = u.SetupWorkspace(d) // important that this is last!

	u.PlayButton = ui.NewButton("Play", ui.NewLabel(ui.Label{
		Text: "Play (P)",
		Font: balance.PlayButtonFont,
	}))
	u.PlayButton.Handle(ui.Click, func(p render.Point) {
		u.Scene.Playtest()
	})
	u.Supervisor.Add(u.PlayButton)

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

// FinishSetup runs the Setup tasks that must be postponed til the end, such
// as rendering the Palette window so that it can accurately show the palette
// loaded from a level.
func (u *EditorUI) FinishSetup(d *Doodle) {
	u.Palette = u.SetupPalette(d)
	u.Resized(d)
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

		u.scrollDoodadFrame(0)
	}

	var innerHeight = int32(u.d.height) - u.MenuBar.Size().H - u.StatusBar.Size().H

	// Tool Bar.
	{
		u.ToolBar.Configure(ui.Config{
			Width:  toolbarWidth,
			Height: innerHeight,
		})
		u.ToolBar.MoveTo(render.NewPoint(
			0,
			u.MenuBar.BoxSize().H,
		))
		u.ToolBar.Compute(d.Engine)
	}

	// Position the workspace around with the other widgets.
	{

		frame := u.Workspace
		frame.MoveTo(render.NewPoint(
			u.ToolBar.Size().W,
			u.MenuBar.Size().H,
		))
		frame.Resize(render.NewRect(
			int32(d.width)-u.Palette.Size().W-u.ToolBar.Size().W,
			int32(d.height)-u.MenuBar.Size().H-u.StatusBar.Size().H,
		))
		frame.Compute(d.Engine)

		u.ExpandCanvas(d.Engine)
	}

	// Position the Play button over the workspace.
	{
		btn := u.PlayButton
		btn.Compute(d.Engine)

		var (
			wsP           = u.Workspace.Point()
			wsSize        = u.Workspace.Size()
			btnSize       = btn.Size()
			padding int32 = 8
		)
		btn.MoveTo(render.NewPoint(
			wsP.X+wsSize.W-btnSize.W-padding,
			wsP.Y+wsSize.H-btnSize.H-padding,
		))
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
		u.StatusMouseText = fmt.Sprintf("Rel:(%d,%d)  Abs:(%s)",
			ev.CursorX.Now,
			ev.CursorY.Now,
			*u.Scene.debWorldIndex,
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
	// Explanation: if I don't, the UI packing algorithm somehow causes widgets
	// to creep away every frame and fly right off the screen. For example the
	// ToolBar's buttons would start packed at the top of the bar but then just
	// move themselves every frame downward and away.
	u.MenuBar.Compute(u.d.Engine)
	u.StatusBar.Compute(u.d.Engine)
	u.Palette.Compute(u.d.Engine)
	u.ToolBar.Compute(u.d.Engine)

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
	u.ToolBar.Present(e, u.ToolBar.Point())
	u.Workspace.Present(e, u.Workspace.Point())
	u.PlayButton.Present(e, u.PlayButton.Point())

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
		doodad, err := doodads.LoadFile(filename)
		if err != nil {
			log.Error("drawing.OnDragStart: %s", err.Error())
		}
		u.startDragActor(doodad)
	}

	// A link event to connect two actors together.
	drawing.OnLinkActors = func(a, b *uix.Actor) {
		// The actors are a uix.Actor which houses a level.Actor which we
		// want to update to map each other's IDs.
		idA, idB := a.Actor.ID(), b.Actor.ID()
		a.Actor.AddLink(idB)
		b.Actor.AddLink(idA)

		// Reset the Link tool.
		drawing.Tool = drawtool.ActorTool
		d.Flash("Linked '%s' and '%s' together", a.Doodad.Title, b.Doodad.Title)
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

			err := drawing.InstallActors(u.Scene.Level.Actors)
			if err != nil {
				log.Error("Error installing actor onDrop to canvas: %s", err)
			}
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

	// Save and Save As common menu handler
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

	type menuButton struct {
		Text  string
		Click func(render.Point)
	}
	buttons := []menuButton{
		menuButton{
			Text: "New Level",
			Click: func(render.Point) {
				d.GotoNewMenu()
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
						saveFunc(answer)
					}
				})
			},
		},
		menuButton{
			Text: "Load",
			Click: func(render.Point) {
				d.GotoLoadMenu()
			},
		},
	}

	for _, btn := range buttons {
		if balance.FreeVersion {
			if btn.Text == "New Doodad" {
				continue
			}
		}

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

	var shareware string
	if balance.FreeVersion {
		shareware = " (shareware)"
	}
	extraLabel := ui.NewLabel(ui.Label{
		Text: fmt.Sprintf("%s v%s%s", branding.AppName, branding.Version, shareware),
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
