package doodle

import (
	"git.kirsle.net/apps/doodle/pkg/enum"
	"git.kirsle.net/apps/doodle/pkg/level"
	"git.kirsle.net/apps/doodle/pkg/log"
	"git.kirsle.net/apps/doodle/pkg/uix"
	"git.kirsle.net/apps/doodle/pkg/windows"
	"git.kirsle.net/go/render"
	"git.kirsle.net/go/render/event"
	"git.kirsle.net/go/ui"
)

/*
MenuScene holds the main dialog menu UIs for:

* New Level
* Open Level
* Settings
*/
type MenuScene struct {
	// Configuration.
	StartupMenu string

	Supervisor *ui.Supervisor

	// Private widgets.
	window *ui.Window

	// Background wallpaper canvas.
	canvas *uix.Canvas

	// Values for the New menu
	newPageType  string
	newWallpaper string

	// Values for the Load/Play menu.
	loadForPlay bool // false = load for edit
}

// Name of the scene.
func (s *MenuScene) Name() string {
	return "Menu"
}

// DebugGetWindow surfaces the underlying private window.
func (s *MenuScene) DebugGetWindow() *ui.Window {
	return s.window
}

// GotoNewMenu loads the MenuScene and shows the "New" window.
func (d *Doodle) GotoNewMenu() {
	log.Info("Loading the MenuScene to the New window")
	scene := &MenuScene{
		StartupMenu: "new",
	}
	d.Goto(scene)
}

// GotoLoadMenu loads the MenuScene and shows the "Load" window.
func (d *Doodle) GotoLoadMenu() {
	log.Info("Loading the MenuScene to the Load window for Edit Mode")
	scene := &MenuScene{
		StartupMenu: "load",
	}
	d.Goto(scene)
}

// GotoPlayMenu loads the MenuScene and shows the "Load" window for playing a
// level, not editing it.
func (d *Doodle) GotoPlayMenu() {
	log.Info("Loading the MenuScene to the Load window for Play Mode")
	scene := &MenuScene{
		StartupMenu: "load",
		loadForPlay: true,
	}
	d.Goto(scene)
}

// GotoSettingsMenu loads the settings screen.
func (d *Doodle) GotoSettingsMenu() {
	log.Info("Loading the MenuScene to the Settings Menu")
	scene := &MenuScene{
		StartupMenu: "settings",
	}
	d.Goto(scene)
}

// Setup the scene.
func (s *MenuScene) Setup(d *Doodle) error {
	s.Supervisor = ui.NewSupervisor()

	// Set up the background wallpaper canvas.
	s.canvas = uix.NewCanvas(100, false)
	s.canvas.Resize(render.Rect{
		W: d.width,
		H: d.height,
	})
	s.canvas.LoadLevel(&level.Level{
		Chunker:   level.NewChunker(100),
		Palette:   level.NewPalette(),
		PageType:  level.Bounded,
		Wallpaper: "notebook.png",
	})

	switch s.StartupMenu {
	case "new":
		if err := s.setupNewWindow(d); err != nil {
			return err
		}
	case "load":
		if err := s.setupLoadWindow(d); err != nil {
			return err
		}
	case "settings":
		if err := s.setupSettingsWindow(d); err != nil {
			return err
		}
	default:
		d.FlashError("No Valid StartupMenu Given to MenuScene")
	}

	// Whatever window we got, give it window manager controls under Supervisor.
	s.window.Supervise(s.Supervisor)
	s.window.Compute(d.Engine)

	// Center the window.
	s.window.MoveTo(render.Point{
		X: (d.width / 2) - (s.window.Size().W / 2),
		Y: 60,
	})

	return nil
}

// configureCanvas updates the settings of the background canvas, so a live
// preview of the wallpaper and wrapping type can be shown.
func (s *MenuScene) configureCanvas(pageType level.PageType, wallpaper string) {
	s.canvas.LoadLevel(&level.Level{
		Chunker:   level.NewChunker(100),
		Palette:   level.NewPalette(),
		PageType:  pageType,
		Wallpaper: wallpaper,
	})
}

// setupNewWindow sets up the UI for the "New" window.
func (s *MenuScene) setupNewWindow(d *Doodle) error {
	window := windows.NewAddEditLevel(windows.AddEditLevel{
		Supervisor: s.Supervisor,
		Engine:     d.Engine,
		OnChangePageTypeAndWallpaper: func(pageType level.PageType, wallpaper string) {
			log.Info("OnChangePageTypeAndWallpaper called: %+v, %+v", pageType, wallpaper)
			s.configureCanvas(pageType, wallpaper)
		},
		OnCreateNewLevel: func(lvl *level.Level) {
			d.Goto(&EditorScene{
				DrawingType: enum.LevelDrawing,
				Level:       lvl,
			})
		},
		OnCancel: func() {
			d.Goto(&MainScene{})
		},
	})
	s.window = window
	window.SetButtons(0)
	window.Show()
	return nil
}

// setupLoadWindow sets up the UI for the "New" window.
func (s *MenuScene) setupLoadWindow(d *Doodle) error {
	window := windows.NewOpenLevelEditor(windows.OpenLevelEditor{
		Supervisor:  s.Supervisor,
		Engine:      d.Engine,
		LoadForPlay: s.loadForPlay,
		OnPlayLevel: func(filename string) {
			d.PlayLevel(filename)
		},
		OnEditLevel: func(filename string) {
			d.EditFile(filename)
		},
		OnCancel: func() {
			d.Goto(&MainScene{})
		},
	})
	s.window = window
	return nil
}

// setupLoadWindow sets up the UI for the "New" window.
func (s *MenuScene) setupSettingsWindow(d *Doodle) error {
	window := windows.NewSettingsWindow(windows.Settings{
		Supervisor: s.Supervisor,
		Engine:     d.Engine,
	})
	window.SetButtons(0)
	s.window = window
	return nil
}

// Loop the editor scene.
func (s *MenuScene) Loop(d *Doodle, ev *event.State) error {
	s.Supervisor.Loop(ev)

	if ev.WindowResized {
		w, h := d.Engine.WindowSize()
		d.width = w
		d.height = h
		log.Info("Resized to %dx%d", d.width, d.height)
		s.canvas.Resize(render.Rect{
			W: d.width,
			H: d.height,
		})
	}

	return nil
}

// Draw the pixels on this frame.
func (s *MenuScene) Draw(d *Doodle) error {
	// Draw the background canvas.
	s.canvas.Present(d.Engine, render.Origin)

	// TODO: if I don't call Compute here, buttons in the Edit Window get all
	// bunched up. Investigate why later.
	s.window.Compute(d.Engine)

	// Draw the window managed by Supervisor.
	s.Supervisor.Present(d.Engine)

	return nil
}

// Destroy the scene.
func (s *MenuScene) Destroy() error {
	return nil
}
