package windows

import (
	"fmt"
	"math"

	"git.kirsle.net/SketchyMaze/doodle/pkg/balance"
	"git.kirsle.net/SketchyMaze/doodle/pkg/level"
	"git.kirsle.net/SketchyMaze/doodle/pkg/log"
	"git.kirsle.net/go/render"
	"git.kirsle.net/go/ui"
)

// FileSystem window shows the file attachments to the current level.
type FileSystem struct {
	// Settings passed in by doodle
	Supervisor *ui.Supervisor
	Engine     render.Engine
	Level      *level.Level

	OnDelete func(filename string) bool
	OnCancel func()

	// Private vars.
	includeBuiltins bool // show built-in doodads in checkbox-list.
}

// NewFileSystemWindow initializes the window.
func NewFileSystemWindow(cfg FileSystem) *ui.Window {
	var (
		windowColor      = render.RGBA(255, 255, 200, 255)
		windowTitleColor = render.RGBA(255, 153, 0, 255)
		windowWidth      = 380
		windowHeight     = 360
		page             = 1
		perPage          = 6
		pages            = 1
		maxPageButtons   = 8

		// columns and sizes to draw the doodad list
		btnHeight = 14
	)

	window := ui.NewWindow("Attached Files")
	window.SetButtons(ui.CloseButton)
	window.ActiveTitleBackground = windowTitleColor // TODO: not working?
	window.InactiveTitleBackground = windowTitleColor.Darken(60)
	window.InactiveTitleForeground = render.Grey
	window.Configure(ui.Config{
		Width:      windowWidth,
		Height:     windowHeight,
		Background: windowColor,
	})

	/////////////
	// Intro text

	introFrame := ui.NewFrame("Intro Frame")
	window.Pack(introFrame, ui.Pack{
		Side:  ui.N,
		FillX: true,
	})

	lines := []struct {
		Text string
		Font render.Text
	}{
		{
			Text: "About",
			Font: balance.LabelFont,
		},
		{
			Text: "These are the files embedded inside your level data. When\n" +
				"a level is Published, it can attach all of its custom doodads,\n" +
				"wallpapers or other custom asset so that it easily plays on\n" +
				"a different computer.",
			Font: balance.UIFont,
		},
		{
			Text: "File Attachments",
			Font: balance.LabelFont,
		},
	}
	for n, row := range lines {
		frame := ui.NewFrame(fmt.Sprintf("Intro Line %d", n))
		introFrame.Pack(frame, ui.Pack{
			Side:  ui.N,
			FillX: true,
		})

		label := ui.NewLabel(ui.Label{
			Text: row.Text,
			Font: row.Font,
		})
		frame.Pack(label, ui.Pack{
			Side: ui.W,
		})
	}

	/////////////
	// Attached files table.
	fsFrame := ui.NewFrame("Doodads Frame")
	fsFrame.Resize(render.Rect{
		W: windowWidth,
		H: btnHeight*perPage + 140,
	})
	window.Pack(fsFrame, ui.Pack{
		Side:  ui.N,
		FillX: true,
	})

	var fileRows = []*ui.Frame{}

	// Get the file attachments.
	files := cfg.Level.ListFiles()
	for _, file := range files {
		file := file
		row := ui.NewFrame("Row: " + file)
		label := ui.NewLabel(ui.Label{
			Text: file,
			Font: balance.UIFont,
		})
		row.Pack(label, ui.Pack{
			Side:    ui.W,
			Padding: 1,
		})

		delBtn := ui.NewButton("Delete: "+file, ui.NewLabel(ui.Label{
			Text: "Delete",
			Font: balance.SmallFont,
		}))
		delBtn.SetStyle(&balance.ButtonDanger)
		delBtn.Handle(ui.Click, func(ed ui.EventData) error {
			if cfg.OnDelete != nil {
				if cfg.OnDelete(file) {
					row.Hide()
				}
			}
			return nil
		})
		cfg.Supervisor.Add(delBtn)
		row.Place(delBtn, ui.Place{
			Right: 4,
		})

		fileRows = append(fileRows, row)
	}

	for i, row := range fileRows {
		fsFrame.Pack(row, ui.Pack{
			Side:  ui.N,
			FillX: true,
			PadY:  2,
		})

		// Hide if too long for 1st page.
		if i >= perPage {
			row.Hide()
		}
	}

	/////////////
	// Buttons at bottom of window

	bottomFrame := ui.NewFrame("Button Frame")
	window.Pack(bottomFrame, ui.Pack{
		Side:  ui.S,
		FillX: true,
	})

	// Pager for the doodads.
	pages = int(
		math.Ceil(
			float64(len(fileRows)) / float64(perPage),
		),
	)
	pagerOnChange := func(newPage, perPage int) {
		page = newPage
		log.Info("Page: %d, %d", page, perPage)

		// Re-evaluate which rows are shown/hidden for the page we're on.
		var (
			minRow  = (page - 1) * perPage
			visible = 0
		)
		for i, row := range fileRows {
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
	}
	pager := ui.NewPager(ui.Pager{
		Name:           "Files List Pager",
		Page:           page,
		Pages:          pages,
		PerPage:        perPage,
		MaxPageButtons: maxPageButtons,
		Font:           balance.MenuFont,
		OnChange:       pagerOnChange,
	})
	pager.Compute(cfg.Engine)
	pager.Supervise(cfg.Supervisor)
	bottomFrame.Place(pager, ui.Place{
		Top:  20,
		Left: 20,
	})

	frame := ui.NewFrame("Button frame")
	buttons := []struct {
		label   string
		primary bool
		f       func()
	}{
		{"Close", false, func() {
			if cfg.OnCancel != nil {
				cfg.OnCancel()
			}
		}},
	}
	for _, button := range buttons {
		button := button

		btn := ui.NewButton(button.label, ui.NewLabel(ui.Label{
			Text: button.label,
			Font: balance.MenuFont,
		}))
		if button.primary {
			btn.SetStyle(&balance.ButtonPrimary)
		}

		btn.Handle(ui.Click, func(ed ui.EventData) error {
			button.f()
			return nil
		})

		btn.Compute(cfg.Engine)
		cfg.Supervisor.Add(btn)

		frame.Pack(btn, ui.Pack{
			Side:   ui.W,
			PadX:   4,
			Expand: true,
			Fill:   true,
		})
	}
	bottomFrame.Pack(frame, ui.Pack{
		Side:    ui.E,
		Padding: 8,
	})

	return window
}
