package doodle

import (
	"git.kirsle.net/apps/doodle/lib/render"
	"git.kirsle.net/apps/doodle/lib/ui"
	"git.kirsle.net/apps/doodle/pkg/drawtool"
	"git.kirsle.net/apps/doodle/pkg/log"
	"git.kirsle.net/apps/doodle/pkg/sprites"
)

// Width of the toolbar frame.
var toolbarWidth int32 = 44      // 38px button (32px sprite + borders) + padding
var toolbarSpriteSize int32 = 32 // 32x32 sprites.

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
		Anchor: ui.N,
	})

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
				u.DoodadTab.Hide()
				u.PaletteTab.Show()
				d.Flash("Pencil Tool selected.")
			},
		},

		{
			Value: drawtool.LineTool.String(),
			Icon:  "assets/sprites/line-tool.png",
			Click: func() {
				u.Canvas.Tool = drawtool.LineTool
				u.DoodadTab.Hide()
				u.PaletteTab.Show()
				d.Flash("Line Tool selected.")
			},
		},

		{
			Value: drawtool.RectTool.String(),
			Icon:  "assets/sprites/rect-tool.png",
			Click: func() {
				u.Canvas.Tool = drawtool.RectTool
				u.DoodadTab.Hide()
				u.PaletteTab.Show()
				d.Flash("Rectangle Tool selected.")
			},
		},

		{
			Value: drawtool.ActorTool.String(),
			Icon:  "assets/sprites/actor-tool.png",
			Click: func() {
				u.Canvas.Tool = drawtool.ActorTool
				u.PaletteTab.Hide()
				u.DoodadTab.Show()
				d.Flash("Actor Tool selected. Drag a Doodad from the drawer into your level.")
			},
		},

		{
			Value: drawtool.LinkTool.String(),
			Icon:  "assets/sprites/link-tool.png",
			Click: func() {
				u.Canvas.Tool = drawtool.LinkTool
				u.PaletteTab.Hide()
				u.DoodadTab.Show()
				d.Flash("Link Tool selected. Click a doodad in your level to link it to another.")
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
			u.activeTool,
			button.Value,
			image,
		)

		var btnSize int32 = btn.BoxThickness(2) + toolbarSpriteSize
		log.Info("BtnSize: %d", btnSize)
		btn.Resize(render.NewRect(btnSize, btnSize))

		btn.Handle(ui.Click, func(p render.Point) {
			button.Click()
		})
		u.Supervisor.Add(btn)

		btnFrame.Pack(btn, ui.Pack{
			Anchor: ui.N,
			PadY:   2,
		})
	}

	frame.Compute(d.Engine)

	return frame
}
