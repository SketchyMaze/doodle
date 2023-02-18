package windows

import (
	"fmt"
	"math"
	"strings"

	"git.kirsle.net/SketchyMaze/doodle/pkg/balance"
	"git.kirsle.net/SketchyMaze/doodle/pkg/doodads"
	"git.kirsle.net/SketchyMaze/doodle/pkg/level"
	"git.kirsle.net/SketchyMaze/doodle/pkg/log"
	"git.kirsle.net/SketchyMaze/doodle/pkg/uix"
	"git.kirsle.net/SketchyMaze/doodle/pkg/usercfg"
	"git.kirsle.net/go/render"
	"git.kirsle.net/go/ui"
)

// DoodadDropper is the doodad palette pop-up window for Editor Mode.
type DoodadDropper struct {
	Supervisor *ui.Supervisor
	Engine     render.Engine

	// Editing settings for an existing level?
	EditLevel *level.Level

	// Callback functions.
	OnStartDragActor func(doodad *doodads.Doodad, actor *level.Actor)
	OnCancel         func()
}

// NewDoodadDropper initializes the window.
func NewDoodadDropper(config DoodadDropper) *ui.Window {
	// Default options.
	var (
		title = "Doodads"

		buttonSize = balance.DoodadButtonSize
		columns    = balance.DoodadDropperCols
		rows       = balance.DoodadDropperRows

		// size of the doodad window
		width  = buttonSize * columns
		height = (buttonSize * rows) + 64 // account for button borders :(

		// Collect the Canvas widgets that all the doodads will be drawn in.
		// When the doodad window is closed or torn down, we can free up
		// the SDL2 textures and avoid a slow memory leak.
		canvases = []*uix.Canvas{}
	)

	// Get all the doodads.
	doodadsAvailable, err := doodads.ListDoodads()
	if err != nil {
		log.Error("NewDoodadDropper: doodads.ListDoodads: %s", err)
	}

	// Load all the doodads, skip hidden ones.
	var items []*doodads.Doodad
	for _, filename := range doodadsAvailable {
		if filename == "_autosave.doodad" {
			continue
		}

		doodad, err := doodads.LoadFile(filename)
		if err != nil {
			log.Error(err.Error())
			doodad = doodads.New(balance.DoodadSize)
		}

		// Skip hidden doodads.
		if doodad.Hidden && !usercfg.Current.ShowHiddenDoodads {
			continue
		}

		doodad.Filename = filename
		items = append(items, doodad)
	}

	window := ui.NewWindow(title)
	window.SetButtons(ui.CloseButton)
	window.Configure(ui.Config{
		Width:      width,
		Height:     height + 30,
		Background: render.Grey,
	})

	// When the window is closed, clear canvas textures. Note: we still cache
	// bitmap images in memory, those would be garbage collected by Go and SDL2
	// textures can always be regenerated.
	window.Handle(ui.CloseWindow, func(ed ui.EventData) error {
		log.Debug("Doodad Dropper: window closed, free %d canvas textures", len(canvases))
		for _, can := range canvases {
			can.Destroy()
		}
		return nil
	})

	tabFrame := ui.NewTabFrame("Category Tabs")
	window.Pack(tabFrame, ui.Pack{
		Side:   ui.N,
		Fill:   true,
		Expand: true,
	})

	// The Category Tabs.
	categories := []struct {
		ID   string
		Name string
	}{
		{"objects", "Objects"},
		{"doors", "Doors"},
		{"gizmos", "Gizmos"},
		{"creatures", "Creatures"},
		{"technical", "Technical"},
		{"", "All"},
	}
	for _, category := range categories {
		tab1 := tabFrame.AddTab(category.Name, ui.NewLabel(ui.Label{
			Text: category.Name,
			Font: balance.TabFont,
		}))
		cans := makeDoodadTab(config, tab1, render.NewRect(width-4, height-60), category.ID, items)
		canvases = append(canvases, cans...)
	}

	tabFrame.Supervise(config.Supervisor)

	window.Hide()
	return window
}

