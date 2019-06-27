package canvas

import (
	"syscall/js"

	"git.kirsle.net/apps/doodle/lib/events"
	"git.kirsle.net/apps/doodle/pkg/log"
)

// AddEventListeners sets up bindings to collect events from the browser.
func (e *Engine) AddEventListeners() {
	s := e.events

	// Mouse movement.
	e.canvas.Value.Call(
		"addEventListener",
		"mousemove",
		js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			var (
				x = args[0].Get("pageX").Int()
				y = args[0].Get("pageY").Int()
			)

			s.CursorX.Push(int32(x))
			s.CursorY.Push(int32(y))
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

				log.Info("Clicked at %d,%d", x, y)

				s.CursorX.Push(int32(x))
				s.CursorY.Push(int32(y))

				// Is a mouse button pressed down?
				checkDown := func(number int) bool {
					if which == number {
						return ev == "mousedown"
					}
					return false
				}

				s.Button1.Push(checkDown(1))
				s.Button2.Push(checkDown(3))
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
	// js.Global().Get("document").Call(
	// 	"addEventListener",
	// 	"keydown",
	// 	js.FuncOf(func(this js.Value, args []js.Value) interface{} {
	// 		log.Info("key: %+v", args)
	// 		var (
	// 			event    = args[0]
	// 			charCode = event.Get("charCode")
	// 			key      = event.Get("key").String()
	// 		)
	//
	// 		switch key {
	// 		case "Enter":
	// 			s.EnterKey.Push(true)
	// 			// default:
	// 			// 	s.KeyName.Push(key)
	// 		}
	//
	// 		log.Info("keypress: code=%s  key=%s", charCode, key)
	//
	// 		return nil
	// 	}),
	// )
}

// Poll for events.
func (e *Engine) Poll() (*events.State, error) {
	return e.events, nil
}
