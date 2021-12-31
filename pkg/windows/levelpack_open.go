package windows

import (
	"fmt"
	"math"

	"git.kirsle.net/apps/doodle/pkg/balance"
	"git.kirsle.net/apps/doodle/pkg/levelpack"
	"git.kirsle.net/apps/doodle/pkg/log"
	"git.kirsle.net/go/render"
	"git.kirsle.net/go/ui"
)

// LevelPack window lets the user open and play a level from a pack.
type LevelPack struct {
	Supervisor *ui.Supervisor
	Engine     render.Engine

	// Callback functions.
	OnPlayLevel   func(pack levelpack.LevelPack, level levelpack.Level)
	OnCloseWindow func()

	// Internal variables
	window   *ui.Window
	tabFrame *ui.TabFrame
}

// NewLevelPackWindow initializes the window.
func NewLevelPackWindow(config LevelPack) *ui.Window {
	// Default options.
	var (
		title = "Select a Level"

		// size of the popup window
		width  = 320
		height = 300
	)

	// Get the available .levelpack files.
	lpFiles, packmap, err := levelpack.LoadAllAvailable()
	if err != nil {
		log.Error("Couldn't list levelpack files: %s", err)
	}

	window := ui.NewWindow(title)
	window.SetButtons(ui.CloseButton)
	window.Configure(ui.Config{
		Width:      width,
		Height:     height,
		Background: render.Grey,
	})
	config.window = window

	frame := ui.NewFrame("Window Body Frame")
	window.Pack(frame, ui.Pack{
		Side:   ui.N,
		Fill:   true,
		Expand: true,
	})

	// Use a TabFrame to organize the "screens" of this window.
	// The default screen is a pager for LevelPacks,
	// And each LevelPack's screen is a pager for its Levels.
	tabFrame := ui.NewTabFrame("Screens Manager")
	tabFrame.SetTabsHidden(true)
	tabFrame.Supervise(config.Supervisor)
	window.Pack(tabFrame, ui.Pack{
		Side:  ui.N,
		FillX: true,
	})
	config.tabFrame = tabFrame

	// Make the tabs.
	indexTab := tabFrame.AddTab("LevelPacks", ui.NewLabel(ui.Label{
		Text: "LevelPacks",
		Font: balance.TabFont,
	}))
	config.makeIndexScreen(indexTab, width, height, lpFiles, packmap, func(screen string) {
		// Callback for user choosing a level pack.
		// Hide the index screen and show the screen for this pack.
		log.Info("Called for tab: %s", screen)
		tabFrame.SetTab(screen)
	})
	for _, filename := range lpFiles {
		tab := tabFrame.AddTab(filename, ui.NewLabel(ui.Label{
			Text: filename,
			Font: balance.TabFont,
		}))
		config.makeDetailScreen(tab, width, height, packmap[filename])
	}

	// Close button.
	if config.OnCloseWindow != nil {
		closeBtn := ui.NewButton("Close Window", ui.NewLabel(ui.Label{
			Text: "Close",
			Font: balance.MenuFont,
		}))
		closeBtn.Handle(ui.Click, func(ed ui.EventData) error {
			config.OnCloseWindow()
			return nil
		})
		config.Supervisor.Add(closeBtn)
		window.Place(closeBtn, ui.Place{
			Bottom: 15,
			Center: true,
		})
	}

	window.Supervise(config.Supervisor)
	window.Hide()
	return window
}

/* Index screen for the LevelPack window.

frame: a TabFrame to populate
*/
func (config LevelPack) makeIndexScreen(frame *ui.Frame, width, height int,
	lpFiles []string, packmap map[string]levelpack.LevelPack, onChoose func(string)) {
	var (
		buttonHeight = 60 // height of each LevelPack button
		buttonWidth  = width - 40

		// pagination values
		page           = 1
		pages          int
		perPage        = 3
		maxPageButtons = 10
	)

	label := ui.NewLabel(ui.Label{
		Text: "Select from a Level Pack below:",
		Font: balance.LabelFont,
	})
	frame.Pack(label, ui.Pack{
		Side: ui.N,
		PadX: 8,
		PadY: 8,
	})

	pages = int(
		math.Ceil(
			float64(len(lpFiles)) / float64(perPage),
		),
	)

	var buttons []*ui.Button
	for i, filename := range lpFiles {
		filename := filename
		lp, ok := packmap[filename]
		if !ok {
			log.Error("Couldn't find %s in packmap!", filename)
			continue
		}

		// Make a frame to hold a complex button layout.
		btnFrame := ui.NewFrame("Frame")
		btnFrame.Resize(render.Rect{
			W: buttonWidth,
			H: buttonHeight,
		})

		// Draw labels...
		label := ui.NewLabel(ui.Label{
			Text: lp.Title,
			Font: balance.LabelFont,
		})
		btnFrame.Pack(label, ui.Pack{
			Side: ui.N,
		})

		description := lp.Description
		if description == "" {
			description = "(No description)"
		}

		byline := ui.NewLabel(ui.Label{
			Text: description,
			Font: balance.MenuFont,
		})
		btnFrame.Pack(byline, ui.Pack{
			Side: ui.N,
		})

		numLevels := ui.NewLabel(ui.Label{
			Text: fmt.Sprintf("[%d levels]", len(lp.Levels)),
			Font: balance.MenuFont,
		})
		btnFrame.Pack(numLevels, ui.Pack{
			Side: ui.N,
		})

		button := ui.NewButton(filename, btnFrame)
		button.Handle(ui.Click, func(ed ui.EventData) error {
			onChoose(filename)
			return nil
		})

		frame.Pack(button, ui.Pack{
			Side: ui.N,
			PadY: 2,
		})
		config.Supervisor.Add(button)

		if i > perPage-1 {
			button.Hide()
		}
		buttons = append(buttons, button)
	}

	pager := ui.NewPager(ui.Pager{
		Name:           "LevelPack Pager",
		Page:           page,
		Pages:          pages,
		PerPage:        perPage,
		MaxPageButtons: maxPageButtons,
		Font:           balance.MenuFont,
		OnChange: func(newPage, perPage int) {
			page = newPage
			log.Info("Page: %d, %d", page, perPage)

			// Re-evaluate which rows are shown/hidden for the page we're on.
			var (
				minRow  = (page - 1) * perPage
				visible = 0
			)
			for i, row := range buttons {
				if visible >= perPage {
					row.Hide()
					continue
				}

				if i < minRow {
					row.Hide()
				} else {
					row.Show()
					visible++
				}
			}
		},
	})
	pager.Compute(config.Engine)
	pager.Supervise(config.Supervisor)
	frame.Pack(pager, ui.Pack{
		Side: ui.N,
		PadY: 2,
	})
}

