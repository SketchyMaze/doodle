package doodle

import (
	"git.kirsle.net/apps/doodle/lib/events"
	"git.kirsle.net/apps/doodle/lib/render"
	"git.kirsle.net/apps/doodle/lib/ui"
	"git.kirsle.net/apps/doodle/pkg/balance"
	"git.kirsle.net/apps/doodle/pkg/branding"
	"git.kirsle.net/apps/doodle/pkg/level"
	"git.kirsle.net/apps/doodle/pkg/log"
	"git.kirsle.net/apps/doodle/pkg/uix"
)

// MainScene implements the main menu of Doodle.
type MainScene struct {
	Supervisor *ui.Supervisor
	frame      *ui.Frame

	// Background wallpaper canvas.
	canvas *uix.Canvas
}

// Name of the scene.
func (s *MainScene) Name() string {
	return "Main"
}

// Setup the scene.
func (s *MainScene) Setup(d *Doodle) error {
	s.Supervisor = ui.NewSupervisor()

	// Set up the background wallpaper canvas.
	s.canvas = uix.NewCanvas(100, false)
	s.canvas.Resize(render.Rect{
		W: int32(d.width),
		H: int32(d.height),
	})
	s.canvas.LoadLevel(d.Engine, &level.Level{
		Chunker:   level.NewChunker(100),
		Palette:   level.NewPalette(),
		PageType:  level.Bounded,
		Wallpaper: "notebook.png",
	})

	// Main UI button frame.
	frame := ui.NewFrame("frame")
	s.frame = frame

	button1 := ui.NewButton("Button1", ui.NewLabel(ui.Label{
		Text: "New Map",
		Font: balance.StatusFont,
	}))
	button1.Handle(ui.Click, func(p render.Point) {
		d.GotoNewMenu()
	})

	button2 := ui.NewButton("Button2", ui.NewLabel(ui.Label{
		Text: "Load Map",
		Font: balance.StatusFont,
	}))
	button2.Handle(ui.Click, func(p render.Point) {
		d.GotoLoadMenu()
	})

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

	if resized := ev.Resized.Read(); resized {
		w, h := d.Engine.WindowSize()
		d.width = w
		d.height = h
		log.Info("Resized to %dx%d", d.width, d.height)
		s.canvas.Resize(render.Rect{
			W: int32(d.width),
			H: int32(d.height),
		})
	}

	return nil
}

// Draw the pixels on this frame.
func (s *MainScene) Draw(d *Doodle) error {
	// Clear the canvas and fill it with white.
	d.Engine.Clear(render.White)

	s.canvas.Present(d.Engine, render.Origin)

	label := ui.NewLabel(ui.Label{
		Text: branding.AppName,
		Font: render.Text{
			Size:   26,
			Color:  render.Pink,
			Stroke: render.SkyBlue,
			Shadow: render.Black,
		},
	})
	label.Compute(d.Engine)
	label.MoveTo(render.Point{
		X: (int32(d.width) / 2) - (label.Size().W / 2),
		Y: 120,
	})
	label.Present(d.Engine, label.Point())

	s.frame.Compute(d.Engine)
	s.frame.MoveTo(render.Point{
		X: (int32(d.width) / 2) - (s.frame.Size().W / 2),
		Y: 200,
	})
	s.frame.Present(d.Engine, s.frame.Point())

	return nil
}

// Destroy the scene.
func (s *MainScene) Destroy() error {
	return nil
}
