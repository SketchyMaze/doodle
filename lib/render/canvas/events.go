package canvas

import (
	"syscall/js"

	"git.kirsle.net/apps/doodle/lib/render/event"
)

// EventClass to categorize JavaScript events.
type EventClass int

// EventClass values.
const (
	MouseEvent EventClass = iota
	ClickEvent
	KeyEvent
	ResizeEvent
	WindowEvent
)

// Event object queues up asynchronous JavaScript events to be processed linearly.
type Event struct {
	Name  string // mouseup, keydown, etc.
	Class EventClass

	// Mouse events.
	X          int
	Y          int
	LeftClick  bool
	RightClick bool

	// Key events.
	KeyName string
	State   bool
	Repeat  bool
}

// AddEventListeners sets up bindings to collect events from the browser.
func (e *Engine) AddEventListeners() {
	// Window resize.
	js.Global().Get("window").Call(
		"addEventListener",
		"resize",
		js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			e.queue <- Event{
				Name:  "resize",
				Class: WindowEvent,
			}
			return nil
		}),
	)

	// Mouse movement.
	e.canvas.Value.Call(
		"addEventListener",
		"mousemove",
		js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			var (
				x = args[0].Get("pageX").Int()
				y = args[0].Get("pageY").Int()
			)

			e.queue <- Event{
				Name:  "mousemove",
				Class: MouseEvent,
				X:     x,
				Y:     y,
			}
			return nil
		}),
	)

	// Mouse clicks.
	for _, ev := range []string{"mouseup", "mousedown"} {
		ev := ev
		e.canvas.Value.Call(
			"addEventListener",
			ev,
			js.FuncOf(func(this js.Value, args []js.Value) interface{} {
				var (
					x     = args[0].Get("pageX").Int()
					y     = args[0].Get("pageY").Int()
					which = args[0].Get("which").Int()
				)

				// Is a mouse button pressed down?
				checkDown := func(number int) bool {
					if which == number {
						return ev == "mousedown"
					}
					return false
				}

				e.queue <- Event{
					Name:       ev,
					Class:      ClickEvent,
					X:          x,
					Y:          y,
					LeftClick:  checkDown(1),
					RightClick: checkDown(3),
				}
				return false
			}),
		)
	}

	// Supress context menu.
	e.canvas.Value.Call(
		"addEventListener",
		"contextmenu",
		js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			args[0].Call("preventDefault")
			return false
		}),
	)

	// Keyboard keys
	for _, ev := range []string{"keydown", "keyup"} {
		ev := ev
		js.Global().Get("document").Call(
			"addEventListener",
			ev,
			js.FuncOf(func(this js.Value, args []js.Value) interface{} {
				var (
					event  = args[0]
					key    = event.Get("key").String()
					repeat = event.Get("repeat").Bool()

					pressed = ev == "keydown"
				)

				if key == "F3" {
					args[0].Call("preventDefault")
				}

				e.queue <- Event{
					Name:    ev,
					Class:   KeyEvent,
					KeyName: key,
					Repeat:  repeat,
					State:   pressed,
				}
				return nil
			}),
		)
	}
}

// PollEvent returns the next event in the queue, or null.
func (e *Engine) PollEvent() *Event {
	select {
	case event := <-e.queue:
		return &event
	default:
		return nil
	}
	return nil
}

// Poll for events.
func (e *Engine) Poll() (*event.State, error) {
	s := e.events

	for event := e.PollEvent(); event != nil; event = e.PollEvent() {
		switch event.Class {
		case WindowEvent:
			s.WindowResized = true
		case MouseEvent:
			s.CursorX = event.X
			s.CursorY = event.Y
		case ClickEvent:
			s.CursorX = event.X
			s.CursorY = event.Y
			s.Button1 = event.LeftClick
			s.Button2 = event.RightClick
		case KeyEvent:
			switch event.KeyName {
			case "Escape":
				if event.Repeat {
					continue
				}

				s.Escape = event.State
			case "Enter":
				if event.Repeat {
					continue
				}

				s.Enter = event.State
			case "F3":
				s.SetKeyDown("F3", event.State)
			case "ArrowUp":
				s.Up = event.State
			case "ArrowLeft":
				s.Left = event.State
			case "ArrowRight":
				s.Right = event.State
			case "ArrowDown":
				s.Down = event.State
			case "Shift":
				s.Shift = event.State
				continue
			case "Alt":
				s.Alt = event.State
			case "Control":
				s.Ctrl = event.State
			case "Backspace":
				s.SetKeyDown(`\b`, event.State)
			default:
				s.SetKeyDown(event.KeyName, event.State)
			}
		}
	}

	return e.events, nil
}
