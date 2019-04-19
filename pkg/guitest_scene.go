package doodle

import (
	"fmt"

	"git.kirsle.net/apps/doodle/lib/events"
	"git.kirsle.net/apps/doodle/lib/render"
	"git.kirsle.net/apps/doodle/lib/ui"
	"git.kirsle.net/apps/doodle/pkg/balance"
	"git.kirsle.net/apps/doodle/pkg/log"
)

// GUITestScene implements the main menu of Doodle.
type GUITestScene struct {
	Supervisor *ui.Supervisor

	// Private widgets.
	Frame  *ui.Frame
	Window *ui.Window
}

// Name of the scene.
func (s *GUITestScene) Name() string {
	return "Main"
}

// Setup the scene.
func (s *GUITestScene) Setup(d *Doodle) error {
	s.Supervisor = ui.NewSupervisor()

	window := ui.NewWindow("Widget Toolkit")
	window.Configure(ui.Config{
		Width:       750,
		Height:      450,
		Background:  render.Grey,
		BorderStyle: ui.BorderRaised,
		BorderSize:  2,
	})
	s.Window = window

	// Left Frame
	leftFrame := ui.NewFrame("Left Frame")
	leftFrame.Configure(ui.Config{
		Background:  render.Grey,
		BorderStyle: ui.BorderSolid,
		BorderSize:  4,
		Width:       100,
	})
	window.Pack(leftFrame, ui.Pack{
		Side:  ui.Left,
		FillY: true,
	})

	// Some left frame buttons.
	for _, label := range []string{"New", "Edit", "Play", "Help"} {
		btn := ui.NewButton("dummy "+label, ui.NewLabel(ui.Label{
			Text: label,
			Font: balance.StatusFont,
		}))
		btn.Handle(ui.Click, func(p render.Point) {
			d.Flash("%s clicked", btn)
		})
		s.Supervisor.Add(btn)
		leftFrame.Pack(btn, ui.Pack{
			Side:  ui.Top,
			FillX: true,
			PadY:  2,
		})
	}

	// Main Frame
	frame := ui.NewFrame("Main Frame")
	frame.Configure(ui.Config{
		Background: render.White,
		BorderSize: 0,
	})
	window.Pack(frame, ui.Pack{
		Side:   ui.Left,
		Expand: true,
		Fill:   true,
	})

	// Right Frame
	rightFrame := ui.NewFrame("Right Frame")
	rightFrame.Configure(ui.Config{
		Background:  render.SkyBlue,
		BorderStyle: ui.BorderSunken,
		BorderSize:  2,
		Width:       80,
	})
	window.Pack(rightFrame, ui.Pack{
		Side: ui.Right,
		Fill: true,
	})

	// A grid of buttons.
	for row := 0; row < 3; row++ {
		rowFrame := ui.NewFrame(fmt.Sprintf("Row%d", row))
		rowFrame.Configure(ui.Config{
			Background: render.RGBA(0, uint8((row*20)+120), 0, 255),
		})
		for col := 0; col < 3; col++ {
			(func(row, col int, frame *ui.Frame) {
				btn := ui.NewButton(fmt.Sprintf("Grid Button %d:%d", col, row),
					ui.NewFrame(fmt.Sprintf("Col%d", col)),
				)
				btn.Configure(ui.Config{
					Height:      20,
					BorderStyle: ui.BorderRaised,
				})
				btn.Handle(ui.Click, func(p render.Point) {
					d.Flash("%s clicked", btn)
				})
				rowFrame.Pack(btn, ui.Pack{
					Side:   ui.Left,
					Expand: true,
					FillX:  true,
				})
				s.Supervisor.Add(btn)
			})(row, col, rowFrame)
		}
		rightFrame.Pack(rowFrame, ui.Pack{
			Side: ui.Top,
			Fill: true,
		})
	}

	// Main frame widgets.
	frame.Pack(ui.NewLabel(ui.Label{
		Text: "Hello World!",
		Font: render.Text{
			Size:  14,
			Color: render.Black,
		},
	}), ui.Pack{
		Side:    ui.Top,
		Anchor:  ui.W,
		Padding: 2,
	})

	cb := ui.NewCheckbox("Overlay",
		&DebugOverlay,
		ui.NewLabel(ui.Label{
			Text: "Toggle Debug Overlay",
			Font: balance.StatusFont,
		}),
	)
	frame.Pack(cb, ui.Pack{
		Side:    ui.Top,
		Anchor:  ui.W,
		Padding: 4,
	})
	cb.Supervise(s.Supervisor)

	frame.Pack(ui.NewLabel(ui.Label{
		Text: "Like Tk!",
		Font: render.Text{
			Size:  16,
			Color: render.Red,
		},
	}), ui.Pack{
		Side:    ui.Bottom,
		Anchor:  ui.E,
		Padding: 8,
	})
	frame.Pack(ui.NewLabel(ui.Label{
		Text: "Frame widget for pack layouts",
		Font: render.Text{
			Size:  14,
			Color: render.Blue,
		},
	}), ui.Pack{
		Side:    ui.Bottom,
		Anchor:  ui.E,
		Padding: 8,
	})

	// Buttom Frame
	btnFrame := ui.NewFrame("btnFrame")
	btnFrame.Configure(ui.Config{
		Background: render.Grey,
	})
	window.Pack(btnFrame, ui.Pack{
		Side: ui.Top,
	})

	button1 := ui.NewButton("Button1", ui.NewLabel(ui.Label{
		Text: "New Map",
		Font: balance.StatusFont,
	}))
	button1.SetBackground(render.Blue)
	button1.Handle(ui.Click, func(p render.Point) {
		d.NewMap()
	})

	log.Info("Button1 bg: %s", button1.Background())

	button2 := ui.NewButton("Button2", ui.NewLabel(ui.Label{
		Text: "Load Map",
		Font: balance.StatusFont,
	}))
	button2.Handle(ui.Click, func(p render.Point) {
		d.Prompt("Map name>", func(name string) {
			d.EditDrawing(name)
		})
	})

	btnFrame.Pack(button1, ui.Pack{
		Side:    ui.Left,
		Padding: 20,
	})
	btnFrame.Pack(button2, ui.Pack{
		Side:    ui.Left,
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

	label := ui.NewLabel(ui.Label{
		Text: "GUITest Doodle v" + Version,
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
		Y: 40,
	})
	label.Present(d.Engine, label.Point())

	s.Window.Compute(d.Engine)
	s.Window.MoveTo(render.Point{
		X: (int32(d.width) / 2) - (s.Window.Size().W / 2),
		Y: 100,
	})
	s.Window.Present(d.Engine, s.Window.Point())

	return nil
}

// Destroy the scene.
func (s *GUITestScene) Destroy() error {
	return nil
}
