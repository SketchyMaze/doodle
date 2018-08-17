package ui

import (
	"sync"

	"git.kirsle.net/apps/doodle/events"
	"git.kirsle.net/apps/doodle/render"
)

// Event is a named event that the supervisor will send.
type Event int

// Events.
const (
	NullEvent Event = iota
	MouseOver
	MouseOut
	MouseDown
	MouseUp
	Click
	KeyDown
	KeyUp
	KeyPress
)

// Supervisor keeps track of widgets of interest to notify them about
// interaction events such as mouse hovers and clicks in their general
// vicinity.
type Supervisor struct {
	lock     sync.RWMutex
	widgets  []Widget
	hovering map[int]interface{}
	clicked  map[int]interface{}
}

// NewSupervisor creates a supervisor.
func NewSupervisor() *Supervisor {
	return &Supervisor{
		widgets:  []Widget{},
		hovering: map[int]interface{}{},
		clicked:  map[int]interface{}{},
	}
}

// Loop to check events and pass them to managed widgets.
func (s *Supervisor) Loop(ev *events.State) {
	var (
		XY = render.Point{
			X: ev.CursorX.Now,
			Y: ev.CursorY.Now,
		}
	)

	// See if we are hovering over any widgets.
	for id, w := range s.widgets {
		var (
			P  = w.Point()
			S  = w.Size()
			P2 = render.Point{
				X: P.X + S.W,
				Y: P.Y + S.H,
			}
		)

		if XY.X >= P.X && XY.X <= P2.X && XY.Y >= P.Y && XY.Y <= P2.Y {
			// Cursor has intersected the widget.
			if _, ok := s.hovering[id]; !ok {
				w.Event(MouseOver, XY)
				s.hovering[id] = nil
			}

			_, isClicked := s.clicked[id]
			if ev.Button1.Now {
				if !isClicked {
					w.Event(MouseDown, XY)
					s.clicked[id] = nil
				}
			} else if isClicked {
				w.Event(MouseUp, XY)
				w.Event(Click, XY)
				delete(s.clicked, id)
			}
		} else {
			// Cursor is not intersecting the widget.
			if _, ok := s.hovering[id]; ok {
				w.Event(MouseOut, XY)
				delete(s.hovering, id)
			}

			if _, ok := s.clicked[id]; ok {
				w.Event(MouseUp, XY)
				delete(s.clicked, id)
			}
		}
	}
}

// Present all widgets managed by the supervisor.
func (s *Supervisor) Present(e render.Engine) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	for _, w := range s.widgets {
		w.Present(e, w.Point())
	}
}

// Add a widget to be supervised.
func (s *Supervisor) Add(w Widget) {
	s.lock.Lock()
	s.widgets = append(s.widgets, w)
	s.lock.Unlock()
}