// Detail screen for a given levelpack.
func (config LevelPack) makeDetailScreen(frame *ui.Frame, width, height int, lp levelpack.LevelPack) *ui.Frame {
	var (
		buttonHeight = 40
		buttonWidth  = width - 40

		page    = 1
		perPage = 3
		pages   = int(
			math.Ceil(
				float64(len(lp.Levels)) / float64(perPage),
			),
		)
		maxPageButtons = 10
	)

	/** Back Button */
	backButton := ui.NewButton("Back", ui.NewLabel(ui.Label{
		Text: "< Back",
		Font: ui.MenuFont,
	}))
	backButton.SetStyle(&balance.ButtonBabyBlue)
	backButton.Handle(ui.Click, func(ed ui.EventData) error {
		config.tabFrame.SetTab("LevelPacks")
		return nil
	})
	config.Supervisor.Add(backButton)
	frame.Pack(backButton, ui.Pack{
		Side: ui.NE,
		PadY: 2,
		PadX: 6,
	})

	// Spacer: the back button is position NW and the rest against N
	// so may overlap.
	spacer := ui.NewFrame("Spacer")
	spacer.Configure(ui.Config{
		Width:  64,
		Height: 30,
	})
	frame.Pack(spacer, ui.Pack{
		Side: ui.N,
	})

	// LevelPack Title label
	label := ui.NewLabel(ui.Label{
		Text: lp.Title,
		Font: balance.LabelFont,
	})
	frame.Pack(label, ui.Pack{
		Side: ui.NW,
		PadX: 8,
		PadY: 2,
	})

	// Description
	if lp.Description != "" {
		label := ui.NewLabel(ui.Label{
			Text: lp.Description,
			Font: balance.MenuFont,
		})
		frame.Pack(label, ui.Pack{
			Side: ui.N,
			PadX: 8,
			PadY: 2,
		})
	}

	// Byline
	if lp.Author != "" {
		label := ui.NewLabel(ui.Label{
			Text: "by " + lp.Author,
			Font: balance.MenuFont,
		})
		frame.Pack(label, ui.Pack{
			Side: ui.N,
			PadX: 8,
			PadY: 2,
		})
	}

	// Loop over all the levels in this pack.
	var buttons []*ui.Button
	for i, level := range lp.Levels {
		level := level

		// Make a frame to hold a complex button layout.
		btnFrame := ui.NewFrame("Frame")
		btnFrame.Resize(render.Rect{
			W: buttonWidth,
			H: buttonHeight,
		})

		title := ui.NewLabel(ui.Label{
			Text: level.Title,
			Font: balance.LabelFont,
		})
		btnFrame.Pack(title, ui.Pack{
			Side: ui.NW,
		})

		btn := ui.NewButton(level.Filename, btnFrame)
		btn.Handle(ui.Click, func(ed ui.EventData) error {
			// Play Level
			if config.OnPlayLevel != nil {
				config.OnPlayLevel(lp, level)
			} else {
				log.Error("LevelPack Window: OnPlayLevel callback not ready")
			}
			return nil
		})

		frame.Pack(btn, ui.Pack{
			Side: ui.N,
			PadY: 2,
		})
		config.Supervisor.Add(btn)

		if i > perPage-1 {
			btn.Hide()
		}
		buttons = append(buttons, btn)
	}

	pager := ui.NewPager(ui.Pager{
		Name:           "Level Pager",
		Page:           page,
		Pages:          pages,
		PerPage:        perPage,
		MaxPageButtons: maxPageButtons,
		Font:           balance.MenuFont,
		OnChange: func(newPage, perPage int) {
			page = newPage
			log.Info("Page: %d, %d", page, perPage)

			// Re-evaluate which rows are shown/hidden for the page we're on.
			var (
				minRow  = (page - 1) * perPage
				visible = 0
			)
			for i, row := range buttons {
				if visible >= perPage {
					row.Hide()
					continue
				}

				if i < minRow {
					row.Hide()
				} else {
					row.Show()
					visible++
				}
			}
		},
	})
	pager.Compute(config.Engine)
	pager.Supervise(config.Supervisor)
	frame.Pack(pager, ui.Pack{
		Side: ui.N,
		PadY: 2,
	})

	return frame
}
