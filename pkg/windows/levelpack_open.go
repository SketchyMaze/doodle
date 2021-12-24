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
	OnPlayLevel func(levelpack, filename string)

	// Internal variables
	window    *ui.Window
	gotoIndex func() // return to index screen
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

	// We'll divide this window into "Screens", where the default
	// screen shows the available level packs and then each level
	// pack gets its own screen showing its levels.
	var indexScreen *ui.Frame
	config.gotoIndex = func() {
		indexScreen.Show()
	}
	indexScreen = config.makeIndexScreen(width, height, func(screen *ui.Frame) {
		// Callback for user choosing a level pack.
		// Hide the index screen and show the screen for this pack.
		indexScreen.Hide()
		screen.Show()
	})
	window.Pack(indexScreen, ui.Pack{
		Side:   ui.N,
		Fill:   true,
		Expand: true,
	})

	window.Supervise(config.Supervisor)
	window.Hide()
	return window
}

// Index screen for the LevelPack window.
func (config LevelPack) makeIndexScreen(width, height int, onChoose func(*ui.Frame)) *ui.Frame {
	var (
		buttonHeight = 60 // height of each LevelPack button
		buttonWidth  = width - 40

		// pagination values
		page           = 1
		pages          int
		perPage        = 3
		maxPageButtons = 10
	)
	frame := ui.NewFrame("Index Screen")

	label := ui.NewLabel(ui.Label{
		Text: "Select from a Level Pack below:",
		Font: balance.LabelFont,
	})
	frame.Pack(label, ui.Pack{
		Side: ui.N,
		PadX: 8,
		PadY: 8,
	})

	// Get the available .levelpack files.
	lpFiles, err := levelpack.ListFiles()
	if err != nil {
		log.Error("Couldn't list levelpack files: %s", err)
	}

	pages = int(
		math.Ceil(
			float64(len(lpFiles)) / float64(perPage),
		),
	)

	var buttons []*ui.Button
	for i, filename := range lpFiles {
		lp, err := levelpack.LoadFile(filename)
		if err != nil {
			log.Error("Couldn't read %s: %s", filename, err)
			continue
		}
		_ = lp

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

		// Generate the detail screen (Frame) for this level pack.
		// Should the user click our button, this screen is shown.
		screen := config.makeDetailScreen(width, height, lp)
		screen.Hide()
		config.window.Pack(screen, ui.Pack{
			Side:   ui.N,
			Fill:   true,
			Expand: true,
		})

		button := ui.NewButton(filename, btnFrame)
		button.Handle(ui.Click, func(ed ui.EventData) error {
			onChoose(screen)
			return nil
		})

		frame.Pack(button, ui.Pack{
			Side: ui.N,
			PadY: 2,
		})
		config.Supervisor.Add(button)

		if i > perPage {
			button.Hide()
		}
		buttons = append(buttons, button)
	}

	pager := ui.NewPager(ui.Pager{
		Name:           "LevelPack Pager",
		Page:           page,
		Pages:          pages,
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

// Detail screen for a given levelpack.
func (config LevelPack) makeDetailScreen(width, height int, lp levelpack.LevelPack) *ui.Frame {
	frame := ui.NewFrame("Detail Screen")

	label := ui.NewLabel(ui.Label{
		Text: "HELLO " + lp.Title,
		Font: balance.LabelFont,
	})
	frame.Pack(label, ui.Pack{
		Side: ui.N,
		PadX: 8,
		PadY: 8,
	})

	backButton := ui.NewButton("Back", ui.NewLabel(ui.Label{
		Text: "< Back to Level Packs",
		Font: ui.MenuFont,
	}))
	backButton.Handle(ui.Click, func(ed ui.EventData) error {
		frame.Hide()
		config.gotoIndex()
		return nil
	})
	config.Supervisor.Add(backButton)
	frame.Pack(backButton, ui.Pack{
		Side: ui.N,
		PadY: 2,
	})

	return frame
}
