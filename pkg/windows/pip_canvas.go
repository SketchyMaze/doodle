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
		windowHeight = 320
	)

	window := ui.NewWindow("Viewport (WORK IN PROGRESS!)")
	window.SetButtons(ui.CloseButton)
	window.Configure(ui.Config{
		Width:      windowWidth,
		Height:     windowHeight,
		Background: render.RGBA(255, 200, 255, 255),
	})

	canvas := uix.NewCanvas(128, true)
	canvas.Name = "Viewport"
	canvas.LoadLevel(cfg.Level)
	canvas.InstallActors(cfg.Level.Actors)
	canvas.Scrollable = true
	canvas.Editable = true
	canvas.Resize(render.NewRect(windowWidth, windowHeight))

	// NOTE: my UI toolkit calls this every tick, if this is "fixed"
	// in the future make one that does.
	window.Handle(ui.MouseMove, func(ed ui.EventData) error {
		canvas.Loop(cfg.Event)
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
	})

	frame := ui.NewFrame("Button frame")
	buttons := []struct {
		label   string
		tooltip string
		down    func()
		f       func()
	}{
		{"^", "Scroll up", func() {
			canvas.ScrollBy(render.NewPoint(0, 64))
		}, nil},
		{"v", "Scroll down", func() {
			canvas.ScrollBy(render.NewPoint(0, -64))
		}, nil},
		{"<", "Scroll left", func() {
			canvas.ScrollBy(render.NewPoint(64, 0))
		}, nil},
		{">", "Scroll right", func() {
			canvas.ScrollBy(render.NewPoint(-64, 0))
		}, nil},
		{"0", "Reset to origin", nil, func() {
			canvas.ScrollTo(render.Origin)
		}},
		{"???", "Load a different drawing", nil, func() {
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
	for _, button := range buttons {
		button := button

		btn := ui.NewButton(button.label, ui.NewLabel(ui.Label{
			Text: button.label,
			Font: balance.MenuFont,
		}))

		if button.down != nil {
			btn.Handle(ui.MouseDown, func(ed ui.EventData) error {
				button.down()
				return nil
			})
		}

		if button.f != nil {
			btn.Handle(ui.Click, func(ed ui.EventData) error {
				button.f()
				return nil
			})
		}

		btn.Compute(cfg.Engine)
		cfg.Supervisor.Add(btn)

		ui.NewTooltip(btn, ui.Tooltip{
			Text: button.tooltip,
			Edge: ui.Top,
		})

		frame.Pack(btn, ui.Pack{
			Side:   ui.W,
			PadX:   4,
			Expand: true,
			Fill:   true,
		})
	}

	// Tool selector.
	toolBtn := ui.NewSelectBox("Tool Select", ui.Label{
		Font: ui.MenuFont,
	})
	toolBtn.AlwaysChange = true
	frame.Pack(toolBtn, ui.Pack{
		Side:   ui.W,
		Expand: true,
	})

	toolBtn.AddItem("Pencil", drawtool.PencilTool, func() {})
	toolBtn.AddItem("Line", drawtool.LineTool, func() {})
	toolBtn.AddItem("Rectangle", drawtool.RectTool, func() {})
	toolBtn.AddItem("Ellipse", drawtool.EllipseTool, func() {})

	// TODO: Actor and Link Tools don't work as the canvas needs
	// hooks for their events. The code in EditorUI#SetupCanvas should
	// be made reusable here.
	// toolBtn.AddItem("Link", drawtool.LinkTool, func() {})
	// toolBtn.AddItem("Actor", drawtool.ActorTool, func() {})

	toolBtn.Handle(ui.Change, func(ed ui.EventData) error {
		selection, _ := toolBtn.GetValue()
		tool, _ := selection.Value.(drawtool.Tool)

		// log.Error("Change: %d, b4: %s", value, canvas.Tool)
		canvas.Tool = tool

		return nil
	})

	ui.NewTooltip(toolBtn, ui.Tooltip{
		Text: "Draw tool (viewport only)",
		Edge: ui.Top,
	})

	toolBtn.Supervise(cfg.Supervisor)
	cfg.Supervisor.Add(toolBtn)

	bottomFrame.Pack(frame, ui.Pack{
		Side: ui.N,
		PadX: 8,
		PadY: 12,
	})

	return window
}
