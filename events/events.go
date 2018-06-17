// Package events manages mouse and keyboard SDL events for Doodle.
package events

import (
	"errors"

	"github.com/veandco/go-sdl2/sdl"
)

// State keeps track of event states.
type State struct {
	// Mouse buttons.
	Button1 BoolFrameState
	Button2 BoolFrameState

	// Cursor positions.
	CursorX Int32FrameState
	CursorY Int32FrameState
}

// New creates a new event state manager.
func New() *State {
	return &State{}
}

// Poll for events.
func (s *State) Poll() (*State, error) {
	for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
		switch t := event.(type) {
		case *sdl.QuitEvent:
			return s, errors.New("quit")
		case *sdl.MouseMotionEvent:
			if DebugMouseEvents {
				log.Debug("[%d ms] MouseMotion  type:%d  id:%d  x:%d  y:%d  xrel:%d  yrel:%d",
					t.Timestamp, t.Type, t.Which, t.X, t.Y, t.XRel, t.YRel,
				)
			}

			// Push the cursor position.
			s.CursorX.Push(t.X)
			s.CursorY.Push(t.Y)
			s.Button1.Push(t.State == 1)
		case *sdl.MouseButtonEvent:
			if DebugClickEvents {
				log.Debug("[%d ms] MouseButton  type:%d  id:%d  x:%d  y:%d  button:%d  state:%d",
					t.Timestamp, t.Type, t.Which, t.X, t.Y, t.Button, t.State,
				)
			}

			// Push the cursor position.
			s.CursorX.Push(t.X)
			s.CursorY.Push(t.Y)

			// Is a mouse button pressed down?
			if t.Button == 1 {
				if DebugClickEvents {
					if t.State == 1 && s.Button1.Now == false {
						log.Debug("Mouse Button1 DOWN")
					} else if t.State == 0 && s.Button1.Now == true {
						log.Debug("Mouse Button1 UP")
					}
				}
				s.Button1.Push(t.State == 1)
			}

			s.Button2.Push(t.Button == 3 && t.State == 1)
		case *sdl.MouseWheelEvent:
			if DebugMouseEvents {
				log.Debug("[%d ms] MouseWheel  type:%d  id:%d  x:%d  y:%d",
					t.Timestamp, t.Type, t.Which, t.X, t.Y,
				)
			}
		case *sdl.KeyboardEvent:
			log.Debug("[%d ms] Keyboard  type:%d  sym:%c  modifiers:%d  state:%d  repeat:%d\n",
				t.Timestamp, t.Type, t.Keysym.Sym, t.Keysym.Mod, t.State, t.Repeat,
			)
		}
	}

	return s, nil
}
