package sdl

import (
	"errors"
	"fmt"

	"git.kirsle.net/apps/doodle/lib/events"
	"github.com/veandco/go-sdl2/sdl"
)

// Debug certain SDL events
var (
	DebugWindowEvents = false
	DebugMouseEvents  = false
	DebugClickEvents  = false
	DebugKeyEvents    = false
)

// Poll for events.
func (r *Renderer) Poll() (*events.State, error) {
	s := r.events

	// helper function to push keyboard key names on keyDown events only.
	pushKey := func(name string, state uint8) {
		if state == 1 {
			s.KeyName.Push(name)
		}
	}

	for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
		switch t := event.(type) {
		case *sdl.QuitEvent:
			return s, errors.New("quit")
		case *sdl.WindowEvent:
			if DebugWindowEvents {
				if t.Event == sdl.WINDOWEVENT_RESIZED {
					fmt.Printf("[%d ms] tick:%d Window Resized to %dx%d",
						t.Timestamp,
						r.ticks,
						t.Data1,
						t.Data2,
					)
				}
			}
			s.Resized.Push(true)
		case *sdl.MouseMotionEvent:
			if DebugMouseEvents {
				fmt.Printf("[%d ms] tick:%d MouseMotion  type:%d  id:%d  x:%d  y:%d  xrel:%d  yrel:%d",
					t.Timestamp, r.ticks, t.Type, t.Which, t.X, t.Y, t.XRel, t.YRel,
				)
			}

			// Push the cursor position.
			s.CursorX.Push(t.X)
			s.CursorY.Push(t.Y)
			s.Button1.Push(t.State == 1)
		case *sdl.MouseButtonEvent:
			if DebugClickEvents {
				fmt.Printf("[%d ms] tick:%d MouseButton  type:%d  id:%d  x:%d  y:%d  button:%d  state:%d",
					t.Timestamp, r.ticks, t.Type, t.Which, t.X, t.Y, t.Button, t.State,
				)
			}

			// Push the cursor position.
			s.CursorX.Push(t.X)
			s.CursorY.Push(t.Y)

			// Is a mouse button pressed down?
			checkDown := func(number uint8, target *events.BoolTick) bool {
				if t.Button == number {
					var eventName string
					if t.State == 1 && target.Now == false {
						eventName = "DOWN"
					} else if t.State == 0 && target.Now == true {
						eventName = "UP"
					}

					if eventName != "" {
						target.Push(eventName == "DOWN")
					}
					return true
				}
				return false
			}

			if checkDown(1, s.Button1) || checkDown(3, s.Button2) || checkDown(2, s.Button3) {
				// Return the event immediately.
				return s, nil
			}
		case *sdl.MouseWheelEvent:
			if DebugMouseEvents {
				fmt.Printf("[%d ms] tick:%d MouseWheel  type:%d  id:%d  x:%d  y:%d",
					t.Timestamp, r.ticks, t.Type, t.Which, t.X, t.Y,
				)
			}
		case *sdl.KeyboardEvent:
			if DebugKeyEvents {
				fmt.Printf("[%d ms] tick:%d Keyboard  type:%d  sym:%c  modifiers:%d  state:%d  repeat:%d\n",
					t.Timestamp, r.ticks, t.Type, t.Keysym.Sym, t.Keysym.Mod, t.State, t.Repeat,
				)
			}

			switch t.Keysym.Scancode {
			case sdl.SCANCODE_ESCAPE:
				if t.Repeat == 1 {
					continue
				}
				s.EscapeKey.Push(t.State == 1)
			case sdl.SCANCODE_RETURN:
				if t.Repeat == 1 {
					continue
				}
				s.EnterKey.Push(t.State == 1)
			case sdl.SCANCODE_F1:
				pushKey("F1", t.State)
			case sdl.SCANCODE_F2:
				pushKey("F2", t.State)
			case sdl.SCANCODE_F3:
				pushKey("F3", t.State)
			case sdl.SCANCODE_F4:
				pushKey("F4", t.State)
			case sdl.SCANCODE_F5:
				pushKey("F5", t.State)
			case sdl.SCANCODE_F6:
				pushKey("F6", t.State)
			case sdl.SCANCODE_F7:
				pushKey("F7", t.State)
			case sdl.SCANCODE_F8:
				pushKey("F8", t.State)
			case sdl.SCANCODE_F9:
				pushKey("F9", t.State)
			case sdl.SCANCODE_F10:
				pushKey("F10", t.State)
			case sdl.SCANCODE_F11:
				pushKey("F11", t.State)
			case sdl.SCANCODE_F12:
				pushKey("F12", t.State)
			case sdl.SCANCODE_UP:
				s.Up.Push(t.State == 1)
			case sdl.SCANCODE_LEFT:
				s.Left.Push(t.State == 1)
			case sdl.SCANCODE_RIGHT:
				s.Right.Push(t.State == 1)
			case sdl.SCANCODE_DOWN:
				s.Down.Push(t.State == 1)
			case sdl.SCANCODE_LSHIFT:
			case sdl.SCANCODE_RSHIFT:
				s.ShiftActive.Push(t.State == 1)
			case sdl.SCANCODE_LALT:
			case sdl.SCANCODE_RALT:
				continue
			case sdl.SCANCODE_LCTRL:
				s.ControlActive.Push(t.State == 1)
			case sdl.SCANCODE_RCTRL:
				s.ControlActive.Push(t.State == 1)
			case sdl.SCANCODE_BACKSPACE:
				// Make it a key event with "\b" as the sequence.
				if t.State == 1 || t.Repeat == 1 {
					s.KeyName.Push(`\b`)
				}
			default:
				// Push the string value of the key.
				if t.State == 1 {
					s.KeyName.Push(string(t.Keysym.Sym))
				}
			}
		}
	}

	return s, nil
}
