package doodle

import (
	"git.kirsle.net/apps/doodle/balance"
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

	button1 := ui.NewButton("Button1", ui.NewLabel(ui.Label{
		Text: "New Map",
		Font: balance.StatusFont,
	}))
	button1.Handle(ui.Click, func(p render.Point) {
		d.NewMap()
	})

	button2 := ui.NewButton("Button2", ui.NewLabel(ui.Label{
		Text: "New Map",
		Font: balance.StatusFont,
	}))
	button2.SetText("Load Map")

	frame.Pack(button1, ui.Pack{
		Anchor: ui.N,
		Fill:   true,
	})
	frame.Pack(button2, ui.Pack{
		Anchor: ui.N,
		PadY:   12,
		Fill:   true,
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

	label := ui.NewLabel(ui.Label{
		Text: "Doodle v" + Version,
		Font: render.Text{
			Size:   26,
			Color:  render.Pink,
			Stroke: render.SkyBlue,
			Shadow: render.Black,
		},
	})
	label.Compute(d.Engine)
	label.MoveTo(render.Point{
		X: (d.width / 2) - (label.Size().W / 2),
		Y: 120,
	})
	label.Present(d.Engine, label.Point())

	s.frame.Compute(d.Engine)
	s.frame.MoveTo(render.Point{
		X: (d.width / 2) - (s.frame.Size().W / 2),
		Y: 200,
	})
	s.frame.Present(d.Engine, s.frame.Point())

	return nil
}

// Destroy the scene.
func (s *MainScene) Destroy() error {
	return nil
}
