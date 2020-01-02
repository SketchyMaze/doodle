package doodle

import (
	"fmt"

	"git.kirsle.net/apps/doodle/pkg/balance"
	"git.kirsle.net/apps/doodle/pkg/enum"
	"git.kirsle.net/apps/doodle/pkg/level"
	"git.kirsle.net/apps/doodle/pkg/log"
	"git.kirsle.net/apps/doodle/pkg/uix"
	"git.kirsle.net/apps/doodle/pkg/userdir"
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

// Setup the scene.
func (s *MenuScene) Setup(d *Doodle) error {
	s.Supervisor = ui.NewSupervisor()

	// Set up the background wallpaper canvas.
	s.canvas = uix.NewCanvas(100, false)
	s.canvas.Resize(render.Rect{
		W: d.width,
		H: d.height,
	})
	s.canvas.LoadLevel(d.Engine, &level.Level{
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
	default:
		d.Flash("No Valid StartupMenu Given to MenuScene")
	}

	return nil
}

// configureCanvas updates the settings of the background canvas, so a live
// preview of the wallpaper and wrapping type can be shown.
func (s *MenuScene) configureCanvas(e render.Engine, pageType level.PageType, wallpaper string) {
	s.canvas.LoadLevel(e, &level.Level{
		Chunker:   level.NewChunker(100),
		Palette:   level.NewPalette(),
		PageType:  pageType,
		Wallpaper: wallpaper,
	})
}

// setupNewWindow sets up the UI for the "New" window.
func (s *MenuScene) setupNewWindow(d *Doodle) error {
	// Default scene options.
	s.newPageType = level.Bounded.String()
	s.newWallpaper = "notebook.png"

	window := ui.NewWindow("New Drawing")
	window.Configure(ui.Config{
		Width:      int(float64(d.width) * 0.75),
		Height:     int(float64(d.height) * 0.75),
		Background: render.Grey,
	})
	window.Compute(d.Engine)

	{
		frame := ui.NewFrame("New Level Frame")
		window.Pack(frame, ui.Pack{
			Side:   ui.N,
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
			Side:  ui.N,
			FillX: true,
		})

		typeFrame := ui.NewFrame("Page Type Options Frame")
		frame.Pack(typeFrame, ui.Pack{
			Side:  ui.N,
			FillX: true,
		})

		type typeObj struct {
			Name  string
			Value level.PageType
		}
		var types = []typeObj{
			{"Unbounded", level.Unbounded},
			{"Bounded", level.Bounded},
			{"No Negative Space", level.NoNegativeSpace},
			// {"Bordered (TODO)", level.Bordered},
		}
		for _, t := range types {
			// TODO: Hide some options for the free version of the game.
			// - At launch only Bounded and Bordered will be available
			//   in the shareware version.
			// - For now, only hide Bordered as it's not yet implemented.
			// --------
			// if balance.FreeVersion {
			// 	if t.Value == level.Bordered {
			// 		continue
			// 	}
			// }

			func(t typeObj) {
				radio := ui.NewRadioButton(t.Name,
					&s.newPageType,
					t.Value.String(),
					ui.NewLabel(ui.Label{
						Text: t.Name,
						Font: balance.MenuFont,
					}),
				)
				radio.Handle(ui.Click, func(p render.Point) {
					s.configureCanvas(d.Engine, t.Value, s.newWallpaper)
				})
				s.Supervisor.Add(radio)
				typeFrame.Pack(radio, ui.Pack{
					Side: ui.W,
					PadX: 4,
				})
			}(t)
		}

		/******************
		 * Frame for selecting Level Wallpaper
		 ******************/

		label2 := ui.NewLabel(ui.Label{
			Text: "Wallpaper",
			Font: balance.LabelFont,
		})
		frame.Pack(label2, ui.Pack{
			Side:  ui.N,
			FillX: true,
		})

		wpFrame := ui.NewFrame("Wallpaper Frame")
		frame.Pack(wpFrame, ui.Pack{
			Side:  ui.N,
			FillX: true,
		})

		type wallpaperObj struct {
			Name  string
			Value string
		}
		var wallpapers = []wallpaperObj{
			{"Notebook", "notebook.png"},
			{"Blueprint", "blueprint.png"},
			{"Legal Pad", "legal.png"},
			{"Pure White", "white.png"},
			// {"Placemat", "placemat.png"},
		}
		for _, t := range wallpapers {
			func(t wallpaperObj) {
				radio := ui.NewRadioButton(t.Name, &s.newWallpaper, t.Value, ui.NewLabel(ui.Label{
					Text: t.Name,
					Font: balance.MenuFont,
				}))
				radio.Handle(ui.Click, func(p render.Point) {
					log.Info("Set wallpaper to %s", t.Value)
					if pageType, ok := level.PageTypeFromString(s.newPageType); ok {
						s.configureCanvas(d.Engine, pageType, t.Value)
					}
				})
				s.Supervisor.Add(radio)
				wpFrame.Pack(radio, ui.Pack{
					Side: ui.W,
					PadX: 4,
				})
			}(t)
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
			Side:  ui.N,
			FillX: true,
			PadY:  8,
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

				// Blueprint theme palette for the dark wallpaper color.
				if lvl.Wallpaper == "blueprint.png" {
					lvl.Palette = level.NewBlueprintPalette()
				}

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
				Side: ui.W,
				PadX: 4,
				PadY: 8,
			})
		}
	}

	s.window = window
	return nil
}

