package doodle

import (
	"fmt"
	"path/filepath"
	"strconv"

	"git.kirsle.net/apps/doodle/pkg/balance"
	"git.kirsle.net/apps/doodle/pkg/branding"
	"git.kirsle.net/apps/doodle/pkg/drawtool"
	"git.kirsle.net/apps/doodle/pkg/enum"
	"git.kirsle.net/apps/doodle/pkg/level"
	"git.kirsle.net/apps/doodle/pkg/license"
	"git.kirsle.net/apps/doodle/pkg/log"
	"git.kirsle.net/apps/doodle/pkg/native"
	"git.kirsle.net/apps/doodle/pkg/uix"
	"git.kirsle.net/apps/doodle/pkg/windows"
	"git.kirsle.net/go/render"
	"git.kirsle.net/go/render/event"
	"git.kirsle.net/go/ui"
)

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
	screen     *ui.Frame // full-window parent frame for layout
	Supervisor *ui.Supervisor
	Canvas     *uix.Canvas
	Workspace  *ui.Frame
	MenuBar    *ui.MenuBar
	StatusBar  *ui.Frame
	ToolBar    *ui.Frame
	PlayButton *ui.Button

	// Popup windows.
	levelSettingsWindow *ui.Window
	aboutWindow         *ui.Window
	doodadWindow        *ui.Window
	paletteEditor       *ui.Window
	layersWindow        *ui.Window
	publishWindow       *ui.Window
	filesystemWindow    *ui.Window
	licenseWindow       *ui.Window

	// Palette window.
	Palette    *ui.Window
	PaletteTab *ui.Frame
	DoodadTab  *ui.Frame

	// ToolBar window.
	activeTool string

	// Draggable Doodad canvas.
	DraggableActor *DraggableActor
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

	// The screen is a full-window-sized frame for laying out the UI.
	u.screen = ui.NewFrame("screen")
	u.screen.Resize(render.NewRect(d.width, d.height))
	u.screen.Compute(d.Engine)

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

	// Preload pop-up windows before they're needed.
	u.SetupPopups(d)

	u.screen.Pack(u.MenuBar, ui.Pack{
		Side:  ui.N,
		FillX: true,
	})

	u.PlayButton = ui.NewButton("Play", ui.NewLabel(ui.Label{
		Text: "Play (P)",
		Font: balance.PlayButtonFont,
	}))
	u.PlayButton.Handle(ui.Click, func(ed ui.EventData) error {
		u.Scene.Playtest()
		return nil
	})
	u.Supervisor.Add(u.PlayButton)

	// Position the Canvas inside the frame.
	u.Workspace.Pack(u.Canvas, ui.Pack{
		Side: ui.N,
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
	// Resize the screen frame to fill the window.
	u.screen.Resize(render.NewRect(d.width, d.height))
	u.screen.Compute(d.Engine)
	menuHeight := 20 // TODO: ideally the MenuBar should know its own height and we can ask

	// Status Bar.
	{
		u.StatusBar.Configure(ui.Config{
			Width: d.width,
		})
		u.StatusBar.MoveTo(render.Point{
			X: 0,
			Y: d.height - u.StatusBar.Size().H,
		})
		u.StatusBar.Compute(d.Engine)
	}

	// Palette panel.
	{
		if balance.HorizontalToolbars {
			u.Palette.Configure(ui.Config{
				Width:  u.d.width,
				Height: paletteWidth,
			})
			u.Palette.MoveTo(render.NewPoint(
				0,
				u.d.height-u.Palette.BoxSize().H-u.StatusBar.Size().H,
			))
		} else {
			u.Palette.Configure(ui.Config{
				Width:  paletteWidth,
				Height: u.d.height - u.StatusBar.Size().H,
			})
			u.Palette.MoveTo(render.NewPoint(
				u.d.width-u.Palette.BoxSize().W,
				menuHeight,
			))
		}
		u.Palette.Compute(d.Engine)
	}

	var (
		innerHeight = u.d.height - menuHeight - u.StatusBar.Size().H
		innerWidth  = u.d.width
	)

	// Tool Bar.
	{
		tbSize := ui.Config{
			Width:  toolbarWidth,
			Height: innerHeight,
		}
		if balance.HorizontalToolbars {
			tbSize.Width = innerWidth
			tbSize.Height = toolbarWidth
		}
		u.ToolBar.Configure(tbSize)
		u.ToolBar.MoveTo(render.NewPoint(
			0,
			menuHeight,
		))
		u.ToolBar.Compute(d.Engine)
	}

	// Position the workspace around with the other widgets.
	{

		frame := u.Workspace
		if balance.HorizontalToolbars {
			frame.MoveTo(render.NewPoint(
				0,
				menuHeight+u.ToolBar.Size().H,
			))
			frame.Resize(render.NewRect(
				d.width,
				d.height-menuHeight-u.StatusBar.Size().H-u.ToolBar.Size().H-u.Palette.Size().H,
			))
		} else {
			frame.MoveTo(render.NewPoint(
				u.ToolBar.Size().W,
				menuHeight,
			))
			frame.Resize(render.NewRect(
				d.width-u.Palette.Size().W-u.ToolBar.Size().W,
				d.height-menuHeight-u.StatusBar.Size().H,
			))
		}
		frame.Compute(d.Engine)

		u.ExpandCanvas(d.Engine)
	}

	// Position the Play button over the workspace.
	{
		btn := u.PlayButton
		btn.Compute(d.Engine)

		var (
			wsP     = u.Workspace.Point()
			wsSize  = u.Workspace.Size()
			btnSize = btn.Size()
			padding = 8
		)
		btn.MoveTo(render.NewPoint(
			wsP.X+wsSize.W-btnSize.W-padding,
			wsP.Y+wsSize.H-btnSize.H-padding,
		))
	}
}

// Loop to process events and update the UI.
func (u *EditorUI) Loop(ev *event.State) error {
	u.cursor = render.NewPoint(ev.CursorX, ev.CursorY)

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
			ev.CursorX,
			ev.CursorY,
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
		filename := "untitled.level"
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
	// Also ignore events if a managed ui.Window is overlapping the canvas.
	// Also ignore if an active modal (popup menu) is on screen.
	if !(stopPropagation || u.Supervisor.IsPointInWindow(u.cursor) || u.Supervisor.GetModal() != nil) {
		u.Canvas.Loop(ev)
	}
	return nil
}

// Present the UI to the screen.
func (u *EditorUI) Present(e render.Engine) {
	// Draw the workspace canvas first. Rationale: we want it to have the lowest
	// Z-index so Tooltips can pop on top of the workspace.
	u.Workspace.Present(e, u.Workspace.Point())

	// TODO: if I don't Compute() the palette window, then, whenever the dev console
	// is open the window will blank out its contents leaving only the outermost Frame.
	// The title bar and borders are gone. But other UI widgets don't do this.
	// FIXME: Scene interface should have a separate ComputeUI() from Loop()?
	u.Palette.Compute(u.d.Engine)

	u.Palette.Present(e, u.Palette.Point())
	u.MenuBar.Present(e, u.MenuBar.Point())
	u.StatusBar.Present(e, u.StatusBar.Point())
	u.ToolBar.Present(e, u.ToolBar.Point())
	u.PlayButton.Present(e, u.PlayButton.Point())

	u.screen.Present(e, render.Origin)

	// Draw any windows being managed by Supervisor.
	u.Supervisor.Present(e)

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
	drawing.OnDragStart = func(actor *level.Actor) {
		log.Warn("drawing.OnDragStart: grab actor %s", actor)
		u.startDragActor(nil, actor)
	}

	// A link event to connect two actors together.
	drawing.OnLinkActors = func(a, b *uix.Actor) {
		// The actors are a uix.Actor which houses a level.Actor which we
		// want to update to map each other's IDs.
		idA, idB := a.Actor.ID(), b.Actor.ID()
		a.Actor.AddLink(idB)
		b.Actor.AddLink(idA)

		// Reset the Link tool.
		d.Flash("Linked '%s' and '%s' together", a.Doodad().Title, b.Doodad().Title)
	}

	// Set up the drop handler for draggable doodads.
	// NOTE: The drag event begins at editor_ui_doodad.go when configuring the
	// Doodad Palette buttons.
	drawing.Handle(ui.Drop, func(ed ui.EventData) error {
		log.Info("Drawing canvas has received a drop!")
		var P = ui.AbsolutePosition(drawing)

		// Was it an actor from the Doodad Palette?
		if actor := u.DraggableActor; actor != nil {
			log.Info("Actor is a %s", actor.doodad.Filename)

			// The actor has been dropped so null it out.
			defer func() {
				u.DraggableActor = nil
			}()

			if u.Scene.Level == nil {
				u.d.Flash("Can't drop doodads onto doodad drawings!")
				return nil
			}

			// If they dropped it onto a UI window, ignore it.
			if u.Supervisor.IsPointInWindow(ed.Point) {
				return nil
			}

			var (
				// Uncenter the drawing from the cursor.
				size     = actor.canvas.Size()
				position = render.Point{
					X: (u.cursor.X - drawing.Scroll.X - (size.W / 2)) - P.X,
					Y: (u.cursor.Y - drawing.Scroll.Y - (size.H / 2)) - P.Y,
				}
			)

			// Was it an already existing actor to re-add to the map?
			if actor.actor != nil {
				actor.actor.Point = position
				u.Scene.Level.Actors.Add(actor.actor)
			} else {
				u.Scene.Level.Actors.Add(&level.Actor{
					Point:    position,
					Filename: actor.doodad.Filename,
				})
			}

			err := drawing.InstallActors(u.Scene.Level.Actors)
			if err != nil {
				log.Error("Error installing actor onDrop to canvas: %s", err)
			}
		}

		return nil
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
func (u *EditorUI) SetupMenuBar(d *Doodle) *ui.MenuBar {
	menu := ui.NewMenuBar("Main Menu")

	// Save and Save As common menu handler
	var (
		drawingType string
		saveFunc    func(filename string)
	)

	switch u.Scene.DrawingType {
	case enum.LevelDrawing:
		drawingType = "level"
		saveFunc = func(filename string) {
			if err := u.Scene.SaveLevel(filename); err != nil {
				d.Flash("Error: %s", err)
			} else {
				d.Flash("Saved level: %s", filename)
			}
		}
	case enum.DoodadDrawing:
		drawingType = "doodad"
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

	////////
	// File menu
	fileMenu := menu.AddMenu("File")
	fileMenu.AddItemAccel("New level", "Ctrl-N*", func() {
		u.Scene.ConfirmUnload(func() {
			d.GotoNewMenu()
		})
	})
	fileMenu.AddItem("New doodad", func() {
		u.Scene.ConfirmUnload(func() {
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
		})
	})
	fileMenu.AddItemAccel("Save", "Ctrl-S*", func() {
		if u.Scene.filename != "" {
			saveFunc(u.Scene.filename)
		} else {
			d.Prompt("Save filename>", func(answer string) {
				if answer != "" {
					saveFunc(answer)
				}
			})
		}
	})
	fileMenu.AddItem("Save as...", func() {
		d.Prompt("Save as filename>", func(answer string) {
			if answer != "" {
				saveFunc(answer)
			}
		})
	})

	if balance.Feature.EmbeddableDoodads && drawingType == "level" {
		fileMenu.AddItem("Publish level", func() {
			u.OpenPublishWindow()
		})
	}

	fileMenu.AddItemAccel("Open...", "Ctrl-O*", func() {
		u.Scene.ConfirmUnload(func() {
			d.GotoLoadMenu()
		})
	})
	fileMenu.AddSeparator()
	fileMenu.AddItem("Close "+drawingType, func() {
		u.Scene.ConfirmUnload(func() {
			d.Goto(&MainScene{})
		})
	})
	fileMenu.AddItemAccel("Quit", "Escape", func() {
		d.ConfirmExit()
	})

	////////
	// Edit menu
	editMenu := menu.AddMenu("Edit")
	editMenu.AddItemAccel("Undo", "Ctrl-Z", func() {
		u.Canvas.UndoStroke()
	})
	editMenu.AddItemAccel("Redo", "Ctrl-Y", func() {
		u.Canvas.RedoStroke()
	})

	////////
	// Level menu
	if u.Scene.DrawingType == enum.LevelDrawing {
		levelMenu := menu.AddMenu("Level")
		levelMenu.AddItem("Page settings", func() {
			log.Info("Opening the window")

			// Open the New Level window in edit-settings mode.
			u.levelSettingsWindow.Hide()
			u.levelSettingsWindow = nil
			u.SetupPopups(u.d)
			u.levelSettingsWindow.Show()
		})
		levelMenu.AddItem("Attached files", func() {
			log.Info("Opening the FileSystem window")
			u.OpenFileSystemWindow()
		})
		levelMenu.AddItemAccel("Playtest", "P", func() {
			u.Scene.Playtest()
		})
	}

	////////
	// View menu
	if balance.Feature.Zoom {
		viewMenu := menu.AddMenu("View")
		viewMenu.AddItemAccel("Zoom in", "+", func() {
			u.Canvas.Zoom++
		})
		viewMenu.AddItemAccel("Zoom out", "-", func() {
			u.Canvas.Zoom--
		})
		viewMenu.AddItemAccel("Reset zoom", "1", func() {
			u.Canvas.Zoom = 0
		})
		viewMenu.AddItemAccel("Scroll drawing to origin", "0", func() {
			u.Canvas.ScrollTo(render.Origin)
		})
	}

	////////
	// Tools menu
	toolMenu := menu.AddMenu("Tools")
	toolMenu.AddItemAccel("Debug overlay", "F3", func() {
		DebugOverlay = !DebugOverlay
		if DebugOverlay {
			d.Flash("Debug overlay enabled. Press F3 to turn it off.")
		}
	})
	toolMenu.AddItemAccel("Command shell", "Enter", func() {
		d.shell.Open = true
	})
	toolMenu.AddSeparator()
	toolMenu.AddItem("Edit Palette", func() {
		u.OpenPaletteWindow()
	})
	if u.Scene.DrawingType == enum.LevelDrawing {
		toolMenu.AddItemAccel("Doodads", "d", func() {
			log.Info("Open the DoodadDropper")
			u.doodadWindow.Show()
		})
	} else if u.Scene.DrawingType == enum.DoodadDrawing {
		toolMenu.AddItem("Layers", func() {
			u.OpenLayersWindow()
		})
	}

	// Draw Tools
	toolMenu.AddItemAccel("Pencil Tool", "F", func() {
		u.Canvas.Tool = drawtool.PencilTool
		u.activeTool = u.Canvas.Tool.String()
		d.Flash("Pencil Tool selected.")
	})
	toolMenu.AddItemAccel("Line Tool", "L", func() {
		u.Canvas.Tool = drawtool.LineTool
		u.activeTool = u.Canvas.Tool.String()
		d.Flash("Line Tool selected.")
	})
	toolMenu.AddItemAccel("Rectangle Tool", "R", func() {
		u.Canvas.Tool = drawtool.RectTool
		u.activeTool = u.Canvas.Tool.String()
		d.Flash("Rectangle Tool selected.")
	})
	toolMenu.AddItemAccel("Ellipse Tool", "C", func() {
		u.Canvas.Tool = drawtool.EllipseTool
		u.activeTool = u.Canvas.Tool.String()
		d.Flash("Ellipse Tool selected.")
	})
	toolMenu.AddItemAccel("Eraser Tool", "x", func() {
		u.Canvas.Tool = drawtool.EraserTool
		u.activeTool = u.Canvas.Tool.String()
		d.Flash("Eraser Tool selected.")
	})

	if u.Scene.DrawingType == enum.LevelDrawing {
		toolMenu.AddItemAccel("Doodads", "d", func() {
			log.Info("Open the DoodadDropper")
			u.doodadWindow.Show()
		})
		toolMenu.AddItem("Link Tool", func() {
			u.Canvas.Tool = drawtool.LinkTool
			u.activeTool = u.Canvas.Tool.String()
			d.Flash("Link Tool selected. Click a doodad in your level to link it to another.")
		})
	}

	////////
	// Help menu
	helpMenu := menu.AddMenu("Help")
	helpMenu.AddItemAccel("User Manual", "F1", func() {
		native.OpenLocalURL(balance.GuidebookPath)
	})
	helpMenu.AddItem("Register", func() {
		if u.licenseWindow == nil {
			cfg := windows.License{
				Supervisor: u.Supervisor,
				Engine:     d.Engine,
				OnCancel: func() {
					u.licenseWindow.Hide()
				},
			}
			cfg.OnLicensed = func() {
				// License status has changed, reload the window!
				if u.licenseWindow != nil {
					u.licenseWindow.Hide()
				}
				u.licenseWindow = windows.MakeLicenseWindow(d.width, d.height, cfg)
			}

			cfg.OnLicensed()
		}
		u.licenseWindow.Show()
	})
	helpMenu.AddItem("About", func() {
		if u.aboutWindow == nil {
			u.aboutWindow = windows.NewAboutWindow(windows.About{
				Supervisor: u.Supervisor,
				Engine:     d.Engine,
			})
			u.aboutWindow.Compute(d.Engine)
			u.aboutWindow.Supervise(u.Supervisor)

			// Center the window.
			u.aboutWindow.MoveTo(render.Point{
				X: (d.width / 2) - (u.aboutWindow.Size().W / 2),
				Y: 60,
			})
		}
		u.aboutWindow.Show()
	})

	menu.Supervise(u.Supervisor)
	menu.Compute(d.Engine)

	return menu
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

	var labelHeight int
	for _, variable := range u.StatusBoxes {
		label := ui.NewLabel(ui.Label{
			TextVariable: variable,
			Font:         balance.StatusFont,
		})
		label.Configure(style)
		label.Compute(d.Engine)
		frame.Pack(label, ui.Pack{
			Side: ui.W,
			PadX: 1,
		})

		if labelHeight == 0 {
			labelHeight = label.BoxSize().H
		}
	}

	var shareware string
	if !license.IsRegistered() {
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
		Side: ui.E,
	})

	// Set the initial good frame size to have the height secured,
	// so when resizing the application window we can just adjust for width.
	frame.Resize(render.Rect{
		W: d.width,
		H: labelHeight + frame.BoxThickness(1),
	})
	frame.Compute(d.Engine)

	return frame
}
