package doodle

import (
	"git.kirsle.net/apps/doodle/pkg/balance"
	"git.kirsle.net/apps/doodle/pkg/drawtool"
	"git.kirsle.net/apps/doodle/pkg/enum"
	"git.kirsle.net/apps/doodle/pkg/sprites"
	"git.kirsle.net/apps/doodle/pkg/usercfg"
	"git.kirsle.net/go/render"
	"git.kirsle.net/go/ui"
	"git.kirsle.net/go/ui/style"
)

// Width of the toolbar frame.
var toolbarWidth = 44      // 38px button (32px sprite + borders) + padding
var toolbarSpriteSize = 32 // 32x32 sprites.

// SetupToolbar configures the UI for the Tools panel.
func (u *EditorUI) SetupToolbar(d *Doodle) *ui.Frame {
	// Horizontal toolbar instead of vertical?
	var (
		isHoz       = usercfg.Current.HorizontalToolbars
		packAlign   = ui.N
		frameSize   = render.NewRect(toolbarWidth, 100)
		tooltipEdge = ui.Right
		btnPack     = ui.Pack{
			Side: packAlign,
			PadY: 2,
		}
	)
	if isHoz {
		packAlign = ui.W
		frameSize = render.NewRect(100, toolbarWidth)
		tooltipEdge = ui.Bottom
		btnPack = ui.Pack{
			Side: packAlign,
			PadX: 2,
		}
	}

	frame := ui.NewFrame("Tool Bar")
	frame.Resize(frameSize)
	frame.Configure(ui.Config{
		BorderSize:  2,
		BorderStyle: ui.BorderRaised,
		Background:  render.Grey,
	})

	btnFrame := ui.NewFrame("Tool Buttons")
	frame.Pack(btnFrame, ui.Pack{
		Side: packAlign,
	})

	// Buttons.
	var buttons = []struct {
		Value   string
		Icon    string
		Tooltip string
		Style   *style.Button
		Click   func()

		// Optional fields.
		NoDoodad bool // tool not available for Doodad editing (Levels only)
	}{
		{
			Value:   drawtool.PencilTool.String(),
			Icon:    "assets/sprites/pencil-tool.png",
			Tooltip: "Pencil Tool",
			Click: func() {
				u.Canvas.Tool = drawtool.PencilTool
				d.Flash("Pencil Tool selected.")
			},
		},

		{
			Value:   drawtool.LineTool.String(),
			Icon:    "assets/sprites/line-tool.png",
			Tooltip: "Line Tool",
			Click: func() {
				u.Canvas.Tool = drawtool.LineTool
				d.Flash("Line Tool selected.")
			},
		},

		{
			Value:   drawtool.RectTool.String(),
			Icon:    "assets/sprites/rect-tool.png",
			Tooltip: "Rectangle Tool",
			Click: func() {
				u.Canvas.Tool = drawtool.RectTool
				d.Flash("Rectangle Tool selected.")
			},
		},

		{
			Value:   drawtool.EllipseTool.String(),
			Icon:    "assets/sprites/ellipse-tool.png",
			Tooltip: "Ellipse Tool",
			Click: func() {
				u.Canvas.Tool = drawtool.EllipseTool
				d.Flash("Ellipse Tool selected.")
			},
		},

		{
			Value:    drawtool.ActorTool.String(),
			Icon:     "assets/sprites/actor-tool.png",
			Tooltip:  "Doodad Tool\nDrag-and-drop objects into your map",
			NoDoodad: true,
			Style:    &balance.ButtonBabyBlue,
			Click: func() {
				u.Canvas.Tool = drawtool.ActorTool
				u.doodadWindow.Show()
				d.Flash("Actor Tool selected. Drag a Doodad from the drawer into your level.")
			},
		},

		{
			Value:    drawtool.LinkTool.String(),
			Icon:     "assets/sprites/link-tool.png",
			Tooltip:  "Link Tool\nConnect doodads to each other",
			Style:    &balance.ButtonPink,
			NoDoodad: true,
			Click: func() {
				u.Canvas.Tool = drawtool.LinkTool
				d.Flash("Link Tool selected. Click a doodad in your level to link it to another.")
			},
		},

		{
			Value:   drawtool.EraserTool.String(),
			Icon:    "assets/sprites/eraser-tool.png",
			Tooltip: "Eraser Tool",
			Style:   &balance.ButtonLightRed,
			Click: func() {
				u.Canvas.Tool = drawtool.EraserTool

				// Set the brush size within range for the eraser.
				if u.Canvas.BrushSize < balance.DefaultEraserBrushSize {
					u.Canvas.BrushSize = balance.DefaultEraserBrushSize
				} else if u.Canvas.BrushSize > balance.MaxEraserBrushSize {
					u.Canvas.BrushSize = balance.MaxEraserBrushSize
				}

				d.Flash("Eraser Tool selected.")
			},
		},
	}
	for _, button := range buttons {
		button := button
		if button.NoDoodad && u.Scene.DrawingType == enum.DoodadDrawing {
			continue
		}

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
		if button.Style != nil {
			btn.SetStyle(button.Style)
		}

		var btnSize = btn.BoxThickness(2) + toolbarSpriteSize
		btn.Resize(render.NewRect(btnSize, btnSize))

		btn.Handle(ui.Click, func(ed ui.EventData) error {
			button.Click()
			return nil
		})
		u.Supervisor.Add(btn)

		ui.NewTooltip(btn, ui.Tooltip{
			Text: button.Tooltip,
			Edge: tooltipEdge,
		})

		btnFrame.Pack(btn, btnPack)
	}

	// Doodad Editor: show the Layers button.
	if u.Scene.DrawingType == enum.DoodadDrawing {
		btn := ui.NewButton("Layers Button", ui.NewLabel(ui.Label{
			Text: "Lyr.",
			Font: balance.MenuFont,
		}))
		btn.Handle(ui.Click, func(ed ui.EventData) error {
			u.OpenLayersWindow()
			return nil
		})
		u.Supervisor.Add(btn)
		btnFrame.Pack(btn, ui.Pack{
			Side: packAlign,
			PadY: 2,
		})
	}

	// Spacer frame.
	frame.Pack(ui.NewFrame("spacer"), ui.Pack{
		Side: packAlign,
		PadY: 8,
	})

	//////////////
	// "Brush Size" label
	bsFrame := ui.NewFrame("Brush Size Frame")
	frame.Pack(bsFrame, ui.Pack{
		Side: packAlign,
	})

	bsLabel := ui.NewLabel(ui.Label{
		Text: "Size:",
		Font: balance.SmallFont,
	})
	bsFrame.Pack(bsLabel, ui.Pack{
		Side: ui.N,
	})

	ui.NewTooltip(bsLabel, ui.Tooltip{
		Text: "Set the line thickness for drawing",
		Edge: tooltipEdge,
	})
	u.Supervisor.Add(bsLabel)

	sizeLabel := ui.NewLabel(ui.Label{
		IntVariable: &u.Canvas.BrushSize,
		Font:        balance.SmallFont,
	})
	sizeLabel.Configure(ui.Config{
		BorderSize:  1,
		BorderStyle: ui.BorderSunken,
		Background:  render.Grey,
	})
	bsFrame.Pack(sizeLabel, ui.Pack{
		Side: ui.N,
		// FillX: true,
		PadY: 0,
	})

	// Brush Size widget
	{
		sizeFrame := ui.NewFrame("Brush Size Frame")
		frame.Pack(sizeFrame, ui.Pack{
			Side: packAlign,
			PadY: 0,
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
			btn.Handle(ui.Click, func(ed ui.EventData) error {
				button.F()
				return nil
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
