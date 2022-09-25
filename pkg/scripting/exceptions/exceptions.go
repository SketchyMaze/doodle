// Package exceptions handles JavaScript errors nicely for the game.
package exceptions

import (
	"fmt"
	"strings"
	"sync"

	"git.kirsle.net/SketchyMaze/doodle/pkg/balance"
	"git.kirsle.net/SketchyMaze/doodle/pkg/log"
	"git.kirsle.net/SketchyMaze/doodle/pkg/native"
	"git.kirsle.net/SketchyMaze/doodle/pkg/shmem"
	"git.kirsle.net/go/render"
	"git.kirsle.net/go/render/event"
	"git.kirsle.net/go/ui"
)

// The exception catching window is a singleton and appears on top with
// its own supervisor apart from whatever the game is currently doing.
var (
	Supervisor    *ui.Supervisor
	Window        *ui.Window
	Disabled      bool         // don't reopen the window again
	lastException string       // text of last exception
	excLabel      *string      // trimmed exception label text
	mu            sync.RWMutex // thread safety

	// Configurables.
	winSize   = render.NewRect(380, 260)
	excSize   = render.NewRect(360, 140)
	charsWide = 50 // to trim the exception text onto screen
	charsTall = 9
)

func init() {
	l := "No Exception Traceback"
	excLabel = &l
}

// Catch a JavaScript exception and maybe show it to the user.
func Catch(exc string, args ...interface{}) {
	if len(args) > 0 {
		exc = fmt.Sprintf(exc, args...)
	}

	log.Error("[JS] Exception: %s", exc)
	if Disabled {
		return
	}

	Setup()

	width, _ := shmem.CurrentRenderEngine.WindowSize()
	Window.MoveTo(render.Point{
		X: (width / 2) - (Window.Size().W / 2),
		Y: 60,
	})

	Window.Show()
	lastException = exc
	*excLabel = trim(exc)
}

// Setup the global supervisor and window the first time - after the render engine has initialized,
// e.g., when you want the window to show up the first time.
func Setup() {
	mu.Lock()
	defer mu.Unlock()

	log.Info("Setup Exceptions")

	if Supervisor == nil {
		Supervisor = ui.NewSupervisor()
	}

	if Window == nil {
		Window = MakeWindow(Exception{
			Supervisor: Supervisor,
			Engine:     shmem.CurrentRenderEngine,
		})
		Window.Compute(shmem.CurrentRenderEngine)
		Window.Supervise(Supervisor)
	}
}

// Teardown the exception UI, if it was loaded. This is done between Scene transitions in the game
// to reduce memory leaks in case the scenes will flush SDL2 caches or w/e.
func Teardown() {
	log.Error("Teardown Exceptions")
	mu.Lock()
	defer mu.Unlock()

	if Window != nil {
		Window = nil
	}

	if Supervisor != nil {
		Supervisor = nil
	}
}

// Handled returns true if the exception window handles the events this tick, e.g.
// so clicking on its window won't let your mouse click also hit things behind it.
func Handled(ev *event.State) bool {
	if Window != nil && !Window.Hidden() {
		// If they hit the Return/Escape key, dismiss the exception.
		if ev.Enter || ev.Escape {
			ev.Escape = false
			Window.Close()
			return true
		}

		Supervisor.Loop(ev)

		// NOTE: if the window Close handler tears down the Supervisor, the
		// previous call to Supervisor.Loop() may run it and then Supervisor
		// no longer exists.
		if Supervisor != nil && Supervisor.IsPointInWindow(shmem.Cursor) {
			return true
		}
	}
	return false
}

// Loop allows the exception window to appear on game tick.
func Draw(e render.Engine) {
	if Supervisor != nil {
		Supervisor.Present(e)
	}
}

// Exception window to show scripting errors in doodads.
type Exception struct {
	// Settings passed in by doodle
	Supervisor *ui.Supervisor
	Engine     render.Engine
}

// Function to trim the raw exception text so it fits neatly within the label.
// In case it's long, use the Copy button to copy to your clipboard.
func trim(input string) string {
	var lines = []string{}
	for _, line := range strings.Split(input, "\n") {
		if len(lines) >= charsTall {
			lines = lines[:charsTall]
			break
		}

		if len(line) > charsWide {
			// Word wrap it.
			for len(line) > charsWide {
				lines = append(lines, line[:charsWide])
				line = line[charsWide:]
			}
			if len(line) > 0 {
				lines = append(lines, line)
			}
			continue
		}

		lines = append(lines, line)
	}
	return strings.Join(lines, "\n")
}

// MakeWindow initializes the window.
func MakeWindow(cfg Exception) *ui.Window {
	window := ui.NewWindow("Exception")
	window.Configure(ui.Config{
		Width:      winSize.W,
		Height:     winSize.H,
		Background: render.Red.Lighten(128),
	})

	window.Handle(ui.CloseWindow, func(ed ui.EventData) error {
		Teardown()
		return ui.ErrStopPropagation
	})

	header := ui.NewLabel(ui.Label{
		Text: "A JavaScript error has occurred in a doodad:",
		Font: balance.UIFont,
	})
	window.Pack(header, ui.Pack{
		Side:    ui.N,
		Padding: 8,
	})

	text := ui.NewLabel(ui.Label{
		TextVariable: excLabel,
		Font:         balance.ExceptionFont,
	})
	text.Configure(ui.Config{
		BorderSize:  1,
		BorderStyle: ui.BorderSunken,
		Background:  render.White,
		Width:       excSize.W,
		Height:      excSize.H,
	})
	window.Pack(text, ui.Pack{
		Side: ui.N,
	})

	frame := ui.NewFrame("Button frame")
	buttons := []struct {
		label   string
		tooltip string
		f       func()
	}{
		{"Dismiss", "", func() {
			Window.Close()
		}},
		{"Copy", "Copy the full text to clipboard", func() {
			native.CopyToClipboard(lastException)
		}},
		{"Don't show again", "Don't show errors like this again\nuntil your next play session", func() {
			Disabled = true
			Window.Close()
		}},
	}
	for i, button := range buttons {
		button := button

		btn := ui.NewButton(button.label, ui.NewLabel(ui.Label{
			Text: button.label,
			Font: balance.MenuFont,
		}))
		if i == 0 {
			btn.SetStyle(&balance.ButtonPrimary)
		}

		btn.Handle(ui.Click, func(ed ui.EventData) error {
			button.f()
			return nil
		})

		btn.Compute(cfg.Engine)

		// Tooltips?
		if len(button.tooltip) > 0 {
			ui.NewTooltip(btn, ui.Tooltip{
				Text: button.tooltip,
				Edge: ui.Bottom,
			})
		}

		cfg.Supervisor.Add(btn)

		frame.Pack(btn, ui.Pack{
			Side:   ui.W,
			PadX:   4,
			Expand: true,
			Fill:   true,
		})
	}
	window.Pack(frame, ui.Pack{
		Side: ui.N,
		PadY: 12,
	})

	return window
}