// Function to generate the TabFrame frame of the Doodads window.
func makeDoodadTab(config DoodadDropper, frame *ui.Frame, size render.Rect, category string, available []*doodads.Doodad) []*uix.Canvas {
	var (
		buttonSize = balance.DoodadButtonSize
		columns    = balance.DoodadDropperCols
		rows       = balance.DoodadDropperRows

		// Count how many doodad buttons we need vs. how many can fit.
		iconsDrawn    int
		iconsPossible = columns * rows

		// pagination values
		page           = 1
		pages          int
		perPage        = 20
		maxPageButtons = 10

		// Collect the created Canvas widgets so we can free SDL2 textures later.
		canvases = []*uix.Canvas{}
	)
	frame.Resize(size)

	// Trim the available doodads to those fitting the category.
	var items = []*doodads.Doodad{}
	for _, candidate := range available {
		if value, ok := candidate.Tags["category"]; ok {
			if category != "" && !strings.Contains(value, category) {
				continue
			}
		} else if category != "" {
			continue
		}
		items = append(items, candidate)
	}

	doodads.SortByName(items)

	// Compute the number of pages for the pager widget.
	pages = int(
		math.Ceil(
			float64(len(items)) / float64(columns*rows),
		),
	)

	// First, draw the empty grid of inset frames to serve as the 'background'
	// of the drawer. This both serves an aesthetic purpose and reserves space
	// in the widget for short page views.
	{
		var (
			decorFrame = ui.NewFrame("Background Slots")
			row        *ui.Frame
		)

		for i := 0; i < iconsPossible; i++ {
			if row == nil || i%columns == 0 {
				row = ui.NewFrame("BG Row")
				decorFrame.Pack(row, ui.Pack{
					Side: ui.N,
				})
			}

			spacer := ui.NewFrame("Spacer")
			spacer.Configure(ui.Config{
				BorderSize:  2,
				BorderStyle: ui.BorderSunken,
				Background:  render.Grey.Darken(20),
			})
			spacer.Resize(render.NewRect(
				buttonSize-2, // TODO: without the -2 the button border
				buttonSize-2, // rests on top of the window border
			))
			spacer.Compute(config.Engine)
			row.Pack(spacer, ui.Pack{
				Side: ui.W,
			})
		}

		decorFrame.Compute(config.Engine)

		// frame.Pack(decorFrame, ui.Pack{
		// 	Side: ui.NW,
		// })
		frame.Place(decorFrame, ui.Place{
			Top:  0,
			Left: 0,
		})
	}

	// Draw the doodad buttons in rows.
	var btnRows = []*ui.Frame{}
	{
		var (
			row      *ui.Frame
			rowCount int // for labeling the ui.Frame for each row

			// the state we end up at when we exhaust all doodads
			lastColumn int // last position in current row
		)

		for i, doodad := range items {
			doodad := doodad

			if row == nil || i%columns == 0 {
				var hidden = rowCount >= rows
				rowCount++

				row = ui.NewFrame(fmt.Sprintf("Doodad Row %d", rowCount))

				row.Resize(render.NewRect(size.W, buttonSize))
				row.Compute(config.Engine)
				btnRows = append(btnRows, row)
				frame.Pack(row, ui.Pack{
					Side: ui.N,
				})

				// Hide overflowing rows until we page to them.
				if hidden {
					row.Hide()
				}

				// New row, new columns.
				lastColumn = 0
			}

			can := uix.NewCanvas(uint8(buttonSize), true) // TODO: dangerous - buttonSize must be small
			can.Name = doodad.Title
			can.SetBackground(balance.DoodadButtonBackground)
			can.LoadDoodad(doodad)

			// Keep the canvas to free textures later.
			canvases = append(canvases, can)

			btn := ui.NewButton(doodad.Title, can)
			can.CroppedSize = true
			btn.Resize(render.NewRect(
				buttonSize-2, // TODO: without the -2 the button border
				buttonSize-2, // rests on top of the window border
			))
			row.Pack(btn, ui.Pack{
				Side: ui.W,
			})

			// Tooltip hover to show the doodad's name.
			tt := ui.NewTooltip(btn, ui.Tooltip{
				Text: doodad.Title,
				Edge: ui.Top,
			})
			tt.Supervise(config.Supervisor)

			// Begin the drag event to grab this Doodad.
			// NOTE: The drag target is the EditorUI.Canvas in
			// editor_ui.go#SetupCanvas()
			btn.Handle(ui.MouseDown, func(ed ui.EventData) error {
				log.Warn("MouseDown on doodad %s (%s)", doodad.Filename, doodad.Title)
				config.OnStartDragActor(doodad, nil)
				return nil
			})
			config.Supervisor.Add(btn)

			// Resize the canvas to fill the button interior.
			btnSize := btn.Size()
			can.Resize(render.NewRect(
				btnSize.W-btn.BoxThickness(2),
				btnSize.H-btn.BoxThickness(2),
			))

			btn.Compute(config.Engine)

			iconsDrawn++
			lastColumn++
		}

		// If we have fewer doodad icons than this page can hold,
		// fill out dummy placeholder cells to maintain the UI shape.
		// TODO: this is very redundant compared to the ATTEMPT above
		// to only do this once. It seems our background widget doesn't
		// size up the full tab height properly, so doodad tabs that
		// have fewer than one page worth (short first page) the sizing
		// was wrong. The below hack pads out the screen for short first
		// pages only. There is still a bug with short LAST pages where
		// it doesn't hold height and the pager buttons come up.
		if iconsDrawn < iconsPossible {
			for i := lastColumn; i < iconsPossible; i++ {
				if row == nil || i%columns == 0 {
					var hidden = rowCount >= rows
					rowCount++

					row = ui.NewFrame(fmt.Sprintf("Doodad Row %d", rowCount))
					row.SetBackground(balance.DoodadButtonBackground)
					btnRows = append(btnRows, row)
					frame.Pack(row, ui.Pack{
						Side: ui.N,
					})

					// Hide overflowing rows until we page to them.
					if hidden {
						row.Hide()
					}
				}

				spacer := ui.NewFrame("Spacer")
				spacer.Configure(ui.Config{
					BorderSize:  2,
					BorderStyle: ui.BorderSunken,
					Background:  render.Grey,
				})
				spacer.Resize(render.NewRect(
					buttonSize-2, // TODO: without the -2 the button border
					buttonSize-2, // rests on top of the window border
				))
				spacer.Compute(config.Engine)
				row.Pack(spacer, ui.Pack{
					Side: ui.W,
				})

				// debug
				// lbl := ui.NewLabel(ui.Label{
				// 	Text: fmt.Sprintf("i=%d\nrow=%d", i, rowCount),
				// })
				// spacer.Pack(lbl, ui.Pack{
				// 	Side: ui.NW,
				// })
			}
		}
	}

	{
		/******************
		 * Confirm/cancel buttons.
		 ******************/

		bottomFrame := ui.NewFrame("Button Frame")
		frame.Pack(bottomFrame, ui.Pack{
			Side:  ui.N,
			FillX: true,
		})

		// Pager for the doodads.
		pager := ui.NewPager(ui.Pager{
			Name:           "Doodad Dropper Pager",
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
					minRow  = (page - 1) * rows
					visible = 0
				)
				for i, row := range btnRows {
					if visible >= rows {
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
		bottomFrame.Place(pager, ui.Place{
			Top:  20,
			Left: 20,
		})

		var buttons = []struct {
			Label string
			F     func(ui.EventData) error
		}{
			// OK button is for editing an existing level.
			{"Close", func(ed ui.EventData) error {
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
			bottomFrame.Place(btn, ui.Place{
				Top:   20,
				Right: 20,
			})
		}
	}

	return canvases
}
