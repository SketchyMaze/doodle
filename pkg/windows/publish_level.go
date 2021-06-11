package windows

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"git.kirsle.net/apps/doodle/pkg/balance"
	"git.kirsle.net/apps/doodle/pkg/doodads"
	"git.kirsle.net/apps/doodle/pkg/level"
	"git.kirsle.net/apps/doodle/pkg/log"
	"git.kirsle.net/go/render"
	"git.kirsle.net/go/ui"
)

// Publish window.
type Publish struct {
	// Settings passed in by doodle
	Supervisor *ui.Supervisor
	Engine     render.Engine
	Level      *level.Level

	OnPublish func()
	OnCancel  func()

	// Private vars.
	includeBuiltins bool // show built-in doodads in checkbox-list.
}

// NewPublishWindow initializes the window.
func NewPublishWindow(cfg Publish) *ui.Window {
	var (
		windowWidth    = 400
		windowHeight   = 300
		page           = 1
		perPage        = 4
		pages          = 1
		maxPageButtons = 8

		// columns and sizes to draw the doodad list
		columns   = 3
		btnWidth  = 120
		btnHeight = 14
	)

	window := ui.NewWindow("Publish Level")
	window.SetButtons(ui.CloseButton)
	window.Configure(ui.Config{
		Width:      windowWidth,
		Height:     windowHeight,
		Background: render.RGBA(200, 200, 255, 255),
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
			Text: "Share your level easily! If you are using custom doodads in\n" +
				"your level, you may attach them directly to your\n" +
				"level file -- so it can easily run on another computer!",
			Font: balance.UIFont,
		},
		{
			Text: "List of Doodads in Your Level",
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
	// Custom Doodads checkbox-list.
	doodadFrame := ui.NewFrame("Doodads Frame")
	doodadFrame.Resize(render.Rect{
		W: windowWidth,
		H: btnHeight*perPage + 100,
	})
	window.Pack(doodadFrame, ui.Pack{
		Side:  ui.N,
		FillX: true,
	})

	// First, the checkbox to show built-in doodads or not.
	builtinRow := ui.NewFrame("Show Builtins Frame")
	doodadFrame.Pack(builtinRow, ui.Pack{
		Side:  ui.N,
		FillX: true,
	})
	builtinCB := ui.NewCheckbox("Show Builtins", &cfg.includeBuiltins, ui.NewLabel(ui.Label{
		Text: "Attach built-in* doodads too",
		Font: balance.UIFont,
	}))
	builtinCB.Supervise(cfg.Supervisor)
	builtinRow.Pack(builtinCB, ui.Pack{
		Side: ui.W,
		PadX: 2,
	})

	// Collect all the doodad names in use in this level.
	unique := map[string]interface{}{}
	names := []string{}
	if cfg.Level != nil {
		for _, actor := range cfg.Level.Actors {
			if _, ok := unique[actor.Filename]; ok {
				continue
			}
			unique[actor.Filename] = nil
			names = append(names, actor.Filename)
		}
	}

	sort.Strings(names)

	// Identify which of the doodads are built-ins.
	usedBuiltins := []string{}
	builtinMap := map[string]interface{}{}
	usedCustom := []string{}
	if builtins, err := doodads.ListBuiltin(); err == nil {
		for _, filename := range builtins {
			if _, ok := unique[filename]; ok {
				usedBuiltins = append(usedBuiltins, filename)
				builtinMap[filename] = nil
			}
		}
	}
	for _, name := range names {
		if _, ok := builtinMap[name]; ok {
			continue
		}
		usedCustom = append(usedCustom, name)
	}

	// Helper function to draw the button rows for a set of doodads.
	mkDoodadRows := func(filenames []string, builtin bool) []*ui.Frame {
		var (
			curRow *ui.Frame // = ui.NewFrame("mkDoodadRows 0")
			frames = []*ui.Frame{}
		)

		for i, name := range filenames {
			if i%columns == 0 {
				curRow = ui.NewFrame(fmt.Sprintf("mkDoodadRows %d", i))
				frames = append(frames, curRow)
			}

			font := balance.UIFont
			if builtin {
				font.Color = render.Blue
				name += "*"
			}

			btn := ui.NewLabel(ui.Label{
				Text: strings.Replace(name, ".doodad", "", 1),
				Font: font,
			})
			btn.Configure(ui.Config{
				Width:  btnWidth,
				Height: btnHeight,
			})
			curRow.Pack(btn, ui.Pack{
				Side: ui.W,
				PadX: 2,
				PadY: 2,
			})
		}

		return frames
	}

	// 1. Draw the built-in doodads in use.
	var (
		btnRows     = []*ui.Frame{}
		builtinRows = []*ui.Frame{}
		customRows  = []*ui.Frame{}
	)
	if len(names) > 0 {
		customRows = mkDoodadRows(usedCustom, false)
		btnRows = append(btnRows, customRows...)
	}
	if len(usedBuiltins) > 0 {
		builtinRows = mkDoodadRows(usedBuiltins, true)
		btnRows = append(btnRows, builtinRows...)
	}

	for i, row := range btnRows {
		doodadFrame.Pack(row, ui.Pack{
			Side:  ui.N,
			FillX: true,
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
			float64(len(btnRows)) / float64(perPage),
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
		for i, row := range btnRows {
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
		Name:           "Doodads List Pager",
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
		{"Export Level", true, func() {
			if cfg.OnPublish != nil {
				cfg.OnPublish()
			}
		}},
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
