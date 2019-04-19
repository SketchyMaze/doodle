package ui

import (
	"errors"
	"sync"
	"time"

	"git.kirsle.net/apps/doodle/lib/events"
	"git.kirsle.net/apps/doodle/lib/render"
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
	Drop
)

// Supervisor keeps track of widgets of interest to notify them about
// interaction events such as mouse hovers and clicks in their general
// vicinity.
type Supervisor struct {
	lock      sync.RWMutex
	targetFPS int
	serial    int                 // ID number of each widget added in order
	widgets   map[int]WidgetSlot  // map of widget ID to WidgetSlot
	hovering  map[int]interface{} // map of widgets under the cursor
	clicked   map[int]interface{} // map of widgets being clicked
	dd        *DragDrop
}

// WidgetSlot holds a widget with a unique ID number in a sorted list.
type WidgetSlot struct {
	id     int
	widget Widget
}

// NewSupervisor creates a supervisor.
func NewSupervisor() *Supervisor {
	return &Supervisor{
		targetFPS: 1000 / 60,
		widgets:   map[int]WidgetSlot{},
		hovering:  map[int]interface{}{},
		clicked:   map[int]interface{}{},
		dd:        NewDragDrop(),
	}
}

// DragStart sets the drag state.
func (s *Supervisor) DragStart() {
	s.dd.Start()
}

// DragStop stops the drag state.
func (s *Supervisor) DragStop() {
	s.dd.Stop()
}

// IsDragging returns whether the drag state is enabled.
func (s *Supervisor) IsDragging() bool {
	return s.dd.IsDragging()
}

// Error messages that may be returned by Supervisor.Loop()
var (
	// The caller should STOP forwarding any mouse or keyboard events to any
	// other handles for the remainder of this tick.
	ErrStopPropagation = errors.New("stop all event propagation")
)

// MainLoop starts the UI main loop, for UI-only applications.
func (s *Supervisor) MainLoop(e render.Engine) error {
	for true {
		start := time.Now()
		e.Clear(render.Green)

		// Poll for events.
		ev, err := e.Poll()
		if err != nil {
			return err
		}

		// TODO: escape key to exit the main loop
		if ev.EscapeKey.Now {
			return nil
		}

		s.Loop(ev)

		// Render the widgets under our care.
		s.Present(e)

		// Commit the pixels to screen.
		e.Present()

		// Delay to maintain the target FPS.
		var delay uint32
		elapsed := time.Now().Sub(start)
		tmp := elapsed / time.Millisecond
		if s.targetFPS-int(tmp) > 0 {
			delay = uint32(s.targetFPS - int(tmp))
		}
		e.Delay(delay)
	}
	return nil
}

// Loop to check events and pass them to managed widgets.
//
// Useful errors returned by this may be:
// - ErrStopPropagation
func (s *Supervisor) Loop(ev *events.State) error {
	var (
		XY = render.Point{
			X: ev.CursorX.Now,
			Y: ev.CursorY.Now,
		}
	)

	// See if we are hovering over any widgets.
	hovering, outside := s.Hovering(XY)

	// If we are dragging something around, do not trigger any mouse events
	// to other widgets but DO notify any widget we dropped on top of!
	if s.dd.IsDragging() {
		if !ev.Button1.Now && !ev.Button2.Now {
			// The mouse has been released. TODO: make mouse button important?
			for _, child := range hovering {
				child.widget.Event(Drop, XY)
			}
			s.DragStop()
		}
		return ErrStopPropagation
	}

	for _, child := range hovering {
		var (
			id = child.id
			w  = child.widget
		)
		if w.Hidden() {
			// TODO: somehow the Supervisor wasn't triggering hidden widgets
			// anyway, but I don't know why. Adding this check for safety.
			continue
		}

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
	}
	for _, child := range outside {
		var (
			id = child.id
			w  = child.widget
		)

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

	return nil
}

// Hovering returns all of the widgets managed by Supervisor that are under
// the mouse cursor. Returns the set of widgets below the cursor and the set
// of widgets not below the cursor.
func (s *Supervisor) Hovering(cursor render.Point) (hovering, outside []WidgetSlot) {
	var XY = cursor // for shorthand
	hovering = []WidgetSlot{}
	outside = []WidgetSlot{}

	// Check all the widgets under our care.
	for child := range s.Widgets() {
		var (
			w  = child.widget
			P  = w.Point()
			S  = w.Size()
			P2 = render.Point{
				X: P.X + S.W,
				Y: P.Y + S.H,
			}
		)

		if XY.X >= P.X && XY.X <= P2.X && XY.Y >= P.Y && XY.Y <= P2.Y {
			// Cursor intersects the widget.
			hovering = append(hovering, child)
		} else {
			outside = append(outside, child)
		}
	}

	return hovering, outside
}

// Widgets returns a channel of widgets managed by the supervisor in the order
// they were added.
func (s *Supervisor) Widgets() <-chan WidgetSlot {
	pipe := make(chan WidgetSlot)
	go func() {
		for i := 0; i < s.serial; i++ {
			if w, ok := s.widgets[i]; ok {
				pipe <- w
			}
		}
		close(pipe)
	}()
	return pipe
}

// Present all widgets managed by the supervisor.
func (s *Supervisor) Present(e render.Engine) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	for child := range s.Widgets() {
		var w = child.widget
		w.Present(e, w.Point())
	}
}

// Add a widget to be supervised.
func (s *Supervisor) Add(w ...Widget) {
	s.lock.Lock()

	for _, child := range w {
		s.widgets[s.serial] = WidgetSlot{
			id:     s.serial,
			widget: child,
		}
		s.serial++
	}
	s.lock.Unlock()
}
