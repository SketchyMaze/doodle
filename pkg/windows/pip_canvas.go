package windows

import (
	"git.kirsle.net/apps/doodle/pkg/balance"
	"git.kirsle.net/apps/doodle/pkg/drawtool"
	"git.kirsle.net/apps/doodle/pkg/level"
	"git.kirsle.net/apps/doodle/pkg/shmem"
	"git.kirsle.net/apps/doodle/pkg/uix"
	"git.kirsle.net/go/render"
	"git.kirsle.net/go/render/event"
	"git.kirsle.net/go/ui"
)

// PiP window.
type PiP struct {
	// Settings passed in by doodle
	Supervisor *ui.Supervisor
	Engine     render.Engine
	Level      *level.Level
	Event      *event.State
	Tool       *drawtool.Tool
	BrushSize  *int

	// Or sensible defaults:
	Width  int
	Height int

	OnCancel func()
}

// MakePiPWindow initializes a license window for any scene.
// The window width/height are the actual SDL2 window dimensions.
func MakePiPWindow(windowWidth, windowHeight int, cfg PiP) *ui.Window {
	win := NewPiPWindow(cfg)
	win.Compute(cfg.Engine)
	win.Supervise(cfg.Supervisor)

	// Center the window.
	size := win.Size()
	win.MoveTo(render.Point{
		X: (windowWidth / 2) - (size.W / 2),
		Y: (windowHeight / 2) - (size.H / 2),
	})

	return win
}

// NewPiPWindow initializes the window.
func NewPiPWindow(cfg PiP) *ui.Window {
	var (
		windowWidth  = 340
		windowHeight = 300
	)

	if cfg.Width+cfg.Height > 0 {
		windowWidth = cfg.Width
		windowHeight = cfg.Height
	}

	var (
		canvasWidth  = windowWidth - 8
		canvasHeight = windowHeight - 4 - 48 // for the titlebar?
	)

	window := ui.NewWindow("Viewport")
	window.SetButtons(ui.CloseButton)
	window.Configure(ui.Config{
		Width:  windowWidth,
		Height: windowHeight,
	})

	canvas := uix.NewCanvas(128, true)
	canvas.Name = "Viewport (WIP)"
	canvas.LoadLevel(cfg.Level)
	canvas.InstallActors(cfg.Level.Actors)
	canvas.Scrollable = true
	canvas.Editable = true
	canvas.Resize(render.NewRect(canvasWidth, canvasHeight))

	// NOTE: my UI toolkit calls this every tick, if this is "fixed"
	// in the future make one that does.
	var (
		curTool  = *cfg.Tool
		curThicc = *cfg.BrushSize
	)
	canvas.Tool = curTool
	window.Handle(ui.MouseMove, func(ed ui.EventData) error {
		canvas.Loop(cfg.Event)

		// Check if bound values have modified.
		if *cfg.Tool != curTool {
			curTool = *cfg.Tool
			canvas.Tool = curTool
		}
		if *cfg.BrushSize != curThicc {
			curThicc = *cfg.BrushSize
			canvas.BrushSize = curThicc
		}
		return nil
	})

	window.Pack(canvas, ui.Pack{
		Side:  ui.N,
		FillX: true,
	})

	/////////////
	// Buttons at bottom of window

	bottomFrame := ui.NewFrame("Button Frame")
	window.Pack(bottomFrame, ui.Pack{
		Side:  ui.N,
		FillX: true,
		PadY:  4,
	})

	frame := ui.NewFrame("Button frame")
	buttons := []struct {
		label   string
		tooltip string
		f       func()
	}{
		{"Smaller", "Shrink this viewport window by 20%", func() {
			// Make a smaller version of the same window, and close.
			cfg.Width = int(float64(windowWidth) * 0.8)
			cfg.Height = int(float64(windowHeight) * 0.8)
			pip := MakePiPWindow(cfg.Width, cfg.Height, cfg)
			pip.MoveTo(window.Point())
			window.Close()
			pip.Show()
		}},
		{"Larger", "Grow this viewport window by 20%", func() {
			// Make a smaller version of the same window, and close.
			cfg.Width = int(float64(windowWidth) * 1.2)
			cfg.Height = int(float64(windowHeight) * 1.2)
			pip := MakePiPWindow(cfg.Width, cfg.Height, cfg)
			pip.MoveTo(window.Point())
			window.Close()
			pip.Show()
		}},
		{"Refresh", "Update the state of doodads placed in this level", func() {
			canvas.ClearActors()
			canvas.InstallActors(cfg.Level.Actors)
		}},
		{"Rename", "Give this viewport window a custom name", func() {
			shmem.Prompt("Give this viewport a name: ", func(answer string) {
				if answer == "" {
					return
				}
				window.Title = answer
			})
		}},
		{"???", "Load a different drawing (experimental!)", func() {
			shmem.Prompt("Filename to open: ", func(answer string) {
				if answer == "" {
					return
				}

				if lvl, err := level.LoadFile(answer); err == nil {
					canvas.ClearActors()
					canvas.LoadLevel(lvl)
					canvas.InstallActors(lvl.Actors)
				} else {
					shmem.Flash(err.Error())
				}
			})
		}},
	}
	for i, button := range buttons {
		// Start axing buttons if window size is too small.
		if windowWidth < 150 && i > 1 {
			break
		} else if windowWidth < 250 && i > 2 {
			break
		}

		button := button

		btn := ui.NewButton(button.label, ui.NewLabel(ui.Label{
			Text: button.label,
			Font: balance.SmallFont,
		}))

		btn.Handle(ui.Click, func(ed ui.EventData) error {
			button.f()
			return nil
		})

		btn.Compute(cfg.Engine)
		cfg.Supervisor.Add(btn)

		ui.NewTooltip(btn, ui.Tooltip{
			Text: button.tooltip,
			Edge: ui.Top,
		})

		frame.Pack(btn, ui.Pack{
			Side:   ui.W,
			PadX:   1,
			Expand: true,
			Fill:   true,
		})
	}

	bottomFrame.Pack(frame, ui.Pack{
		Side: ui.N,
		PadX: 8,
		PadY: 0,
	})

	return window
}
