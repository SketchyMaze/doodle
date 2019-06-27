package canvas

import (
	"syscall/js"

	"git.kirsle.net/apps/doodle/lib/events"
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
func (e *Engine) Poll() (*events.State, error) {
	s := e.events

	for event := e.PollEvent(); event != nil; event = e.PollEvent() {
		switch event.Class {
		case WindowEvent:
			s.Resized.Push(true)
		case MouseEvent:
			s.CursorX.Push(int32(event.X))
			s.CursorY.Push(int32(event.Y))
		case ClickEvent:
			s.CursorX.Push(int32(event.X))
			s.CursorY.Push(int32(event.Y))
			s.Button1.Push(event.LeftClick)
			s.Button2.Push(event.RightClick)
		case KeyEvent:
			switch event.KeyName {
			case "Escape":
				if event.Repeat {
					continue
				}

				if event.State {
					s.EscapeKey.Push(true)
				}
			case "Enter":
				if event.Repeat {
					continue
				}

				if event.State {
					s.EnterKey.Push(true)
				}
			case "F3":
				if event.State {
					s.KeyName.Push("F3")
				}
			case "ArrowUp":
				s.Up.Push(event.State)
			case "ArrowLeft":
				s.Left.Push(event.State)
			case "ArrowRight":
				s.Right.Push(event.State)
			case "ArrowDown":
				s.Down.Push(event.State)
			case "Shift":
				s.ShiftActive.Push(event.State)
				continue
			case "Alt":
			case "Control":
				continue
			case "Backspace":
				if event.State {
					s.KeyName.Push(`\b`)
				}
			default:
				if event.State {
					s.KeyName.Push(event.KeyName)
				} else {
					s.KeyName.Push("")
				}
			}
		}
	}

	return e.events, nil
}
