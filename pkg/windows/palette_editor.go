package windows

import (
	"fmt"
	"math"

	"git.kirsle.net/apps/doodle/pkg/balance"
	"git.kirsle.net/apps/doodle/pkg/level"
	"git.kirsle.net/apps/doodle/pkg/log"
	"git.kirsle.net/apps/doodle/pkg/shmem"
	"git.kirsle.net/go/render"
	"git.kirsle.net/go/ui"
	"git.kirsle.net/go/ui/style"
)

// PaletteEditor lets you customize the level palette in Edit Mode.
type PaletteEditor struct {
	Supervisor *ui.Supervisor
	Engine     render.Engine
	IsDoodad   bool // you're editing a doodad instead of a level?

	// Pointer to the currently edited palette, be it
	// from a level or a doodad.
	EditPalette *level.Palette

	// Callback functions.
	OnChange   func()
	OnAddColor func()
	OnCancel   func()
}

// NewPaletteEditor initializes the window.
func NewPaletteEditor(config PaletteEditor) *ui.Window {
	// Default options.
	var (
		title = "Level Palette"

		buttonSize = balance.DoodadButtonSize
		columns    = balance.DoodadDropperCols
		rows       = []*ui.Frame{}

		// size of the popup window
		width  = buttonSize * columns
		height = (buttonSize * balance.DoodadDropperRows) + 64 // account for button borders :(

		// Column sizes of the palette table.
		col1 = 30  // ID no.
		col2 = 24  // Color
		col3 = 130 // Name
		col4 = 140 // Attributes
		// col5 = 150 // Delete

		// pagination values
		page    = 1
		perPage = 5
	)
	if config.IsDoodad {
		title = "Doodad Palette"
	}

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
		{"ID", col1},
		{"Col", col2},
		{"Name", col3},
		{"Attributes", col4},
		// {"Delete", col5},
	}
	header := ui.NewFrame("Palette Header")
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

	// Draw the main table of Palette rows.
	if pal := config.EditPalette; pal != nil {
		for i, swatch := range pal.Swatches {
			var idStr = fmt.Sprintf("%d", i)
			swatch := swatch

			row := ui.NewFrame("Swatch " + idStr)
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
				TextVariable: &swatch.Name,
			}))
			btnName.Configure(ui.Config{
				Width:  col3,
				Height: 24,
			})
			btnName.Handle(ui.Click, func(ed ui.EventData) error {
				shmem.Prompt("New swatch name ["+swatch.Name+"]: ", func(answer string) {
					log.Warn("Answer: %s", answer)
					if answer != "" {
						swatch.Name = answer
						if config.OnChange != nil {
							config.OnChange()
						}
					}
				})
				return nil
			})
			config.Supervisor.Add(btnName)

			// Color Choice button.
			btnColor := ui.NewButton("Color", ui.NewFrame("Color Frame"))
			btnColor.SetStyle(&style.Button{
				Background:      swatch.Color,
				HoverBackground: swatch.Color.Lighten(40),
				OutlineColor:    render.Black,
				OutlineSize:     1,
				BorderStyle:     style.BorderRaised,
				BorderSize:      2,
			})
			btnColor.Configure(ui.Config{
				Background: swatch.Color,
				Width:      col2,
				Height:     24,
			})
			btnColor.Handle(ui.Click, func(ed ui.EventData) error {
				shmem.Prompt(fmt.Sprintf(
					"New color in hex notation [%s]: ", swatch.Color.ToHex()), func(answer string) {
					if answer != "" {
						color, err := render.HexColor(answer)
						if err != nil {
							shmem.Flash("Error with that color code: %s", err)
							return
						}

						swatch.Color = color

						// TODO: redundant from above, consolidate these
						fmt.Printf("Set button style to: %s\n", swatch.Color)
						btnColor.SetStyle(&style.Button{
							Background:      swatch.Color,
							HoverBackground: swatch.Color.Lighten(40),
							OutlineColor:    render.Black,
							OutlineSize:     1,
							BorderStyle:     style.BorderRaised,
							BorderSize:      2,
						})

						if config.OnChange != nil {
							config.OnChange()
						}
					}
				})
				return nil
			})
			config.Supervisor.Add(btnColor)

			// Attribute flags.
			attrFrame := ui.NewFrame("Attributes")
			attrFrame.Configure(ui.Config{
				Width:  col4,
				Height: 24,
			})
			attributes := []struct {
				Label string
				Var   *bool
			}{
				{
					Label: "Solid",
					Var:   &swatch.Solid,
				},
				{
					Label: "Fire",
					Var:   &swatch.Fire,
				},
				{
					Label: "Water",
					Var:   &swatch.Water,
				},
			}

			// Do not show in Doodad editing mode.
			if !config.IsDoodad {
				for _, attr := range attributes {
					attr := attr
					btn := ui.NewCheckButton(attr.Label, attr.Var, ui.NewLabel(ui.Label{
						Text: attr.Label,
						Font: balance.MenuFont,
					}))
					btn.Handle(ui.Click, func(ed ui.EventData) error {
						if config.OnChange != nil {
							config.OnChange()
						}
						return nil
					})
					config.Supervisor.Add(btn)
					attrFrame.Pack(btn, ui.Pack{
						Side: ui.W,
					})
				}
			}

			// Pack all the widgets.
			row.Pack(idLabel, ui.Pack{
				Side: ui.W,
				PadX: 2,
			})
			row.Pack(btnColor, ui.Pack{
				Side: ui.W,
				PadX: 2,
			})
			row.Pack(btnName, ui.Pack{
				Side: ui.W,
				PadX: 2,
			})
			row.Pack(attrFrame, ui.Pack{
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
			Name: "Palette Editor Pager",
			Page: page,
			Pages: int(math.Ceil(
				float64(len(rows)) / float64(perPage),
			)),
			PerPage:        perPage,
			MaxPageButtons: 6,
			Font:           balance.MenuFont,
			OnChange: func(newPage, perPage int) {
				page = newPage
				log.Info("Page: %d, %d", page, perPage)

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
			{"Add Color", func(ed ui.EventData) error {
				if config.OnAddColor != nil {
					config.OnAddColor()
				}

				if config.OnChange != nil {
					config.OnChange()
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
			Top:   20,
			Right: 20,
		})
	}

	window.Hide()
	return window
}
