package windows

import (
	"fmt"
	"math"

	"git.kirsle.net/SketchyMaze/doodle/pkg/balance"
	"git.kirsle.net/SketchyMaze/doodle/pkg/doodads"
	"git.kirsle.net/SketchyMaze/doodle/pkg/shmem"
	"git.kirsle.net/go/render"
	"git.kirsle.net/go/ui"
)

// Layers shows the layers when editing a doodad file.
type Layers struct {
	Supervisor *ui.Supervisor
	Engine     render.Engine

	// Pointer to the currently edited level.
	EditDoodad  *doodads.Doodad
	ActiveLayer int    // pointer to selected layer
	activeLayer string // cached string for radio button

	// Callback functions.
	OnChange   func(*doodads.Doodad) // Doodad data was modified, reload the Canvas etc.
	OnAddLayer func()                // "Add Layer" button was clicked
	OnCancel   func()                // Close button was clicked.

	// Editor should change the active layer
	OnChangeLayer func(index int)
}

// NewLayerWindow initializes the window.
func NewLayerWindow(config Layers) *ui.Window {
	// Default options.
	var (
		title = "Layers"
		rows  = []*ui.Frame{}

		// size of the popup window
		width  = 320
		height = 300

		// Column sizes of the palette table.
		col1 = 40  // Index
		col3 = 120 // Name
		col4 = 60  // Edit button
		// col5 = 150 // Delete

		// pagination values
		page    = 1
		perPage = 5
	)

	config.activeLayer = fmt.Sprintf("%d", config.ActiveLayer)

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

	// Draw the header row.
	headers := []struct {
		Name string
		Size int
	}{
		{"Index", col1},
		{"Name", col3},
		{"Edit", col4},
	}
	header := ui.NewFrame("Header")
	for _, col := range headers {
		labelFrame := ui.NewFrame(col.Name)
		labelFrame.Configure(ui.Config{
			Width:  col.Size,
			Height: 24,
		})

		label := ui.NewLabel(ui.Label{
			Text: col.Name,
			Font: balance.MenuFontBold,
		})
		labelFrame.Pack(label, ui.Pack{
			Side: ui.N,
		})

		header.Pack(labelFrame, ui.Pack{
			Side:    ui.W,
			Padding: 2,
		})
	}

	header.Compute(config.Engine)
	frame.Pack(header, ui.Pack{
		Side: ui.N,
	})

	// Draw the rows for each Layer in the given doodad.
	if doodad := config.EditDoodad; doodad != nil {
		for i, _ := range doodad.Layers {
			i := i // rescope
			var idStr = fmt.Sprintf("%d", i)

			row := ui.NewFrame("Layer " + idStr)
			rows = append(rows, row)

			// Off the end of the first page?
			if i >= perPage {
				row.Hide()
			}

			// ID label.
			idLabel := ui.NewLabel(ui.Label{
				Text: idStr + ".",
				Font: balance.MenuFont,
			})
			idLabel.Configure(ui.Config{
				Width:  col1,
				Height: 24,
			})

			// Name button (click to rename the swatch)
			btnName := ui.NewButton("Name", ui.NewLabel(ui.Label{
				TextVariable: &doodad.Layers[i].Name,
			}))
			btnName.Configure(ui.Config{
				Width:  col3,
				Height: 24,
			})
			btnName.Handle(ui.Click, func(ed ui.EventData) error {
				shmem.Prompt("New layer name ["+doodad.Layers[i].Name+"]: ", func(answer string) {
					if answer != "" {
						doodad.Layers[i].Name = answer
						if config.OnChange != nil {
							config.OnChange(config.EditDoodad)
						}
					}
				})
				return nil
			})
			config.Supervisor.Add(btnName)

			// Edit button (open layer for editing)
			// btnEdit := ui.NewButton("Edit", ui.NewLabel(ui.Label{
			// 	Text: "Edit",
			// }))
			btnEdit := ui.NewRadioButton("Edit",
				&config.activeLayer, idStr, ui.NewLabel(ui.Label{
					Text: "Edit",
				}))
			btnEdit.Configure(ui.Config{
				Width:  col4,
				Height: 24,
			})
			btnEdit.Handle(ui.Click, func(ed ui.EventData) error {
				if config.OnChangeLayer != nil {
					config.OnChangeLayer(i)
				}
				return nil
			})
			config.Supervisor.Add(btnEdit)

			// Pack all the widgets.
			row.Pack(idLabel, ui.Pack{
				Side: ui.W,
				PadX: 2,
			})
			row.Pack(btnName, ui.Pack{
				Side: ui.W,
				PadX: 2,
			})
			row.Pack(btnEdit, ui.Pack{
				Side: ui.W,
				PadX: 2,
			})

			row.Compute(config.Engine)
			frame.Pack(row, ui.Pack{
				Side: ui.N,
				PadY: 2,
			})
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
			Name: "Layers Window Pager",
			Page: page,
			Pages: int(math.Ceil(
				float64(len(rows)) / float64(perPage),
			)),
			PerPage:        perPage,
			MaxPageButtons: 10,
			Font:           balance.MenuFont,
			OnChange: func(newPage, perPage int) {
				page = newPage

				// Re-evaluate which rows are shown/hidden for this page.
				var (
					minRow  = (page - 1) * perPage
					visible = 0
				)
				for i, row := range rows {
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
		bottomFrame.Place(pager, ui.Place{
			Top:  20,
			Left: 20,
		})

		btnFrame := ui.NewFrame("Window Buttons")
		var buttons = []struct {
			Label string
			F     func(ui.EventData) error
		}{
			{"Add Layer", func(ed ui.EventData) error {
				if config.OnAddLayer != nil {
					config.OnAddLayer()
				}

				if config.OnChange != nil {
					config.OnChange(config.EditDoodad)
				}
				return nil
			}},
			{"Close", func(ed ui.EventData) error {
				if config.OnCancel != nil {
					config.OnCancel()
				}
				return nil
			}},
		}
		for _, t := range buttons {
			btn := ui.NewButton(t.Label, ui.NewLabel(ui.Label{
				Text: t.Label,
				Font: balance.MenuFont,
			}))
			btn.Handle(ui.Click, t.F)
			btn.Compute(config.Engine)
			config.Supervisor.Add(btn)

			btnFrame.Pack(btn, ui.Pack{
				Side: ui.W,
				PadX: 4,
			})
		}
		bottomFrame.Place(btnFrame, ui.Place{
			Top:    60,
			Center: true,
		})
	}

	window.Hide()
	return window
}
