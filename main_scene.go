package doodle

import (
	"git.kirsle.net/apps/doodle/events"
	"git.kirsle.net/apps/doodle/render"
	"git.kirsle.net/apps/doodle/ui"
)

// MainScene implements the main menu of Doodle.
type MainScene struct {
}

// Name of the scene.
func (s *MainScene) Name() string {
	return "Main"
}

// Setup the scene.
func (s *MainScene) Setup(d *Doodle) error {
	return nil
}

// Loop the editor scene.
func (s *MainScene) Loop(d *Doodle, ev *events.State) error {
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

	button := ui.NewButton(*ui.NewLabel(render.Text{
		Text:  "New Map",
		Size:  14,
		Color: render.Black,
	}))
	button.Compute(d.Engine)

	button.MoveTo(render.Point{
		X: (d.width / 2) - (button.Size().W / 2),
		Y: 200,
	})
	button.Present(d.Engine)

	button.SetText("Load Map")
	button.Compute(d.Engine)
	button.MoveTo(render.Point{
		X: (d.width / 2) - (button.Size().W / 2),
		Y: 260,
	})
	button.Present(d.Engine)

	return nil
}

// Destroy the scene.
func (s *MainScene) Destroy() error {
	return nil
}
