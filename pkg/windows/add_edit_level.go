package windows

import (
	"fmt"
	"regexp"
	"strconv"

	"git.kirsle.net/SketchyMaze/doodle/pkg/balance"
	"git.kirsle.net/SketchyMaze/doodle/pkg/enum"
	"git.kirsle.net/SketchyMaze/doodle/pkg/level"
	"git.kirsle.net/SketchyMaze/doodle/pkg/log"
	"git.kirsle.net/SketchyMaze/doodle/pkg/modal"
	"git.kirsle.net/SketchyMaze/doodle/pkg/native"
	"git.kirsle.net/SketchyMaze/doodle/pkg/shmem"
	magicform "git.kirsle.net/SketchyMaze/doodle/pkg/uix/magic-form"
	"git.kirsle.net/SketchyMaze/doodle/pkg/wallpaper"
	"git.kirsle.net/go/render"
	"git.kirsle.net/go/ui"
)

// AddEditLevel is the "Create New Level & Edit Level Properties" window
type AddEditLevel struct {
	Supervisor *ui.Supervisor
	Engine     render.Engine

	// Editing settings for an existing level?
	EditLevel *level.Level

	// Show the "New Doodad" tab by default?
	NewDoodad bool

	// Callback functions.
	OnChangePageTypeAndWallpaper func(pageType level.PageType, wallpaper string)
	OnCreateNewLevel             func(*level.Level)
	OnCreateNewDoodad            func(width, height int)
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
		Height:     290,
		Background: render.Grey,
	})

	// Tabbed UI for New Level or New Doodad.
	tabframe := ui.NewTabFrame("Level Tabs")
	window.Pack(tabframe, ui.Pack{
		Side:   ui.N,
		Fill:   true,
		Expand: true,
	})

	// Add the tabs.
	config.setupLevelFrame(tabframe) // Level Properties (always)
	if config.EditLevel == nil {
		// New Doodad properties (New window only)
		config.setupDoodadFrame(tabframe)
	} else {
		// Additional Level tabs (existing level only)
		config.setupGameRuleFrame(tabframe)
	}

	tabframe.Supervise(config.Supervisor)

	// Show the doodad tab?
	if config.NewDoodad {
		tabframe.SetTab("doodad")
	}

	window.Hide()
	return window
}

