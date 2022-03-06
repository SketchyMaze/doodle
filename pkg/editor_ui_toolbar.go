package doodle

import (
	"fmt"

	"git.kirsle.net/apps/doodle/pkg/balance"
	"git.kirsle.net/apps/doodle/pkg/drawtool"
	"git.kirsle.net/apps/doodle/pkg/enum"
	"git.kirsle.net/apps/doodle/pkg/sprites"
	"git.kirsle.net/apps/doodle/pkg/usercfg"
	"git.kirsle.net/go/render"
	"git.kirsle.net/go/ui"
	"git.kirsle.net/go/ui/style"
)

// Global toolbarWidth, TODO: editor_ui.go wants it
var toolbarWidth int

// SetupToolbar configures the UI for the Tools panel.
func (u *EditorUI) SetupToolbar(d *Doodle) *ui.Frame {
	// Horizontal toolbar instead of vertical?
	var (
		toolbarSpriteSize = 24 // size of sprite images
		frameSize         render.Rect
		isHoz             = usercfg.Current.HorizontalToolbars
		buttonsPerRow     = 2
		packAlign         = ui.N
		tooltipEdge       = ui.Right
		btnRowPack        = ui.Pack{
			Side: packAlign,
			PadY: 1,
			Fill: true,
		}
		btnPack = ui.Pack{
			Side: ui.W,
			PadX: 1,
		}
	)
	if isHoz {
		packAlign = ui.W
		tooltipEdge = ui.Bottom
		btnRowPack = ui.Pack{
			Side: packAlign,
			PadX: 2,
		}
		btnPack = ui.Pack{
			Side: ui.N,
			PadY: 1,
		}
	}

	// Button Layout Controls:
	// We can draw 2 buttons per row, but for very small screens
	// e.g. mobile in portrait orientation, draw 1 button per row.
	buttonsPerRow = 1
	if isHoz {
		if d.width < enum.ScreenWidthSmall {
			// Narrow screens
			buttonsPerRow = 2
		}
	} else {
		if d.width >= enum.ScreenWidthSmall {
			// Screen wider than 600px = can spare room for 2 buttons per row.
			buttonsPerRow = 2
		}
	}

	// Compute toolbar size to accommodate all buttons (+10 for borders/padding)
	toolbarWidth = buttonsPerRow * (toolbarSpriteSize + 10)
	frameSize = render.NewRect(toolbarWidth, 100)

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
			Value:   drawtool.PanTool.String(),
			Icon:    "assets/sprites/pan-tool.png",
			Tooltip: "Pan Tool",
			Click: func() {
				u.Canvas.Tool = drawtool.PanTool
				d.Flash("Pan Tool selected.")
			},
		},

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
			Value:   drawtool.TextTool.String(),
			Icon:    "assets/sprites/text-tool.png",
			Tooltip: "Text Tool",
			Click: func() {
				u.Canvas.Tool = drawtool.TextTool
				u.OpenTextTool()
				d.Flash("Text Tool selected.")
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
				u.OpenDoodadDropper()
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

	// Arrange the buttons 2x2.
	var btnRow *ui.Frame
	for i, button := range buttons {
		button := button
		if button.NoDoodad && u.Scene.DrawingType == enum.DoodadDrawing {
			continue
		}

		if buttonsPerRow == 1 || i%buttonsPerRow == 0 {
			btnRow = ui.NewFrame(fmt.Sprintf("Button Row %d", i))
			btnFrame.Pack(btnRow, btnRowPack)
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
		btn.SetBorderSize(1)
		btn.Resize(render.NewRect(btnSize, btnSize))

		btn.Handle(ui.Click, func(ed ui.EventData) error {
			button.Click()
			return nil
		})
		u.Supervisor.Add(btn)

		tt := ui.NewTooltip(btn, ui.Tooltip{
			Text: button.Tooltip,
			Edge: tooltipEdge,
		})
		tt.Supervise(u.Supervisor)

		btnRow.Pack(btn, btnPack)
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
			btn.SetBorderSize(1)
			btn.Handle(ui.Click, func(ed ui.EventData) error {
				button.F()
				return nil
			})
			u.Supervisor.Add(btn)

			// Which side to pack on?
			var side = ui.W
			if !isHoz && buttonsPerRow == 1 {
				// Vertical layout w/ narrow one-button-per-row, the +-
				// buttons stick out so stack them vertically.
				side = ui.S
			}
			sizeBtnFrame.Pack(btn, ui.Pack{
				Side: side,
			})
		}
	}

	frame.Compute(d.Engine)

	return frame
}
