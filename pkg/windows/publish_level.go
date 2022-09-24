package windows

import (
	"fmt"
	"math"
	"strings"

	"git.kirsle.net/SketchyMaze/doodle/pkg/balance"
	"git.kirsle.net/SketchyMaze/doodle/pkg/level"
	"git.kirsle.net/SketchyMaze/doodle/pkg/level/publishing"
	"git.kirsle.net/SketchyMaze/doodle/pkg/log"
	magicform "git.kirsle.net/SketchyMaze/doodle/pkg/uix/magic-form"
	"git.kirsle.net/go/render"
	"git.kirsle.net/go/ui"
)

// Publish window.
type Publish struct {
	// Settings passed in by doodle
	Supervisor *ui.Supervisor
	Engine     render.Engine
	Level      *level.Level

	OnPublish func(builtinToo bool)
	OnCancel  func()

	// Private vars.
	includeBuiltins bool // show built-in doodads in checkbox-list.
}

// NewPublishWindow initializes the window.
func NewPublishWindow(cfg Publish) *ui.Window {
	var (
		windowWidth    = 380
		windowHeight   = 220
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
	// Custom Doodads checkbox-list.
	doodadFrame := ui.NewFrame("Doodads Frame")
	doodadFrame.Resize(render.Rect{
		W: windowWidth,
		H: btnHeight*perPage + 40,
	})

	// Collect the doodads named in this level.
	usedBuiltins, usedCustom := publishing.GetUsedDoodadNames(cfg.Level)

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
	if len(usedCustom) > 0 {
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
	_ = pager

	/////////////
	// Intro text

	introFrame := ui.NewFrame("Intro Frame")
	window.Pack(introFrame, ui.Pack{
		Side:  ui.N,
		FillX: true,
	})

	// Render the form, putting it all together.
	form := magicform.Form{
		Supervisor: cfg.Supervisor,
		Engine:     cfg.Engine,
		Vertical:   true,
		LabelWidth: 100,
	}
	form.Create(introFrame, []magicform.Field{
		{
			Label: "About",
			Font:  balance.LabelFont,
		},
		{
			Label: "Share your level easily! If you are using custom doodads in\n" +
				"your level, you may attach them directly to your level file\n" +
				"so it can easily run on another computer!",
			Font: balance.UIFont,
		},
		{
			Label:        "Attach custom doodads when I save the level",
			Font:         balance.UIFont,
			BoolVariable: &cfg.Level.SaveDoodads,
		},
		{
			Label: "Attach built-in doodads too",
			Font: balance.UIFont.Update(render.Text{
				Color: render.Red,
			}),
			BoolVariable: &cfg.Level.SaveBuiltins,
			Tooltip: ui.Tooltip{
				Edge: ui.Top,
				Text: "If enabled, the attached doodads will override the built-ins\n" +
					"for this level. Bugfixes or updates to the built-ins will not\n" +
					"affect your level, either.",
			},
		},
		{
			Label: "The above settings are saved with your level file, and each\n" +
				"time you save, custom doodads will be re-attached.",
			Font: balance.UIFont,
		},
		// Pager is broken, Supervisor doesn't pick it up, TODO
		/*{
			Label: "Doodads currently used on this level:",
			Font:  balance.LabelFont,
		},
		{
			Frame: doodadFrame,
		},
		{
			Label: "* Built-in doodad",
			Font:  balance.UIFont,
		},
		{
			Pager: pager,
		},*/
		{
			Buttons: []magicform.Field{
				{
					ButtonStyle: &balance.ButtonPrimary,
					Label:       "Save Level Now",
					OnClick: func() {
						if cfg.OnPublish != nil {
							cfg.OnPublish(cfg.includeBuiltins)
						}
					},
				},
				{
					Type:  magicform.Button,
					Label: "Close",
					OnClick: func() {
						if cfg.OnCancel != nil {
							cfg.OnCancel()
						}
					},
				},
			},
		},
	})

	return window
}
