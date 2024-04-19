package windows

import (
	"fmt"
	"sort"
	"strings"

	"git.kirsle.net/SketchyMaze/doodle/pkg/balance"
	"git.kirsle.net/SketchyMaze/doodle/pkg/enum"
	"git.kirsle.net/SketchyMaze/doodle/pkg/level"
	"git.kirsle.net/SketchyMaze/doodle/pkg/log"
	"git.kirsle.net/SketchyMaze/doodle/pkg/native"
	"git.kirsle.net/SketchyMaze/doodle/pkg/userdir"
	"git.kirsle.net/go/render"
	"git.kirsle.net/go/ui"
)

// OpenLevelEditor is the "Open a Level to Edit It" window
//
// DEPRECATED: in favor of OpenDrawing which has a nicer listbox instead
// of the ugly button grid of this window.
//
// You can invoke this old window by the shell commands
// `$ d.GotoPlayMenu()` or `$ d.GotoLoadMenu()`
type OpenLevelEditor struct {
	Supervisor *ui.Supervisor
	Engine     render.Engine

	// Load it for playing instead of editing?
	LoadForPlay bool

	// Callback functions.
	OnPlayLevel func(filename string)
	OnEditLevel func(filename string)
	OnCancel    func()
}

// NewOpenLevelEditor initializes the window.
func NewOpenLevelEditor(config OpenLevelEditor) *ui.Window {
	var (
		width, height = config.Engine.WindowSize()
		columns       = 4
	)

	// Show fewer columns on smaller devices.
	if width <= enum.ScreenWidthXSmall {
		columns = 1
	} else if width <= enum.ScreenWidthSmall {
		columns = 2
	} else if width <= enum.ScreenWidthMedium {
		columns = 3
	}

	window := ui.NewWindow("Open Drawing")
	window.Configure(ui.Config{
		Width:      int(float64(width) * 0.75),
		Height:     int(float64(height) * 0.75),
		Background: render.Grey,
	})

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

		// Sort them!
		sort.Slice(
			levels, func(i, j int) bool {
				return strings.ToLower(levels[i]) < strings.ToLower(levels[j])
			},
		)

		lvlRow := ui.NewFrame("Level Row 0")
		frame.Pack(lvlRow, ui.Pack{
			Side:  ui.N,
			FillX: true,
			PadY:  1,
		})
		for i, lvl := range levels {
			func(i int, lvl string) {
				btn := ui.NewButton("Level Btn", ui.NewLabel(ui.Label{
					Text: lvl,
					Font: balance.MenuFont.Update(render.Text{
						PadY: 2,
					}),
				}))
				btn.Handle(ui.Click, func(ed ui.EventData) error {
					if config.LoadForPlay {
						config.OnPlayLevel(lvl)
					} else {
						config.OnEditLevel(lvl)
					}
					return nil
				})
				config.Supervisor.Add(btn)
				lvlRow.Pack(btn, ui.Pack{
					Side:   ui.W,
					Expand: true,
					Fill:   true,
				})

				if columns == 1 || i > 0 && (i+1)%columns == 0 {
					lvlRow = ui.NewFrame(fmt.Sprintf("Level Row %d", i))
					frame.Pack(lvlRow, ui.Pack{
						Side:  ui.N,
						FillX: true,
						PadY:  1,
					})
				}
			}(i, lvl)
		}

		// Browse button for local filesystem.
		browseLevelFrame := ui.NewFrame("Browse Level Frame")
		frame.Pack(browseLevelFrame, ui.Pack{
			Side:   ui.N,
			Expand: true,
			FillX:  true,
			PadY:   1,
		})

		browseLevelButton := ui.NewButton("Browse Level", ui.NewLabel(ui.Label{
			Text: "Browse...",
			Font: balance.MenuFont,
		}))
		browseLevelButton.SetStyle(&balance.ButtonPrimary)
		browseLevelFrame.Pack(browseLevelButton, ui.Pack{
			Side: ui.W,
		})

		browseLevelButton.Handle(ui.Click, func(ed ui.EventData) error {
			filename, err := native.OpenFile("Choose a .level file", "*.level")
			if err != nil {
				log.Error("Couldn't show file dialog: %s", err)
				return nil
			}

			if config.LoadForPlay {
				config.OnPlayLevel(filename)
			} else {
				config.OnEditLevel(filename)
			}
			return nil
		})
		config.Supervisor.Add(browseLevelButton)

		/******************
		 * Frame for selecting User Doodads
		 ******************/

		// Doodads not shown if we're loading a map to play.
		if !config.LoadForPlay {
			label2 := ui.NewLabel(ui.Label{
				Text: "Doodads",
				Font: balance.LabelFont,
			})
			frame.Pack(label2, ui.Pack{
				Side:  ui.N,
				FillX: true,
			})

			files, _ := userdir.ListDoodads()

			// Sort them!
			sort.Slice(
				files, func(i, j int) bool {
					return strings.ToLower(files[i]) < strings.ToLower(files[j])
				},
			)

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
					btn.Handle(ui.Click, func(ed ui.EventData) error {
						config.OnEditLevel(dd)
						return nil
					})
					config.Supervisor.Add(btn)
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

			// Browse button for local filesystem.
			browseDoodadFrame := ui.NewFrame("Browse Doodad Frame")
			frame.Pack(browseDoodadFrame, ui.Pack{
				Side:   ui.N,
				Expand: true,
				FillX:  true,
				PadY:   1,
			})

			browseDoodadButton := ui.NewButton("Browse Doodad", ui.NewLabel(ui.Label{
				Text: "Browse...",
				Font: balance.MenuFont,
			}))
			browseDoodadButton.SetStyle(&balance.ButtonPrimary)
			browseDoodadFrame.Pack(browseDoodadButton, ui.Pack{
				Side: ui.W,
			})

			browseDoodadButton.Handle(ui.Click, func(ed ui.EventData) error {
				filename, err := native.OpenFile("Choose a .doodad file", "*.doodad")
				if err != nil {
					log.Error("Couldn't show file dialog: %s", err)
					return nil
				}

				if config.LoadForPlay {
					config.OnPlayLevel(filename)
				} else {
					config.OnEditLevel(filename)
				}
				return nil
			})
			config.Supervisor.Add(browseDoodadButton)
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
			F     func(ui.EventData) error
		}{
			{"Cancel", func(ed ui.EventData) error {
				config.OnCancel()
				return nil
			}},
		}
		for _, t := range buttons {
			btn := ui.NewButton(t.Label, ui.NewLabel(ui.Label{
				Text: t.Label,
				Font: balance.MenuFont,
			}))
			btn.Handle(ui.Click, t.F)
			config.Supervisor.Add(btn)
			bottomFrame.Pack(btn, ui.Pack{
				Side: ui.W,
				PadX: 4,
				PadY: 8,
			})
		}
	}

	return window
}
