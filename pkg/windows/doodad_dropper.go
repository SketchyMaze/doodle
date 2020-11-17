package windows

import (
	"fmt"
	"math"

	"git.kirsle.net/apps/doodle/pkg/balance"
	"git.kirsle.net/apps/doodle/pkg/doodads"
	"git.kirsle.net/apps/doodle/pkg/level"
	"git.kirsle.net/apps/doodle/pkg/log"
	"git.kirsle.net/apps/doodle/pkg/uix"
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

		// pagination values
		page           = 1
		pages          int
		perPage        = 20
		maxPageButtons = 10
	)

	window := ui.NewWindow(title)
	window.SetButtons(ui.CloseButton)
	window.Configure(ui.Config{
		Width:      width,
		Height:     height,
		Background: render.Grey,
	})

	frame := ui.NewFrame("Window Body Frame")
	window.Pack(frame, ui.Pack{
		Side:   ui.N,
		Fill:   true,
		Expand: true,
	})

	/*******
	 * Display the Doodads in rows of buttons
	 *******/

	doodadsAvailable, err := doodads.ListDoodads()
	if err != nil {
		log.Error("NewDoodadDropper: doodads.ListDoodads: %s", err)
	}

	// Load all the doodads, skip hidden ones.
	var items []*doodads.Doodad
	for _, filename := range doodadsAvailable {
		doodad, err := doodads.LoadFile(filename)
		if err != nil {
			log.Error(err.Error())
			doodad = doodads.New(balance.DoodadSize)
		}

		// Skip hidden doodads.
		if doodad.Hidden && !balance.ShowHiddenDoodads {
			continue
		}

		doodad.Filename = filename
		items = append(items, doodad)
	}

	// Compute the number of pages for the pager widget.
	pages = int(
		math.Ceil(
			float64(len(items)) / float64(columns*rows),
		),
	)

	// Draw the doodad buttons in rows.
	var btnRows = []*ui.Frame{}
	{
		var (
			row      *ui.Frame
			rowCount int // for labeling the ui.Frame for each row

			// TODO: pre-size btnRows by calculating how many needed
		)

		for i, doodad := range items {
			doodad := doodad

			if row == nil || i%columns == 0 {
				var hidden = rowCount >= rows
				rowCount++

				row = ui.NewFrame(fmt.Sprintf("Doodad Row %d", rowCount))
				row.SetBackground(balance.DoodadButtonBackground)
				btnRows = append(btnRows, row)
				frame.Pack(row, ui.Pack{
					Side: ui.N,
					// Fill: true,
				})

				// Hide overflowing rows until we scroll to them.
				if hidden {
					row.Hide()
				}
			}

			can := uix.NewCanvas(int(buttonSize), true)
			can.Name = doodad.Title
			can.SetBackground(balance.DoodadButtonBackground)
			can.LoadDoodad(doodad)

			btn := ui.NewButton(doodad.Title, can)
			btn.Resize(render.NewRect(
				buttonSize-2, // TODO: without the -2 the button border
				buttonSize-2, // rests on top of the window border
			))
			row.Pack(btn, ui.Pack{
				Side: ui.W,
			})

			// Tooltip hover to show the doodad's name.
			ui.NewTooltip(btn, ui.Tooltip{
				Text: doodad.Title,
				Edge: ui.Top,
			})

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
		}
	}

	{
		/******************
		 * Confirm/cancel buttons.
		 ******************/

		bottomFrame := ui.NewFrame("Button Frame")
		frame.Pack(bottomFrame, ui.Pack{
			Side:  ui.S,
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

	window.Hide()
	return window
}
