package windows

import (
	"git.kirsle.net/apps/doodle/pkg/balance"
	"git.kirsle.net/apps/doodle/pkg/level"
	"git.kirsle.net/apps/doodle/pkg/shmem"
	"git.kirsle.net/go/render"
	"git.kirsle.net/go/ui"
)

// AddEditLevel is the "Create New Level & Edit Level Properties" window
type AddEditLevel struct {
	Supervisor *ui.Supervisor
	Engine     render.Engine

	// Editing settings for an existing level?
	EditLevel *level.Level

	// Callback functions.
	OnChangePageTypeAndWallpaper func(pageType level.PageType, wallpaper string)
	OnCreateNewLevel             func(*level.Level)
	OnCancel                     func()
}

// NewAddEditLevel initializes the window.
func NewAddEditLevel(config AddEditLevel) *ui.Window {
	// Default options.
	var (
		newPageType  = level.Bounded.String()
		newWallpaper = "notebook.png"
		isNewLevel   = config.EditLevel == nil
		title        = "New Drawing"
	)

	// Given a level to edit?
	if config.EditLevel != nil {
		newPageType = config.EditLevel.PageType.String()
		newWallpaper = config.EditLevel.Wallpaper
		title = "Page Settings"
	}

	window := ui.NewWindow(title)
	window.SetButtons(ui.CloseButton)
	window.Configure(ui.Config{
		Width:      400,
		Height:     180,
		Background: render.Grey,
	})

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
					&newPageType,
					t.Value.String(),
					ui.NewLabel(ui.Label{
						Text: t.Name,
						Font: balance.MenuFont,
					}),
				)
				radio.Handle(ui.Click, func(ed ui.EventData) error {
					config.OnChangePageTypeAndWallpaper(t.Value, newWallpaper)
					return nil
				})
				config.Supervisor.Add(radio)
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
				radio := ui.NewRadioButton(t.Name, &newWallpaper, t.Value, ui.NewLabel(ui.Label{
					Text: t.Name,
					Font: balance.MenuFont,
				}))
				radio.Handle(ui.Click, func(ed ui.EventData) error {
					if pageType, ok := level.PageTypeFromString(newPageType); ok {
						config.OnChangePageTypeAndWallpaper(pageType, t.Value)
					}
					return nil
				})
				config.Supervisor.Add(radio)
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
		frame.Pack(bottomFrame, ui.Pack{
			Side:  ui.N,
			FillX: true,
			PadY:  8,
		})

		var buttons = []struct {
			Label string
			F     func(ui.EventData) error
		}{
			{"Continue", func(ed ui.EventData) error {
				shmem.Flash("Create new map with %s page type and %s wallpaper", newPageType, newWallpaper)
				pageType, ok := level.PageTypeFromString(newPageType)
				if !ok {
					shmem.Flash("Invalid Page Type '%s'", newPageType)
					return nil
				}

				lvl := level.New()
				lvl.Palette = level.DefaultPalette()
				lvl.Wallpaper = newWallpaper
				lvl.PageType = pageType

				// Blueprint theme palette for the dark wallpaper color.
				if lvl.Wallpaper == "blueprint.png" {
					lvl.Palette = level.NewBlueprintPalette()
				}

				config.OnCreateNewLevel(lvl)
				return nil
			}},

			{"Cancel", func(ed ui.EventData) error {
				config.OnCancel()
				return nil
			}},

			// OK button is for editing an existing level.
			{"OK", func(ed ui.EventData) error {
				config.OnCancel()
				return nil
			}},
		}
		for _, t := range buttons {
			// If we're editing settings on an existing level, skip the Continue.
			if (isNewLevel && t.Label == "OK") || (!isNewLevel && t.Label != "OK") {
				continue
			}
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

	window.Hide()
	return window
}
