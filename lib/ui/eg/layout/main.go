package main

import (
	"time"

	"git.kirsle.net/apps/doodle/lib/render"
	"git.kirsle.net/apps/doodle/lib/render/sdl"
	"git.kirsle.net/apps/doodle/lib/ui"
)

var TargetFPS = 1000 / 60

func main() {
	engine := sdl.New("Test Layout GUI", 1024, 768)

	if err := engine.Setup(); err != nil {
		panic(err)
	}

	super := ui.NewSupervisor()

	window := ui.NewWindow("Test Window")
	window.Configure(ui.Config{
		Width:      750,
		Height:     400,
		Background: render.Grey,
	})
	window.MoveTo(render.NewPoint(80, 80))

	leftPanel := ui.NewFrame("Left Panel")
	leftPanel.Configure(ui.Config{
		// AutoResize:  true,
		Width:       200,
		Background:  render.SkyBlue,
		BorderStyle: ui.BorderRaised,
		BorderSize:  2,
	})
	window.Pack(leftPanel, ui.Pack{
		Side:   ui.Left,
		Fill:   ui.FillY,
		Expand: true,
	})

	body := ui.NewFrame("Body Panel")
	body.Configure(ui.Config{
		Background:  render.RGBA(255, 0, 0, 64),
		BorderStyle: ui.BorderSunken,
		BorderSize:  2,
	})
	window.Pack(body, ui.Pack{
		Side:   ui.Left,
		Expand: true,
	})

	label1 := ui.NewLabel(ui.Label{
		Text: "Hello world!",
		Font: render.Text{
			Size:   24,
			Color:  render.Red,
			Stroke: render.Purple,
		},
	})
	body.Pack(label1, ui.Pack{
		Side: ui.Top,
	})

	window.Frame().SetBackground(render.Yellow)
	window.Compute(engine)
	// window.Present(engine, window.Point())

	super.Add(window)
	super.MainLoop(engine)
	//
	for true {

		start := time.Now()
		engine.Clear(render.White)

		// poll for events
		ev, err := engine.Poll()
		if err != nil {
			panic(err)
		}

		// escape key to close the window
		if ev.EscapeKey.Now {
			break
		}

		super.Loop(ev)
		window.Compute(engine)
		window.Present(engine, window.Point())
		engine.Present()

		// Delay to maintain the target frames per second.
		var delay uint32
		elapsed := time.Now().Sub(start)
		tmp := elapsed / time.Millisecond
		if TargetFPS-int(tmp) > 0 { // make sure it won't roll under
			delay = uint32(TargetFPS - int(tmp))
		}
		engine.Delay(delay)
	}
}