// Creates the Create/Edit Level tab ("index").
func (config AddEditLevel) setupLevelFrame(tf *ui.TabFrame) {
	// Default options.
	var (
		tabLabel     = "New Level"
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
		tabLabel = "Properties"
		newPageType = config.EditLevel.PageType.String()
		newWallpaper = config.EditLevel.Wallpaper
		paletteName = textCurrentPalette
	}

	frame := tf.AddTab("index", ui.NewLabel(ui.Label{
		Text: tabLabel,
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
			Options:     balance.Wallpapers,
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
		var (
			levelSizeStr    = fmt.Sprintf("%dx%d", config.EditLevel.MaxWidth, config.EditLevel.MaxHeight)
			levelSizeRegexp = regexp.MustCompile(`^(\d+)x(\d+)$`)
		)
		fields = append(fields, []magicform.Field{
			{
				Label:        "Limits (bounded):",
				Font:         balance.UIFont,
				TextVariable: &levelSizeStr,
				OnClick: func() {
					shmem.Prompt(fmt.Sprintf("Enter new limits in WxH format or [%s]: ", levelSizeStr), func(answer string) {
						if answer == "" {
							return
						}

						match := levelSizeRegexp.FindStringSubmatch(answer)
						if match == nil {
							return
						}

						levelSizeStr = match[0]
						width, _ := strconv.Atoi(match[1])
						height, _ := strconv.Atoi(match[2])

						config.EditLevel.MaxWidth = int64(width)
						config.EditLevel.MaxHeight = int64(height)
					})
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
	var okLabel = "Apply"
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
		doodadWidth  = 64
		doodadHeight = doodadWidth
	)

	frame := tf.AddTab("doodad", ui.NewLabel(ui.Label{
		Text: "New Doodad",
		Font: balance.TabFont,
	}))

	/******************
	 * Frame for selecting Page Type
	 ******************/

	var sizeOptions = []magicform.Option{
		{Label: "32", Value: 32},
		{Label: "64", Value: 64},
		{Label: "96", Value: 96},
		{Label: "128", Value: 128},
		{Label: "200", Value: 200},
		{Label: "256", Value: 256},
		{Label: "Custom...", Value: 0},
	}

	form := magicform.Form{
		Supervisor: config.Supervisor,
		Engine:     config.Engine,
		Vertical:   true,
		LabelWidth: 90,
	}
	form.Create(frame, []magicform.Field{
		{
			Label:       "Width:",
			Font:        balance.LabelFont,
			Type:        magicform.Selectbox,
			IntVariable: &doodadWidth,
			Options:     sizeOptions,
			OnSelect: func(v interface{}) {
				if v.(int) == 0 {
					shmem.Prompt("Enter a custom size for the doodad width: ", func(answer string) {
						if a, err := strconv.Atoi(answer); err == nil && a > 0 {
							doodadWidth = a
						} else {
							shmem.FlashError("Doodad size should be a number greater than zero.")
						}
					})
				}
			},
		},
		{
			Label:       "Height:",
			Font:        balance.LabelFont,
			Type:        magicform.Selectbox,
			IntVariable: &doodadHeight,
			Options:     sizeOptions,
			OnSelect: func(v interface{}) {
				if v.(int) == 0 {
					shmem.Prompt("Enter a custom size for the doodad height: ", func(answer string) {
						if a, err := strconv.Atoi(answer); err == nil && a > 0 {
							doodadHeight = a
						} else {
							shmem.FlashError("Doodad size should be a number greater than zero.")
						}
					})
				}
			},
		},
		{
			Buttons: []magicform.Field{
				{
					Label:       "Continue",
					Font:        balance.UIFont,
					ButtonStyle: &balance.ButtonPrimary,
					OnClick: func() {
						if config.OnCreateNewDoodad != nil {
							config.OnCreateNewDoodad(doodadWidth, doodadHeight)
						} else {
							shmem.FlashError("OnCreateNewDoodad not attached")
						}
					},
				},
				{
					Label:       "Cancel",
					Font:        balance.UIFont,
					ButtonStyle: &balance.ButtonPrimary,
					OnClick: func() {
						if config.OnCancel != nil {
							config.OnCancel()
						} else {
							shmem.FlashError("OnCancel not attached")
						}
					},
				},
			},
		},
	})

}

// Creates the Game Rules frame for existing level (set difficulty, etc.)
func (config AddEditLevel) setupGameRuleFrame(tf *ui.TabFrame) {
	frame := tf.AddTab("GameRules", ui.NewLabel(ui.Label{
		Text: "Game Rules",
		Font: balance.TabFont,
	}))

	form := magicform.Form{
		Supervisor: config.Supervisor,
		Engine:     config.Engine,
		Vertical:   true,
		LabelWidth: 120,
		PadY:       2,
	}
	fields := []magicform.Field{
		{
			Label: "Game Rules are specific to this level and can change some of\n" +
				"the game's default behaviors.",
			Font: balance.UIFont,
		},
		{
			Label:       "Difficulty:",
			Font:        balance.UIFont,
			SelectValue: config.EditLevel.GameRule.Difficulty,
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
				config.EditLevel.GameRule.Difficulty = value
				log.Info("Set level difficulty to: %d (%s)", value, value)
			},
		},
		{
			Label:        "Survival Mode (silver high score)",
			Font:         balance.UIFont,
			BoolVariable: &config.EditLevel.GameRule.Survival,
			Tooltip: ui.Tooltip{
				Text: "Use for levels where dying at least once is very likely\n" +
					"(e.g. Azulian Tag). The silver high score will be for\n" +
					"longest time rather than fastest time. The gold high\n" +
					"score will still be for fastest time.",
				Edge: ui.Top,
			},
		},
	}

	form.Create(frame, fields)
}
