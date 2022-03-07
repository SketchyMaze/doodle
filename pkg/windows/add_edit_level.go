package windows

import (
	"strconv"

	"git.kirsle.net/apps/doodle/pkg/balance"
	"git.kirsle.net/apps/doodle/pkg/enum"
	"git.kirsle.net/apps/doodle/pkg/level"
	"git.kirsle.net/apps/doodle/pkg/log"
	"git.kirsle.net/apps/doodle/pkg/modal"
	"git.kirsle.net/apps/doodle/pkg/native"
	"git.kirsle.net/apps/doodle/pkg/shmem"
	magicform "git.kirsle.net/apps/doodle/pkg/uix/magic-form"
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
	OnCreateNewDoodad            func(size int)
	OnReload                     func()
	OnCancel                     func()
}

// NewAddEditLevel initializes the window.
func NewAddEditLevel(config AddEditLevel) *ui.Window {
	// Default options.
	var (
		title = "New Drawing"
	)

	// Given a level to edit?
	if config.EditLevel != nil {
		title = "Level Properties"
	}

	window := ui.NewWindow(title)
	window.SetButtons(ui.CloseButton)
	window.Configure(ui.Config{
		Width:      400,
		Height:     280,
		Background: render.Grey,
	})

	// Tabbed UI for New Level or New Doodad.
	tabframe := ui.NewTabFrame("Level Tabs")
	if config.EditLevel != nil {
		tabframe.SetTabsHidden(true)
	}
	window.Pack(tabframe, ui.Pack{
		Side:   ui.N,
		Fill:   true,
		Expand: true,
	})

	// Add the tabs.
	config.setupLevelFrame(tabframe)
	config.setupDoodadFrame(tabframe)

	tabframe.Supervise(config.Supervisor)

	window.Hide()
	return window
}

