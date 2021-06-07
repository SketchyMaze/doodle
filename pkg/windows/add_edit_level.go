package windows

import (
	"git.kirsle.net/apps/doodle/pkg/balance"
	"git.kirsle.net/apps/doodle/pkg/level"
	"git.kirsle.net/apps/doodle/pkg/modal"
	"git.kirsle.net/apps/doodle/pkg/native"
	"git.kirsle.net/apps/doodle/pkg/shmem"
	"git.kirsle.net/apps/doodle/pkg/wallpaper"
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

		// Default text for the Palette drop-down for already-existing levels.
		// (needs --experimental feature flag to enable the UI).
		textCurrentPalette = "Keep current palette"

		// For NEW levels, if a custom wallpaper is selected from disk, cache
		// it in these vars. For pre-existing levels, the wallpaper updates
		// immediately in the live config.EditLevel object.
		newWallpaperB64 string
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

		typeFrame := ui.NewFrame("Page Type Options Frame")
		frame.Pack(typeFrame, ui.Pack{
			Side:  ui.N,
			FillX: true,
		})

		label1 := ui.NewLabel(ui.Label{
			Text: "Page Type:",
			Font: balance.LabelFont,
		})
		typeFrame.Pack(label1, ui.Pack{
			Side:  ui.W,
		})

		type typeObj struct {
			Name  string
			Value level.PageType
		}
		var types = []typeObj{
			{"Bounded", level.Bounded},
			{"Unbounded", level.Unbounded},
			{"No Negative Space", level.NoNegativeSpace},
			// {"Bordered (TODO)", level.Bordered},
		}

		typeBtn := ui.NewSelectBox("Type Select", ui.Label{
			Font: ui.MenuFont,
		})
		typeFrame.Pack(typeBtn, ui.Pack{
			Side: ui.W,
			Expand: true,
		})

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
			typeBtn.AddItem(t.Name, t.Value, func() {})
		}

		// If editing an existing level, pre-select the right page type.
		if config.EditLevel != nil {
			typeBtn.SetValue(config.EditLevel.PageType)
		}

		typeBtn.Handle(ui.Change, func(ed ui.EventData) error {
			if selection, ok := typeBtn.GetValue(); ok {
				if pageType, ok := selection.Value.(level.PageType); ok {
					newPageType = pageType.String()
					config.OnChangePageTypeAndWallpaper(pageType, newWallpaper)
				}
			}
			return nil
		})

		typeBtn.Supervise(config.Supervisor)
		config.Supervisor.Add(typeBtn)

		/******************
		 * Frame for selecting Level Wallpaper
		 ******************/

		wpFrame := ui.NewFrame("Wallpaper Frame")
		frame.Pack(wpFrame, ui.Pack{
			Side:  ui.N,
			FillX: true,
			PadY: 2,
		})

		label2 := ui.NewLabel(ui.Label{
			Text: "Wallpaper:",
			Font: balance.LabelFont,
		})
		wpFrame.Pack(label2, ui.Pack{
			Side:  ui.W,
			PadY: 2,
		})

		type wallpaperObj struct {
			Name  string
			Value string
		}
		var wallpapers = []wallpaperObj{
			{"Notebook", "notebook.png"},
			{"Legal Pad", "legal.png"},
			{"Graph paper", "graph.png"},
			{"Dotted paper", "dots.png"},
			{"Blueprint", "blueprint.png"},
			{"Pure White", "white.png"},
			// {"Placemat", "placemat.png"},
		}

		wallBtn := ui.NewSelectBox("Wallpaper Select", ui.Label{
			Font: balance.MenuFont,
		})
		wallBtn.AlwaysChange = true
		wpFrame.Pack(wallBtn, ui.Pack{
			Side: ui.W,
			Expand: true,
		})

		for _, t := range wallpapers {
			wallBtn.AddItem(t.Name, t.Value, func() {})
		}

		// Add custom wallpaper options.
		if balance.Feature.CustomWallpaper {
			wallBtn.AddSeparator()
			wallBtn.AddItem("Custom wallpaper...", balance.CustomWallpaperFilename, func() {})
		}

		// If editing a level, select the current wallpaper.
		if config.EditLevel != nil {
			wallBtn.SetValue(config.EditLevel.Wallpaper)
		}

		wallBtn.Handle(ui.Change, func(ed ui.EventData) error {
			if selection, ok := wallBtn.GetValue(); ok {
				if filename, ok := selection.Value.(string); ok {
					// Picking the Custom option?
					if filename == balance.CustomWallpaperFilename {
						filename, err := native.OpenFile("Choose a custom wallpaper:", "*.png *.jpg *.gif")
						if err == nil {
							b64data, err := wallpaper.FileToB64(filename)
							if err != nil {
								shmem.Flash("Error loading wallpaper: %s", err)
								return nil
							}

							// If editing a level, apply the update straight away.
							if config.EditLevel != nil {
								config.EditLevel.SetFile(balance.CustomWallpaperEmbedPath, []byte(b64data))
								newWallpaper = balance.CustomWallpaperFilename

								// Trigger the page type change to the caller.
								if pageType, ok := level.PageTypeFromString(newPageType); ok {
									config.OnChangePageTypeAndWallpaper(pageType, balance.CustomWallpaperFilename)
								}
							} else {
								// Hold onto the new wallpaper until the level is created.
								newWallpaper = balance.CustomWallpaperFilename
								newWallpaperB64 = b64data
							}
						}
						return nil
					}

					if pageType, ok := level.PageTypeFromString(newPageType); ok {
						config.OnChangePageTypeAndWallpaper(pageType, filename)
						newWallpaper = filename
					}
				}
			}
			return nil
		})

		wallBtn.Supervise(config.Supervisor)
		config.Supervisor.Add(wallBtn)

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

			palBtn := ui.NewSelectBox("Palette Select", ui.Label{
				Font: balance.MenuFont,
			})
			palBtn.AlwaysChange = true

			palFrame.Pack(palBtn, ui.Pack{
				Side: ui.W,
				Expand: true,
			})

			if config.EditLevel != nil {
				palBtn.AddItem(paletteName, paletteName, func() {})
				palBtn.AddSeparator();
			}

			for _, palName := range level.DefaultPaletteNames {
				palName := palName
				palBtn.AddItem(palName, palName, func() {})
			}

			palBtn.Handle(ui.Change, func(ed ui.EventData) error {
				if val, ok := palBtn.GetValue(); ok {
					val2, _ := val.Value.(string)
					paletteName = val2
				}
				return nil
			})

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

				// Was a custom wallpaper selected for our NEW level?
				if lvl.Wallpaper == balance.CustomWallpaperFilename && len(newWallpaperB64) > 0 {
					lvl.SetFile(balance.CustomWallpaperEmbedPath, []byte(newWallpaperB64))
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
