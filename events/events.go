// Package events manages mouse and keyboard SDL events for Doodle.
package events

import (
	"errors"

	"github.com/veandco/go-sdl2/sdl"
)

// State keeps track of event states.
type State struct {
	// Mouse buttons.
	Button1 *BoolTick
	Button2 *BoolTick

	// Screenshot key.
	ScreenshotKey *BoolTick
	Up            *BoolTick
	Left          *BoolTick
	Right         *BoolTick
	Down          *BoolTick

	// Cursor positions.
	CursorX *Int32Tick
	CursorY *Int32Tick
}

// New creates a new event state manager.
func New() *State {
	return &State{
		Button1:       &BoolTick{},
		Button2:       &BoolTick{},
		ScreenshotKey: &BoolTick{},
		Up:            &BoolTick{},
		Left:          &BoolTick{},
		Right:         &BoolTick{},
		Down:          &BoolTick{},
		CursorX:       &Int32Tick{},
		CursorY:       &Int32Tick{},
	}
}

// Poll for events.
func (s *State) Poll(ticks uint64) (*State, error) {
	for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
		switch t := event.(type) {
		case *sdl.QuitEvent:
			return s, errors.New("quit")
		case *sdl.MouseMotionEvent:
			if DebugMouseEvents {
				log.Debug("[%d ms] tick:%d MouseMotion  type:%d  id:%d  x:%d  y:%d  xrel:%d  yrel:%d",
					t.Timestamp, ticks, t.Type, t.Which, t.X, t.Y, t.XRel, t.YRel,
				)
			}

			// Push the cursor position.
			s.CursorX.Push(t.X)
			s.CursorY.Push(t.Y)
			s.Button1.Push(t.State == 1)
		case *sdl.MouseButtonEvent:
			if DebugClickEvents {
				log.Debug("[%d ms] tick:%d MouseButton  type:%d  id:%d  x:%d  y:%d  button:%d  state:%d",
					t.Timestamp, ticks, t.Type, t.Which, t.X, t.Y, t.Button, t.State,
				)
			}

			// Push the cursor position.
			s.CursorX.Push(t.X)
			s.CursorY.Push(t.Y)

			// Is a mouse button pressed down?
			if t.Button == 1 {
				var eventName string
				if DebugClickEvents {
					if t.State == 1 && s.Button1.Now == false {
						eventName = "DOWN"
					} else if t.State == 0 && s.Button1.Now == true {
						eventName = "UP"
					}
				}

				if eventName != "" {
					log.Debug("tick:%d  Mouse Button1 %s BEFORE: %+v",
						ticks,
						eventName,
						s.Button1,
					)
					s.Button1.Push(eventName == "DOWN")
					log.Debug("tick:%d  Mouse Button1 %s AFTER: %+v",
						ticks,
						eventName,
						s.Button1,
					)

					// Return the event immediately.
					return s, nil
				}
			}

			// s.Button2.Push(t.Button == 3 && t.State == 1)
		case *sdl.MouseWheelEvent:
			if DebugMouseEvents {
				log.Debug("[%d ms] tick:%d MouseWheel  type:%d  id:%d  x:%d  y:%d",
					t.Timestamp, ticks, t.Type, t.Which, t.X, t.Y,
				)
			}
		case *sdl.KeyboardEvent:
			log.Debug("[%d ms] tick:%d Keyboard  type:%d  sym:%c  modifiers:%d  state:%d  repeat:%d\n",
				t.Timestamp, ticks, t.Type, t.Keysym.Sym, t.Keysym.Mod, t.State, t.Repeat,
			)
			if t.Repeat == 1 {
				continue
			}

			switch t.Keysym.Scancode {
			case sdl.SCANCODE_F12:
				s.ScreenshotKey.Push(t.State == 1)
			case sdl.SCANCODE_UP:
				s.Up.Push(t.State == 1)
			case sdl.SCANCODE_LEFT:
				s.Left.Push(t.State == 1)
			case sdl.SCANCODE_RIGHT:
				s.Right.Push(t.State == 1)
			case sdl.SCANCODE_DOWN:
				s.Down.Push(t.State == 1)
			}
		}
	}

	return s, nil
}