// Creates the Create/Edit Level tab ("index").
func (config AddEditLevel) setupLevelFrame(tf *ui.TabFrame) {
	// Default options.
	var (
		newPageType  = level.Bounded.String()
		newWallpaper = "notebook.png"
		paletteName  = level.DefaultPaletteNames[0]
		isNewLevel   = config.EditLevel == nil

		// Default text for the Palette drop-down for already-existing levels.
		// (needs --experimental feature flag to enable the UI).
		textCurrentPalette = "Keep current palette"

		// For NEW levels, if a custom wallpaper is selected from disk, cache
		// it in these vars. For pre-existing levels, the wallpaper updates
		// immediately in the live config.EditLevel object.
		newWallpaperB64 string
	)

	// Given a level to edit?
	if !isNewLevel {
		newPageType = config.EditLevel.PageType.String()
		newWallpaper = config.EditLevel.Wallpaper
		paletteName = textCurrentPalette
	}

	frame := tf.AddTab("index", ui.NewLabel(ui.Label{
		Text: "New Level",
		Font: balance.TabFont,
	}))

	/******************
	 * Frame for selecting Page Type
	 ******************/

	// Selected "Page Type" property.
	var pageType = level.Bounded
	if !isNewLevel {
		pageType = config.EditLevel.PageType
	}

	form := magicform.Form{
		Supervisor: config.Supervisor,
		Engine:     config.Engine,
		Vertical:   true,
		LabelWidth: 120,
		PadY:       2,
	}
	fields := []magicform.Field{
		{
			Label: "Page type:",
			Font:  balance.UIFont,
			Options: []magicform.Option{
				{
					Label: "Bounded",
					Value: level.Bounded,
				},
				{
					Label: "Bounded",
					Value: level.Bounded,
				},
				{
					Label: "Unbounded",
					Value: level.Unbounded,
				},
				{
					Label: "No Negative Space",
					Value: level.NoNegativeSpace,
				},
			},
			SelectValue: pageType,
			OnSelect: func(v interface{}) {
				value, _ := v.(level.PageType)
				newPageType = value.String() // for the "New" screen background
				config.OnChangePageTypeAndWallpaper(value, newWallpaper)
			},
		},
	}

	/******************
	 * Wallpaper settings
	 ******************/

	var selectedWallpaper = "notebook.png"
	if config.EditLevel != nil {
		selectedWallpaper = config.EditLevel.Wallpaper
	}

	fields = append(fields, []magicform.Field{
		{
			Label:       "Wallpaper:",
			Font:        balance.UIFont,
			SelectValue: selectedWallpaper,
			Options: []magicform.Option{
				{
					Label: "Notebook",
					Value: "notebook.png",
				},
				{
					Label: "Legal Pad",
					Value: "legal.png",
				},
				{
					Label: "Graph paper",
					Value: "graph.png",
				},
				{
					Label: "Dotted paper",
					Value: "dots.png",
				},
				{
					Label: "Blueprint",
					Value: "blueprint.png",
				},
				{
					Label: "Pure white",
					Value: "white.png",
				},
				{
					Separator: true,
				},
				{
					Label: "Custom wallpaper...",
					Value: balance.CustomWallpaperFilename,
				},
			},
			OnSelect: func(v interface{}) {
				if filename, ok := v.(string); ok {
					// Picking the Custom option?
					if filename == balance.CustomWallpaperFilename {
						filename, err := native.OpenFile("Choose a custom wallpaper:", "*.png *.jpg *.gif")
						if err == nil {
							b64data, err := wallpaper.FileToB64(filename)
							if err != nil {
								shmem.Flash("Error loading wallpaper: %s", err)
								return
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
						return
					}

					if pageType, ok := level.PageTypeFromString(newPageType); ok {
						config.OnChangePageTypeAndWallpaper(pageType, filename)
						newWallpaper = filename
					}
				}
			},
		},
	}...)

	/******************
	 * Frame for picking a default color palette.
	 ******************/

	// For new level or --experimental only.
	if config.EditLevel == nil || balance.Feature.ChangePalette {
		var (
			palettes = []magicform.Option{}
		)

		if config.EditLevel != nil {
			palettes = append(palettes, []magicform.Option{
				{
					Label: paletteName, // "Keep current palette"
					Value: paletteName,
				},
				{
					Separator: true,
				},
			}...)
		}

		for _, palName := range level.DefaultPaletteNames {
			palettes = append(palettes, magicform.Option{
				Label: palName,
				Value: palName,
			})
		}

		// Add form fields.
		fields = append(fields, []magicform.Field{
			{
				Label:   "Palette:",
				Font:    balance.UIFont,
				Options: palettes,
				OnSelect: func(v interface{}) {
					value, _ := v.(string)
					paletteName = value
				},
			},
		}...)
	}

	/******************
	 * Extended options for editing existing level (vs. Create New screen)
	 ******************/

	if config.EditLevel != nil {
		fields = append(fields, []magicform.Field{
			{
				Label:       "Difficulty:",
				Font:        balance.UIFont,
				SelectValue: config.EditLevel.Difficulty,
				Tooltip: ui.Tooltip{
					Text: "Peaceful: enemies may not attack\n" +
						"Normal: default difficulty\n" +
						"Hard: enemies may be more aggressive",
					Edge: ui.Top,
				},
				Options: []magicform.Option{
					{
						Label: "Peaceful",
						Value: enum.Peaceful,
					},
					{
						Label: "Normal (recommended)",
						Value: enum.Normal,
					},
					{
						Label: "Hard",
						Value: enum.Hard,
					},
				},
				OnSelect: func(v interface{}) {
					value, _ := v.(enum.Difficulty)
					config.EditLevel.Difficulty = value
					log.Info("Set level difficulty to: %d (%s)", value, value)
				},
			},
			{
				Label: "Metadata",
				Font:  balance.LabelFont,
			},
			{
				Label:        "Title:",
				Font:         balance.UIFont,
				TextVariable: &config.EditLevel.Title,
				PromptUser: func(answer string) {
					config.EditLevel.Title = answer
				},
			},
			{
				Label:        "Author:",
				Font:         balance.UIFont,
				TextVariable: &config.EditLevel.Author,
				PromptUser: func(answer string) {
					config.EditLevel.Author = answer
				},
			},
		}...)
	}

	// The confirm/cancel buttons.
	var okLabel = "Ok"
	if config.EditLevel == nil {
		okLabel = "Continue"
	}
	fields = append(fields, []magicform.Field{
		{
			Buttons: []magicform.Field{
				{
					ButtonStyle: &balance.ButtonPrimary,
					Label:       okLabel,
					Font:        balance.UIFont,
					OnClick: func() {
						// Is it a NEW level?
						if config.EditLevel == nil {
							shmem.Flash("Create new map with %s page type and %s wallpaper", newPageType, newWallpaper)
							pageType, ok := level.PageTypeFromString(newPageType)
							if !ok {
								shmem.Flash("Invalid Page Type '%s'", newPageType)
								return
							}

							lvl := level.New()
							lvl.Palette = level.DefaultPalettes[paletteName]
							lvl.Wallpaper = newWallpaper
							lvl.PageType = pageType

							// Was a custom wallpaper selected for our NEW level?
							if lvl.Wallpaper == balance.CustomWallpaperFilename && len(newWallpaperB64) > 0 {
								lvl.SetFile(balance.CustomWallpaperEmbedPath, []byte(newWallpaperB64))
							}

							if config.OnCreateNewLevel != nil {
								config.OnCreateNewLevel(lvl)
							} else {
								shmem.FlashError("OnCreateNewLevel not attached")
							}
						} else {
							// Editing an existing level.

							// If we're editing a level, did we select a new palette?
							// Warn the user about if they want to change palettes.
							if paletteName != textCurrentPalette {
								modal.Confirm(
									"Are you sure you want to change the level palette?\n" +
										"Existing pixels drawn on your level may change, and\n" +
										"if the new palette is smaller, some pixels may be\n" +
										"lost from your level. OK to continue?",
								).WithTitle("Change Level Palette").Then(func() {
									// Install the new level palette.
									config.EditLevel.ReplacePalette(level.DefaultPalettes[paletteName])
									if config.OnReload != nil {
										config.OnReload()
									}
								})
								return
							}

							config.OnCancel()
						}
					},
				},
				{
					Label: "Cancel",
					Font:  balance.UIFont,
					OnClick: func() {
						config.OnCancel()
					},
				},
			},
		},
	}...)

	form.Create(frame, fields)
}

// Creates the "New Doodad" frame.
func (config AddEditLevel) setupDoodadFrame(tf *ui.TabFrame) {
	// Default options.
	var (
		doodadSize = 64
	)

	frame := tf.AddTab("doodad", ui.NewLabel(ui.Label{
		Text: "New Doodad",
		Font: balance.TabFont,
	}))

	/******************
	 * Frame for selecting Page Type
	 ******************/

	typeFrame := ui.NewFrame("Doodad Options Frame")
	frame.Pack(typeFrame, ui.Pack{
		Side:  ui.N,
		FillX: true,
	})

	label1 := ui.NewLabel(ui.Label{
		Text: "Doodad sprite size (square):",
		Font: balance.LabelFont,
	})
	typeFrame.Pack(label1, ui.Pack{
		Side: ui.W,
	})

	// A selectbox to suggest some sizes or let the user enter a custom.
	sizeBtn := ui.NewSelectBox("Size Select", ui.Label{
		Font: ui.MenuFont,
	})
	typeFrame.Pack(sizeBtn, ui.Pack{
		Side:   ui.W,
		Expand: true,
	})

	for _, row := range []struct {
		Name  string
		Value int
	}{
		{"32", 32},
		{"64", 64},
		{"96", 96},
		{"128", 128},
		{"200", 200},
		{"256", 256},
		{"Custom...", 0},
	} {
		row := row
		sizeBtn.AddItem(row.Name, row.Value, func() {})
	}

	sizeBtn.SetValue(doodadSize)
	sizeBtn.Handle(ui.Change, func(ed ui.EventData) error {
		if selection, ok := sizeBtn.GetValue(); ok {
			if size, ok := selection.Value.(int); ok {
				if size == 0 {
					shmem.Prompt("Enter a custom size for the doodad width and height: ", func(answer string) {
						if a, err := strconv.Atoi(answer); err == nil && a > 0 {
							doodadSize = a
						} else {
							shmem.FlashError("Doodad size should be a number greater than zero.")
						}
					})
				} else {
					doodadSize = size
				}
			}
		}
		return nil
	})

	sizeBtn.Supervise(config.Supervisor)
	config.Supervisor.Add(sizeBtn)

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
			if config.OnCreateNewDoodad != nil {
				config.OnCreateNewDoodad(doodadSize)
			} else {
				shmem.FlashError("OnCreateNewDoodad not attached")
			}
			return nil
		}},

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
