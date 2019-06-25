package doodle

import (
	"git.kirsle.net/apps/doodle/lib/events"
	"git.kirsle.net/apps/doodle/lib/render"
	"git.kirsle.net/apps/doodle/lib/ui"
	"git.kirsle.net/apps/doodle/pkg/balance"
	"git.kirsle.net/apps/doodle/pkg/enum"
	"git.kirsle.net/apps/doodle/pkg/level"
	"git.kirsle.net/apps/doodle/pkg/log"
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

	// Values for the New menu
	newPageType  string
	newWallpaper string
}

// Name of the scene.
func (s *MenuScene) Name() string {
	return "Menu"
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
	log.Info("Loading the MenuScene to the Load window")
	scene := &MenuScene{
		StartupMenu: "load",
	}
	d.Goto(scene)
}

// Setup the scene.
func (s *MenuScene) Setup(d *Doodle) error {
	s.Supervisor = ui.NewSupervisor()

	switch s.StartupMenu {
	case "new":
		if err := s.setupNewWindow(d); err != nil {
			return err
		}
	case "load":
		if err := s.setupLoadWindow(d); err != nil {
			return err
		}
	default:
		d.Flash("No Valid StartupMenu Given to MenuScene")
	}

	return nil
}

// setupNewWindow sets up the UI for the "New" window.
func (s *MenuScene) setupNewWindow(d *Doodle) error {
	// Default scene options.
	s.newPageType = level.Bounded.String()
	s.newWallpaper = "notebook.png"

	window := ui.NewWindow("New Drawing")
	window.Configure(ui.Config{
		Width:      int32(float64(d.width) * 0.8),
		Height:     int32(float64(d.height) * 0.8),
		Background: render.Grey,
	})
	window.Compute(d.Engine)

	{
		frame := ui.NewFrame("New Level Frame")
		window.Pack(frame, ui.Pack{
			Anchor: ui.N,
			Fill:   true,
			Expand: true,
		})

		/******************
		 * Frame for selecting Page Type
		 ******************/

		label1 := ui.NewLabel(ui.Label{
			Text: "Page Type",
			Font: balance.LabelFont,
		})
		frame.Pack(label1, ui.Pack{
			Anchor: ui.N,
			FillX:  true,
		})

		typeFrame := ui.NewFrame("Page Type Options Frame")
		frame.Pack(typeFrame, ui.Pack{
			Anchor: ui.N,
			FillX:  true,
		})

		var types = []struct {
			Name  string
			Value level.PageType
		}{
			{"Unbounded", level.Unbounded},
			{"Bounded", level.Bounded},
			{"No Negative Space", level.NoNegativeSpace},
			{"Bordered", level.Bordered},
		}
		for _, t := range types {
			// Hide some options for the free version of the game.
			if balance.FreeVersion {
				if t.Value != level.Bounded {
					continue
				}
			}

			radio := ui.NewRadioButton(t.Name,
				&s.newPageType,
				t.Value.String(),
				ui.NewLabel(ui.Label{
					Text: t.Name,
					Font: balance.MenuFont,
				}),
			)
			s.Supervisor.Add(radio)
			typeFrame.Pack(radio, ui.Pack{
				Anchor: ui.W,
				PadX:   4,
			})
		}

		/******************
		 * Frame for selecting Level Wallpaper
		 ******************/

		label2 := ui.NewLabel(ui.Label{
			Text: "Wallpaper",
			Font: balance.LabelFont,
		})
		frame.Pack(label2, ui.Pack{
			Anchor: ui.N,
			FillX:  true,
		})

		wpFrame := ui.NewFrame("Wallpaper Frame")
		frame.Pack(wpFrame, ui.Pack{
			Anchor: ui.N,
			FillX:  true,
		})

		var wallpapers = []struct {
			Name  string
			Value string
		}{
			{"Notebook", "notebook.png"},
			{"Blueprint", "blueprint.png"},
			{"Legal Pad", "legal.png"},
			{"Placemat", "placemat.png"},
		}
		for _, t := range wallpapers {
			radio := ui.NewRadioButton(t.Name, &s.newWallpaper, t.Value, ui.NewLabel(ui.Label{
				Text: t.Name,
				Font: balance.MenuFont,
			}))
			s.Supervisor.Add(radio)
			wpFrame.Pack(radio, ui.Pack{
				Anchor: ui.W,
				PadX:   4,
			})
		}

		/******************
		 * Confirm/cancel buttons.
		 ******************/

		bottomFrame := ui.NewFrame("Button Frame")
		// bottomFrame.Configure(ui.Config{
		// 	BorderSize:  1,
		// 	BorderStyle: ui.BorderSunken,
		// 	BorderColor: render.Black,
		// })
		// bottomFrame.SetBackground(render.Grey)
		frame.Pack(bottomFrame, ui.Pack{
			Anchor: ui.N,
			FillX:  true,
			PadY:   8,
		})

		var buttons = []struct {
			Label string
			F     func(render.Point)
		}{
			{"Continue", func(p render.Point) {
				d.Flash("Create new map with %s page type and %s wallpaper", s.newPageType, s.newWallpaper)
				pageType, ok := level.PageTypeFromString(s.newPageType)
				if !ok {
					d.Flash("Invalid Page Type '%s'", s.newPageType)
					return
				}

				lvl := level.New()
				lvl.Palette = level.DefaultPalette()
				lvl.Wallpaper = s.newWallpaper
				lvl.PageType = pageType

				d.Goto(&EditorScene{
					DrawingType: enum.LevelDrawing,
					Level:       lvl,
				})
			}},

			{"Cancel", func(p render.Point) {
				d.Goto(&MainScene{})
			}},
		}
		for _, t := range buttons {
			btn := ui.NewButton(t.Label, ui.NewLabel(ui.Label{
				Text: t.Label,
				Font: balance.MenuFont,
			}))
			btn.Handle(ui.Click, t.F)
			s.Supervisor.Add(btn)
			bottomFrame.Pack(btn, ui.Pack{
				Anchor: ui.W,
				PadX:   4,
				PadY:   8,
			})
		}
	}

	s.window = window
	return nil
}

// setupLoadWindow sets up the UI for the "New" window.
func (s *MenuScene) setupLoadWindow(d *Doodle) error {
	return nil
}

// Loop the editor scene.
func (s *MenuScene) Loop(d *Doodle, ev *events.State) error {
	s.Supervisor.Loop(ev)
	return nil
}

// Draw the pixels on this frame.
func (s *MenuScene) Draw(d *Doodle) error {
	// Clear the canvas and fill it with white.
	d.Engine.Clear(render.White)

	s.window.Compute(d.Engine)
	s.window.MoveTo(render.Point{
		X: (int32(d.width) / 2) - (s.window.Size().W / 2),
		Y: 60,
	})
	s.window.Present(d.Engine, s.window.Point())

	return nil
}

// Destroy the scene.
func (s *MenuScene) Destroy() error {
	return nil
}
