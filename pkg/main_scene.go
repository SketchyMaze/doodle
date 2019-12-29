package doodle

import (
	"git.kirsle.net/apps/doodle/pkg/balance"
	"git.kirsle.net/apps/doodle/pkg/branding"
	"git.kirsle.net/apps/doodle/pkg/level"
	"git.kirsle.net/apps/doodle/pkg/log"
	"git.kirsle.net/apps/doodle/pkg/scripting"
	"git.kirsle.net/apps/doodle/pkg/uix"
	"git.kirsle.net/go/render"
	"git.kirsle.net/go/render/event"
	"git.kirsle.net/go/ui"
)

// MainScene implements the main menu of Doodle.
type MainScene struct {
	Supervisor *ui.Supervisor
	frame      *ui.Frame

	// Background wallpaper canvas.
	scripting *scripting.Supervisor
	canvas    *uix.Canvas
}

// Name of the scene.
func (s *MainScene) Name() string {
	return "Main"
}

// Setup the scene.
func (s *MainScene) Setup(d *Doodle) error {
	s.Supervisor = ui.NewSupervisor()

	if err := s.SetupDemoLevel(d); err != nil {
		return err
	}

	// Main UI button frame.
	frame := ui.NewFrame("frame")
	s.frame = frame

	var buttons = []struct {
		Name string
		Func func()
	}{
		{
			Name: "Play a Level",
			Func: d.GotoPlayMenu,
		},
		{
			Name: "Create a New Level",
			Func: d.GotoNewMenu,
		},
		{
			Name: "Edit a Level",
			Func: d.GotoLoadMenu,
		},
	}
	for _, button := range buttons {
		button := button
		btn := ui.NewButton(button.Name, ui.NewLabel(ui.Label{
			Text: button.Name,
			Font: balance.StatusFont,
		}))
		btn.Handle(ui.Click, func(p render.Point) {
			button.Func()
		})
		s.Supervisor.Add(btn)
		frame.Pack(btn, ui.Pack{
			Side: ui.N,
			PadY:   8,
			// Fill:   true,
			FillX: true,
		})
	}

	return nil
}

// SetupDemoLevel configures the wallpaper behind the New screen,
// which demos a title screen demo level.
func (s *MainScene) SetupDemoLevel(d *Doodle) error {
	// Set up the background wallpaper canvas.
	s.canvas = uix.NewCanvas(100, false)
	s.canvas.Scrollable = true
	s.canvas.Resize(render.Rect{
		W: d.width,
		H: d.height,
	})

	s.scripting = scripting.NewSupervisor()
	s.canvas.SetScriptSupervisor(s.scripting)

	// Title screen level to load.
	if lvl, err := level.LoadFile("example1.level"); err == nil {
		s.canvas.LoadLevel(d.Engine, lvl)
		s.canvas.InstallActors(lvl.Actors)

		// Load all actor scripts.
		if err := s.scripting.InstallScripts(lvl); err != nil {
			log.Error("Error with title screen level scripts: %s", err)
		}

		// Run all actors scripts main function to start them off.
		if err := s.canvas.InstallScripts(); err != nil {
			log.Error("Error running actor main() functions: %s", err)
		}
	} else {
		log.Error("Error loading title-screen.level: %s", err)
	}

	return nil
}

// Loop the editor scene.
func (s *MainScene) Loop(d *Doodle, ev *event.State) error {
	s.Supervisor.Loop(ev)

	if err := s.scripting.Loop(); err != nil {
		log.Error("MainScene.Loop: scripting.Loop: %s", err)
	}

	s.canvas.Loop(ev)

	if ev.WindowResized {
		w, h := d.Engine.WindowSize()
		d.width = w
		d.height = h
		log.Info("Resized to %dx%d", d.width, d.height)
		s.canvas.Resize(render.Rect{
			W: d.width,
			H: d.height,
		})
	}

	return nil
}

// Draw the pixels on this frame.
func (s *MainScene) Draw(d *Doodle) error {
	// Clear the canvas and fill it with white.
	d.Engine.Clear(render.White)

	s.canvas.Present(d.Engine, render.Origin)

	// Draw a sheen over the level for clarity.
	d.Engine.DrawBox(render.RGBA(255, 255, 254, 128), render.Rect{
		X: 0,
		Y: 0,
		W: d.width,
		H: d.height,
	})

	label := ui.NewLabel(ui.Label{
		Text: branding.AppName,
		Font: render.Text{
			Size:   46,
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
		Y: 260,
	})
	s.frame.Present(d.Engine, s.frame.Point())

	return nil
}

// Destroy the scene.
func (s *MainScene) Destroy() error {
	return nil
}
