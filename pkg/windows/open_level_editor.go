package windows

import (
	"fmt"

	"git.kirsle.net/apps/doodle/pkg/balance"
	"git.kirsle.net/apps/doodle/pkg/level"
	"git.kirsle.net/apps/doodle/pkg/userdir"
	"git.kirsle.net/go/render"
	"git.kirsle.net/go/ui"
)

// OpenLevelEditor is the "Open a Level to Edit It" window
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
	)

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
					Font: balance.MenuFont,
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

				if i > 0 && (i+1)%4 == 0 {
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
		if !config.LoadForPlay && !balance.FreeVersion {
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
