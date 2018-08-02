package doodle

import (
	"git.kirsle.net/apps/doodle/events"
	"git.kirsle.net/apps/doodle/render"
	"git.kirsle.net/apps/doodle/ui"
)

// MainScene implements the main menu of Doodle.
type MainScene struct {
	Supervisor *ui.Supervisor
	frame      *ui.Frame
}

// Name of the scene.
func (s *MainScene) Name() string {
	return "Main"
}

// Setup the scene.
func (s *MainScene) Setup(d *Doodle) error {
	s.Supervisor = ui.NewSupervisor()

	frame := ui.NewFrame("frame")
	s.frame = frame
	s.frame.Configure(ui.Config{
		// Width:       400,
		// Height:      200,
		Background:  render.Purple,
		BorderStyle: ui.BorderSolid,
		BorderSize:  1,
		BorderColor: render.Blue,
	})

	button1 := ui.NewButton("Button1", ui.NewLabel(render.Text{
		Text:  "New Map",
		Size:  14,
		Color: render.Black,
	}))
	// button1.Compute(d.Engine)
	// button1.MoveTo(render.Point{
	// 	X: (d.width / 2) - (button1.Size().W / 2),
	// 	Y: 200,
	// })
	button1.Handle("Click", func(p render.Point) {
		d.NewMap()
	})

	button2 := ui.NewButton("Button2", ui.NewLabel(render.Text{
		Text:  "New Map",
		Size:  14,
		Color: render.Black,
	}))
	button2.SetText("Load Map")
	// button2.Compute(d.Engine)
	// button2.MoveTo(render.Point{
	// 	X: (d.width / 2) - (button2.Size().W / 2),
	// 	Y: 260,
	// })

	var align = ui.E
	frame.Pack(button1, ui.Pack{
		Anchor:  align,
		Padding: 12,
		Fill:    true,
	})
	frame.Pack(button2, ui.Pack{
		Anchor:  align,
		Padding: 12,
		Fill:    true,
	})

	s.Supervisor.Add(button1)
	s.Supervisor.Add(button2)

	return nil
}

// Loop the editor scene.
func (s *MainScene) Loop(d *Doodle, ev *events.State) error {
	s.Supervisor.Loop(ev)
	return nil
}

// Draw the pixels on this frame.
func (s *MainScene) Draw(d *Doodle) error {
	// Clear the canvas and fill it with white.
	d.Engine.Clear(render.White)

	label := ui.NewLabel(render.Text{
		Text:   "Doodle v" + Version,
		Size:   26,
		Color:  render.Pink,
		Stroke: render.SkyBlue,
		Shadow: render.Black,
	})
	label.Compute(d.Engine)
	label.MoveTo(render.Point{
		X: (d.width / 2) - (label.Size().W / 2),
		Y: 120,
	})
	label.Present(d.Engine)

	s.frame.Compute(d.Engine)
	s.frame.MoveTo(render.Point{
		X: (d.width / 2) - (s.frame.Size().W / 2),
		Y: 200,
	})
	s.frame.Present(d.Engine)

	s.Supervisor.Present(d.Engine)

	return nil
}

// Destroy the scene.
func (s *MainScene) Destroy() error {
	return nil
}
