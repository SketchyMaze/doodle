package doodle

import (
	"fmt"

	"git.kirsle.net/apps/doodle/events"
	"git.kirsle.net/apps/doodle/render"
	"git.kirsle.net/apps/doodle/ui"
)

// GUITestScene implements the main menu of Doodle.
type GUITestScene struct {
	Supervisor *ui.Supervisor

	// Private widgets.
	Frame  *ui.Frame
	Window *ui.Frame
}

// Name of the scene.
func (s *GUITestScene) Name() string {
	return "Main"
}

// Setup the scene.
func (s *GUITestScene) Setup(d *Doodle) error {
	s.Supervisor = ui.NewSupervisor()

	window := ui.NewFrame("window")
	s.Window = window
	window.Configure(ui.Config{
		Width:       750,
		Height:      450,
		Background:  render.Grey,
		BorderStyle: ui.BorderRaised,
		BorderSize:  2,
	})

	// Title Bar
	titleBar := ui.NewLabel(render.Text{
		Text:   "Widget Toolkit",
		Size:   12,
		Color:  render.White,
		Stroke: render.DarkBlue,
	})
	titleBar.Configure(ui.Config{
		Background: render.Blue,
	})
	window.Pack(titleBar, ui.Pack{
		Anchor: ui.N,
		Fill:   true,
	})

	// Window Body
	body := ui.NewFrame("Window Body")
	body.Configure(ui.Config{
		Background: render.Yellow,
	})
	window.Pack(body, ui.Pack{
		Anchor: ui.N,
		Expand: true,
	})

	// Left Frame
	leftFrame := ui.NewFrame("Left Frame")
	leftFrame.Configure(ui.Config{
		Background:  render.Grey,
		BorderStyle: ui.BorderSolid,
		BorderSize:  4,
		Width:       100,
	})
	body.Pack(leftFrame, ui.Pack{
		Anchor: ui.W,
		FillY:  true,
	})

	// Some left frame buttons.
	for _, label := range []string{"New", "Edit", "Play", "Help"} {
		btn := ui.NewButton("dummy "+label, ui.NewLabel(render.Text{
			Text:  label,
			Size:  12,
			Color: render.Black,
		}))
		s.Supervisor.Add(btn)
		leftFrame.Pack(btn, ui.Pack{
			Anchor: ui.N,
			Fill:   true,
			PadY:   2,
		})
	}

	// Main Frame
	frame := ui.NewFrame("Main Frame")
	frame.Configure(ui.Config{
		Background: render.White,
		BorderSize: 0,
	})
	body.Pack(frame, ui.Pack{
		Anchor: ui.W,
		Expand: true,
	})

	// Right Frame
	rightFrame := ui.NewFrame("Right Frame")
	rightFrame.Configure(ui.Config{
		Background:  render.SkyBlue,
		BorderStyle: ui.BorderSunken,
		BorderSize:  2,
		Width:       80,
	})
	body.Pack(rightFrame, ui.Pack{
		Anchor: ui.W,
		Fill:   true,
	})

	// A grid of buttons.
	for row := 0; row < 3; row++ {
		rowFrame := ui.NewFrame(fmt.Sprintf("Row%d", row))
		for col := 0; col < 3; col++ {
			btn := ui.NewButton("X",
				ui.NewFrame(fmt.Sprintf("Col%d", col)),
			)
			btn.Configure(ui.Config{
				Height:      20,
				BorderStyle: ui.BorderRaised,
			})
			rowFrame.Pack(btn, ui.Pack{
				Anchor: ui.W,
				Expand: true,
			})
			s.Supervisor.Add(btn)
		}
		rightFrame.Pack(rowFrame, ui.Pack{
			Anchor: ui.N,
			Fill:   true,
		})
	}

	frame.Pack(ui.NewLabel(render.Text{
		Text:  "Hello World!",
		Size:  14,
		Color: render.Black,
	}), ui.Pack{
		Anchor:  ui.NW,
		Padding: 2,
	})

	cb := ui.NewCheckbox("Overlay",
		&DebugOverlay,
		ui.NewLabel(render.Text{
			Text:  "Toggle Debug Overlay",
			Size:  14,
			Color: render.Black,
		}),
	)
	frame.Pack(cb, ui.Pack{
		Anchor:  ui.NW,
		Padding: 4,
	})
	cb.Supervise(s.Supervisor)
	frame.Pack(ui.NewLabel(render.Text{
		Text:  "Like Tk!",
		Size:  16,
		Color: render.Red,
	}), ui.Pack{
		Anchor:  ui.SE,
		Padding: 8,
	})
	frame.Pack(ui.NewLabel(render.Text{
		Text:  "Frame widget for pack layouts",
		Size:  14,
		Color: render.Blue,
	}), ui.Pack{
		Anchor:  ui.SE,
		Padding: 8,
	})

	// Buttom Frame
	btnFrame := ui.NewFrame("btnFrame")
	btnFrame.Configure(ui.Config{
		Background: render.Grey,
	})
	window.Pack(btnFrame, ui.Pack{
		Anchor: ui.N,
	})

	button1 := ui.NewButton("Button1", ui.NewLabel(render.Text{
		Text:  "New Map",
		Size:  14,
		Color: render.Black,
	}))
	button1.SetBackground(render.Blue)
	button1.Handle("Click", func(p render.Point) {
		d.NewMap()
	})

	log.Info("Button1 bg: %s", button1.Background())

	button2 := ui.NewButton("Button2", ui.NewLabel(render.Text{
		Text:  "New Map",
		Size:  14,
		Color: render.Black,
	}))
	button2.SetText("Load Map")

	var align = ui.W
	btnFrame.Pack(button1, ui.Pack{
		Anchor:  align,
		Padding: 20,
	})
	btnFrame.Pack(button2, ui.Pack{
		Anchor:  align,
		Padding: 20,
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

	s.Window.Compute(d.Engine)
	s.Window.MoveTo(render.Point{
		X: (d.width / 2) - (s.Window.Size().W / 2),
		Y: 100,
	})
	s.Window.Present(d.Engine)

	s.Supervisor.Present(d.Engine)

	// os.Exit(1)

	return nil
}

// Destroy the scene.
func (s *GUITestScene) Destroy() error {
	return nil
}
