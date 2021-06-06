package windows

import (
	"git.kirsle.net/apps/doodle/pkg/balance"
	"git.kirsle.net/apps/doodle/pkg/level"
	"git.kirsle.net/apps/doodle/pkg/modal"
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
		paletteName  = level.DefaultPaletteNames[0]
		isNewLevel   = config.EditLevel == nil
		title        = "New Drawing"

		textCurrentPalette = "Keep current palette"
	)

	// Given a level to edit?
	if config.EditLevel != nil {
		newPageType = config.EditLevel.PageType.String()
		newWallpaper = config.EditLevel.Wallpaper
		paletteName = textCurrentPalette
		title = "Page Settings"
	}

	window := ui.NewWindow(title)
	window.SetButtons(ui.CloseButton)
	window.Configure(ui.Config{
		Width:      400,
		Height:     240,
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
			PadY: 2,
		})

		wpFrame := ui.NewFrame("Wallpaper Frame")
		frame.Pack(wpFrame, ui.Pack{
			Side:  ui.N,
			FillX: true,
			PadY: 2,
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

				// If a new level, the Blueprint button has a tooltip
				// hinting to pick the Blueprint palette to match.
				if config.EditLevel == nil && t.Name == "Blueprint" {
					ui.NewTooltip(radio, ui.Tooltip{
						Text: "Dark theme! Make sure to also\npick a Blueprint color palette!",
						Edge: ui.Top,
					})
				}
			}(t)
		}

		/******************
		 * Frame for picking a default color palette.
		 ******************/

		// For new level or --experimental only.
		if config.EditLevel == nil || balance.Feature.ChangePalette {
			palFrame := ui.NewFrame("Palette Frame")
			frame.Pack(palFrame, ui.Pack{
				Side:  ui.N,
				FillX: true,
				PadY: 4,
			})

			label3 := ui.NewLabel(ui.Label{
				Text: "Palette: ",
				Font: balance.LabelFont,
			})
			palFrame.Pack(label3, ui.Pack{
				Side:  ui.W,
			})

			palBtn := ui.NewMenuButton("Palette Button", ui.NewLabel(ui.Label{
				TextVariable: &paletteName,
				Font: balance.MenuFont,
			}))

			palFrame.Pack(palBtn, ui.Pack{
				Side: ui.W,
				// FillX: true,
				Expand: true,
			})

			if config.EditLevel != nil {
				palBtn.AddItem(paletteName, func() {
					paletteName = textCurrentPalette
				})
				palBtn.AddSeparator();
			}

			for _, palName := range level.DefaultPaletteNames {
				palName := palName
				// palette := level.DefaultPalettes[palName]
				palBtn.AddItem(palName, func() {
					paletteName = palName
				})
			}

			config.Supervisor.Add(palBtn)
			palBtn.Supervise(config.Supervisor)
		}

		/******************
		 * Frame for giving the level a title.
		 ******************/

		if config.EditLevel != nil {
			label3 := ui.NewLabel(ui.Label{
				Text: "Metadata",
				Font: balance.LabelFont,
			})
			frame.Pack(label3, ui.Pack{
				Side:  ui.N,
				FillX: true,
			})

			type metadataObj struct {
				Label   string
				Binding *string
				Update  func(string)
			}
			var metaRows = []metadataObj{
				{"Title:", &config.EditLevel.Title, func(v string) { config.EditLevel.Title = v }},
				{"Author:", &config.EditLevel.Author, func(v string) { config.EditLevel.Author = v }},
			}

			for _, mr := range metaRows {
				mr := mr
				mrFrame := ui.NewFrame("Metadata " + mr.Label + "Frame")
				frame.Pack(mrFrame, ui.Pack{
					Side:  ui.N,
					FillX: true,
					PadY:  2,
				})

				// The label.
				mrLabel := ui.NewLabel(ui.Label{
					Text: mr.Label,
					Font: balance.MenuFont,
				})
				mrLabel.Configure(ui.Config{
					Width: 75,
				})
				mrFrame.Pack(mrLabel, ui.Pack{
					Side: ui.W,
				})

				// The button.
				mrButton := ui.NewButton(mr.Label, ui.NewLabel(ui.Label{
					TextVariable: mr.Binding,
					Font:         balance.MenuFont,
				}))
				mrButton.Handle(ui.Click, func(ed ui.EventData) error {
					shmem.Prompt("Enter a new "+mr.Label, func(answer string) {
						if answer != "" {
							mr.Update(answer)
						}
					})
					return nil
				})
				config.Supervisor.Add(mrButton)
				mrFrame.Pack(mrButton, ui.Pack{
					Side:   ui.W,
					Expand: true,
					PadX:   2,
				})
			}
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
				lvl.Palette = level.DefaultPalettes[paletteName]
				lvl.Wallpaper = newWallpaper
				lvl.PageType = pageType

				config.OnCreateNewLevel(lvl)
				return nil
			}},

			{"Cancel", func(ed ui.EventData) error {
				config.OnCancel()
				return nil
			}},

			// OK button is for editing an existing level.
			{"OK", func(ed ui.EventData) error {
				// If we're editing a level, did we select a new palette?
				if paletteName != textCurrentPalette {
					modal.Confirm(
						"Are you sure you want to change the level palette?\n"+
						"Existing pixels drawn on your level may change, and\n"+
						"if the new palette is smaller, some pixels may be\n"+
						"lost from your level. OK to continue?",
					).WithTitle("Change Level Palette").Then(func() {
						config.OnCancel();
					})
					return nil
				}

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
