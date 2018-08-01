package doodle

import (
	"git.kirsle.net/apps/doodle/events"
	"git.kirsle.net/apps/doodle/render"
	"git.kirsle.net/apps/doodle/ui"
)

// GUITestScene implements the main menu of Doodle.
type GUITestScene struct {
	Supervisor *ui.Supervisor
	frame      *ui.Frame
	window     *ui.Frame
}

// Name of the scene.
func (s *GUITestScene) Name() string {
	return "Main"
}

// Setup the scene.
func (s *GUITestScene) Setup(d *Doodle) error {
	s.Supervisor = ui.NewSupervisor()

	window := ui.NewFrame()
	s.window = window
	window.Configure(ui.Config{
		Width:       400,
		Height:      400,
		Background:  render.Grey,
		BorderStyle: ui.BorderRaised,
		BorderSize:  2,
	})

	titleBar := ui.NewLabel(render.Text{
		Text:   "Alert",
		Size:   12,
		Color:  render.White,
		Stroke: render.Black,
	})
	titleBar.Configure(ui.Config{
		Background:   render.Blue,
		OutlineSize:  1,
		OutlineColor: render.Black,
	})
	window.Pack(titleBar, ui.Pack{
		Anchor: ui.N,
		FillX:  true,
	})

	msgFrame := ui.NewFrame()
	msgFrame.Configure(ui.Config{
		Background:  render.Grey,
		BorderStyle: ui.BorderRaised,
		BorderSize:  1,
	})
	window.Pack(msgFrame, ui.Pack{
		Anchor:  ui.N,
		Fill:    true,
		Padding: 4,
	})

	btnFrame := ui.NewFrame()
	btnFrame.Configure(ui.Config{
		Background: render.DarkRed,
	})
	window.Pack(btnFrame, ui.Pack{
		Anchor:  ui.N,
		Padding: 4,
	})

	msg := ui.NewLabel(render.Text{
		Text:  "Hello World!",
		Size:  14,
		Color: render.Black,
	})
	msgFrame.Pack(msg, ui.Pack{
		Anchor:  ui.NW,
		Padding: 2,
	})

	button1 := ui.NewButton(*ui.NewLabel(render.Text{
		Text:  "New Map",
		Size:  14,
		Color: render.Black,
	}))
	button1.SetBackground(render.Blue)
	button1.Handle("Click", func(p render.Point) {
		d.NewMap()
	})

	log.Info("Button1 bg: %s", button1.Background())

	button2 := ui.NewButton(*ui.NewLabel(render.Text{
		Text:  "New Map",
		Size:  14,
		Color: render.Black,
	}))
	button2.SetText("Load Map")

	var align = ui.W
	btnFrame.Pack(button1, ui.Pack{
		Anchor:  align,
		Padding: 20,
		Fill:    true,
	})
	btnFrame.Pack(button2, ui.Pack{
		Anchor:  align,
		Padding: 20,
		Fill:    true,
	})

	s.Supervisor.Add(button1)
	s.Supervisor.Add(button2)

	return nil
}

// Loop the editor scene.
func (s *GUITestScene) Loop(d *Doodle, ev *events.State) error {
	s.Supervisor.Loop(ev)
	return nil
}

// Draw the pixels on this frame.
func (s *GUITestScene) Draw(d *Doodle) error {
	// Clear the canvas and fill it with white.
	d.Engine.Clear(render.White)

	label := ui.NewLabel(render.Text{
		Text:   "GUITest Doodle v" + Version,
		Size:   26,
		Color:  render.Pink,
		Stroke: render.SkyBlue,
		Shadow: render.Black,
	})
	label.Compute(d.Engine)
	label.MoveTo(render.Point{
		X: (d.width / 2) - (label.Size().W / 2),
		Y: 40,
	})
	label.Present(d.Engine)

	s.window.Compute(d.Engine)
	s.window.MoveTo(render.Point{
		X: (d.width / 2) - (s.window.Size().W / 2),
		Y: 100,
	})
	s.window.Present(d.Engine)

	s.Supervisor.Present(d.Engine)

	return nil
}

// Destroy the scene.
func (s *GUITestScene) Destroy() error {
	return nil
}