// setupLoadWindow sets up the UI for the "New" window.
func (s *MenuScene) setupLoadWindow(d *Doodle) error {
	window := ui.NewWindow("Open Drawing")
	window.Configure(ui.Config{
		Width:      int(float64(d.width) * 0.75),
		Height:     int(float64(d.height) * 0.75),
		Background: render.Grey,
	})
	window.Compute(d.Engine)

	{
		frame := ui.NewFrame("Open Drawing Frame")
		window.Pack(frame, ui.Pack{
			Side:   ui.N,
			Fill:   true,
			Expand: true,
		})

		/******************
		 * Frame for selecting User Levels
		 ******************/

		label1 := ui.NewLabel(ui.Label{
			Text: "Levels",
			Font: balance.LabelFont,
		})
		frame.Pack(label1, ui.Pack{
			Side:  ui.N,
			FillX: true,
		})

		// Get the user's levels.
		levels, _ := userdir.ListLevels()

		// Embedded levels, TODO
		sysLevels, _ := level.ListSystemLevels()
		levels = append(levels, sysLevels...)

		lvlRow := ui.NewFrame("Level Row 0")
		frame.Pack(lvlRow, ui.Pack{
			Side:  ui.N,
			FillX: true,
			PadY:  1,
		})
		for i, lvl := range levels {
			func(i int, lvl string) {
				log.Info("Add file %s to row %s", lvl, lvlRow.Name)
				btn := ui.NewButton("Level Btn", ui.NewLabel(ui.Label{
					Text: lvl,
					Font: balance.MenuFont,
				}))
				btn.Handle(ui.Click, func(p render.Point) {
					if s.loadForPlay {
						d.PlayLevel(lvl)
					} else {
						d.EditFile(lvl)
					}
				})
				s.Supervisor.Add(btn)
				lvlRow.Pack(btn, ui.Pack{
					Side:   ui.W,
					Expand: true,
					Fill:   true,
				})

				if i > 0 && (i+1)%4 == 0 {
					log.Warn("i=%d wrapped at mod 4", i)
					lvlRow = ui.NewFrame(fmt.Sprintf("Level Row %d", i))
					frame.Pack(lvlRow, ui.Pack{
						Side:  ui.N,
						FillX: true,
						PadY:  1,
					})
				}
			}(i, lvl)
		}

		/******************
		 * Frame for selecting User Doodads
		 ******************/

		// Doodads not shown if we're loading a map to play, nor are they
		// available to the free version.
		if !s.loadForPlay && !balance.FreeVersion {
			label2 := ui.NewLabel(ui.Label{
				Text: "Doodads",
				Font: balance.LabelFont,
			})
			frame.Pack(label2, ui.Pack{
				Side:  ui.N,
				FillX: true,
			})

			files, _ := userdir.ListDoodads()
			ddRow := ui.NewFrame("Doodad Row 0")
			frame.Pack(ddRow, ui.Pack{
				Side:  ui.N,
				FillX: true,
				PadY:  1,
			})
			for i, dd := range files {
				func(i int, dd string) {
					btn := ui.NewButton("Doodad Btn", ui.NewLabel(ui.Label{
						Text: dd,
						Font: balance.MenuFont,
					}))
					btn.Handle(ui.Click, func(p render.Point) {
						d.EditFile(dd)
					})
					s.Supervisor.Add(btn)
					ddRow.Pack(btn, ui.Pack{
						Side:   ui.W,
						Expand: true,
						Fill:   true,
					})

					if i > 0 && (i+1)%4 == 0 {
						ddRow = ui.NewFrame(fmt.Sprintf("Doodad Row %d", i))
						frame.Pack(ddRow, ui.Pack{
							Side:  ui.N,
							FillX: true,
							PadY:  1,
						})
					}
				}(i, dd)
			}
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
			Side:  ui.N,
			FillX: true,
			PadY:  8,
		})

		var buttons = []struct {
			Label string
			F     func(render.Point)
		}{
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
				Side: ui.W,
				PadX: 4,
				PadY: 8,
			})
		}
	}

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
	// Clear the canvas and fill it with white.
	d.Engine.Clear(render.White)

	// Draw the background canvas.
	s.canvas.Present(d.Engine, render.Origin)

	s.window.Compute(d.Engine)
	s.window.MoveTo(render.Point{
		X: (d.width / 2) - (s.window.Size().W / 2),
		Y: 60,
	})
	s.window.Present(d.Engine, s.window.Point())

	return nil
}

// Destroy the scene.
func (s *MenuScene) Destroy() error {
	return nil
}
