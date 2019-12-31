package doodle

import (
	"git.kirsle.net/apps/doodle/pkg/balance"
	"git.kirsle.net/apps/doodle/pkg/drawtool"
	"git.kirsle.net/apps/doodle/pkg/sprites"
	"git.kirsle.net/go/render"
	"git.kirsle.net/go/ui"
)

// Width of the toolbar frame.
var toolbarWidth = 44      // 38px button (32px sprite + borders) + padding
var toolbarSpriteSize = 32 // 32x32 sprites.

// SetupToolbar configures the UI for the Tools panel.
func (u *EditorUI) SetupToolbar(d *Doodle) *ui.Frame {
	frame := ui.NewFrame("Tool Bar")
	frame.Resize(render.NewRect(toolbarWidth, 100))
	frame.Configure(ui.Config{
		BorderSize:  2,
		BorderStyle: ui.BorderRaised,
		Background:  render.Grey,
	})

	btnFrame := ui.NewFrame("Tool Buttons")
	frame.Pack(btnFrame, ui.Pack{
		Side: ui.N,
	})

	// Helper functions to toggle the correct palette panel.
	var (
		showSwatchPalette = func() {
			u.DoodadTab.Hide()
			u.PaletteTab.Show()
		}
		showDoodadPalette = func() {
			u.PaletteTab.Hide()
			u.DoodadTab.Show()
		}
	)

	// Buttons.
	var buttons = []struct {
		Value string
		Icon  string
		Click func()
	}{
		{
			Value: drawtool.PencilTool.String(),
			Icon:  "assets/sprites/pencil-tool.png",
			Click: func() {
				u.Canvas.Tool = drawtool.PencilTool
				showSwatchPalette()
				d.Flash("Pencil Tool selected.")
			},
		},

		{
			Value: drawtool.LineTool.String(),
			Icon:  "assets/sprites/line-tool.png",
			Click: func() {
				u.Canvas.Tool = drawtool.LineTool
				showSwatchPalette()
				d.Flash("Line Tool selected.")
			},
		},

		{
			Value: drawtool.RectTool.String(),
			Icon:  "assets/sprites/rect-tool.png",
			Click: func() {
				u.Canvas.Tool = drawtool.RectTool
				showSwatchPalette()
				d.Flash("Rectangle Tool selected.")
			},
		},

		{
			Value: drawtool.EllipseTool.String(),
			Icon:  "assets/sprites/ellipse-tool.png",
			Click: func() {
				u.Canvas.Tool = drawtool.EllipseTool
				showSwatchPalette()
				d.Flash("Ellipse Tool selected.")
			},
		},

		{
			Value: drawtool.ActorTool.String(),
			Icon:  "assets/sprites/actor-tool.png",
			Click: func() {
				u.Canvas.Tool = drawtool.ActorTool
				showDoodadPalette()
				d.Flash("Actor Tool selected. Drag a Doodad from the drawer into your level.")
			},
		},

		{
			Value: drawtool.LinkTool.String(),
			Icon:  "assets/sprites/link-tool.png",
			Click: func() {
				u.Canvas.Tool = drawtool.LinkTool
				showDoodadPalette()
				d.Flash("Link Tool selected. Click a doodad in your level to link it to another.")
			},
		},

		{
			Value: drawtool.EraserTool.String(),
			Icon:  "assets/sprites/eraser-tool.png",
			Click: func() {
				u.Canvas.Tool = drawtool.EraserTool

				// Set the brush size within range for the eraser.
				if u.Canvas.BrushSize < balance.DefaultEraserBrushSize {
					u.Canvas.BrushSize = balance.DefaultEraserBrushSize
				} else if u.Canvas.BrushSize > balance.MaxEraserBrushSize {
					u.Canvas.BrushSize = balance.MaxEraserBrushSize
				}

				showSwatchPalette()
				d.Flash("Eraser Tool selected.")
			},
		},
	}
	for _, button := range buttons {
		button := button
		image, err := sprites.LoadImage(d.Engine, button.Icon)
		if err != nil {
			panic(err)
		}

		btn := ui.NewRadioButton(
			button.Value,
			&u.activeTool,
			button.Value,
			image,
		)

		var btnSize = btn.BoxThickness(2) + toolbarSpriteSize
		btn.Resize(render.NewRect(btnSize, btnSize))

		btn.Handle(ui.Click, func(p render.Point) {
			button.Click()
		})
		u.Supervisor.Add(btn)

		btnFrame.Pack(btn, ui.Pack{
			Side: ui.N,
			PadY: 2,
		})
	}

	// Spacer frame.
	frame.Pack(ui.NewFrame("spacer"), ui.Pack{
		Side: ui.N,
		PadY: 8,
	})

	// "Brush Size" label
	bsLabel := ui.NewLabel(ui.Label{
		Text: "Size:",
		Font: balance.LabelFont,
	})
	frame.Pack(bsLabel, ui.Pack{
		Side: ui.N,
	})

	// Brush Size widget
	{
		sizeFrame := ui.NewFrame("Brush Size Frame")
		frame.Pack(sizeFrame, ui.Pack{
			Side: ui.N,
			PadY: 0,
		})

		sizeLabel := ui.NewLabel(ui.Label{
			IntVariable: &u.Canvas.BrushSize,
			Font:        balance.SmallMonoFont,
		})
		sizeLabel.Configure(ui.Config{
			BorderSize:  1,
			BorderStyle: ui.BorderSunken,
			Background:  render.Grey,
		})
		sizeFrame.Pack(sizeLabel, ui.Pack{
			Side:  ui.N,
			FillX: true,
			PadY:  2,
		})

		sizeBtnFrame := ui.NewFrame("Size Increment Button Frame")
		sizeFrame.Pack(sizeBtnFrame, ui.Pack{
			Side:  ui.N,
			FillX: true,
		})

		var incButtons = []struct {
			Label string
			F     func()
		}{
			{
				Label: "-",
				F: func() {
					// Select next smaller brush size.
					for i := len(balance.BrushSizeOptions) - 1; i >= 0; i-- {
						if balance.BrushSizeOptions[i] < u.Canvas.BrushSize {
							u.Canvas.BrushSize = balance.BrushSizeOptions[i]
							break
						}
					}
				},
			},
			{
				Label: "+",
				F: func() {
					// Select next bigger brush size.
					for _, size := range balance.BrushSizeOptions {
						if size > u.Canvas.BrushSize {
							u.Canvas.BrushSize = size
							break
						}
					}

					// Limit the eraser brush size, too big and it's slow because
					// the eraser has to scan and remember pixels to be able to
					// Undo the erase and restore them.
					if u.Canvas.Tool == drawtool.EraserTool && u.Canvas.BrushSize > balance.MaxEraserBrushSize {
						u.Canvas.BrushSize = balance.MaxEraserBrushSize
					}
				},
			},
		}
		for _, button := range incButtons {
			button := button
			btn := ui.NewButton("BrushSize"+button.Label, ui.NewLabel(ui.Label{
				Text: button.Label,
				Font: balance.SmallMonoFont,
			}))
			btn.Handle(ui.Click, func(p render.Point) {
				button.F()
			})
			u.Supervisor.Add(btn)
			sizeBtnFrame.Pack(btn, ui.Pack{
				Side: ui.W,
			})
		}
	}

	frame.Compute(d.Engine)

	return frame
}
