package scripting

import (
	"errors"
	"sync"

	"git.kirsle.net/apps/doodle/pkg/keybind"
	"github.com/robertkrimen/otto"
)

// Event name constants.
const (
	CollideEvent = "OnCollide" // another doodad collides with us
	EnterEvent   = "OnEnter"   // a doodad is fully inside us
	LeaveEvent   = "OnLeave"   // a doodad no longer collides with us
	UseEvent     = "OnUse"     // player pressed the Use key while touching us

	// Controllable (player character) doodad events
	KeypressEvent = "OnKeypress" // i.e. arrow keys
)

// Event return errors.
var (
	ErrReturnFalse = errors.New("JS callback function returned false")
)

// Events API for Doodad scripts.
type Events struct {
	registry map[string][]otto.Value
	lock     sync.RWMutex
}

// NewEvents initializes the Events API.
func NewEvents() *Events {
	return &Events{
		registry: map[string][]otto.Value{},
	}
}

// OnCollide fires when another actor collides with yours.
func (e *Events) OnCollide(call otto.FunctionCall) otto.Value {
	return e.register(CollideEvent, call.Argument(0))
}

// RunCollide invokes the OnCollide handler function.
func (e *Events) RunCollide(v interface{}) error {
	return e.run(CollideEvent, v)
}

// OnUse fires when another actor collides with yours.
func (e *Events) OnUse(call otto.FunctionCall) otto.Value {
	return e.register(UseEvent, call.Argument(0))
}

// RunUse invokes the OnUse handler function.
func (e *Events) RunUse(v interface{}) error {
	return e.run(UseEvent, v)
}

// OnLeave fires when another actor stops colliding with yours.
func (e *Events) OnLeave(call otto.FunctionCall) otto.Value {
	return e.register(LeaveEvent, call.Argument(0))
}

// RunLeave invokes the OnLeave handler function.
func (e *Events) RunLeave(v interface{}) error {
	return e.run(LeaveEvent, v)
}

// OnKeypress fires when another actor collides with yours.
func (e *Events) OnKeypress(call otto.FunctionCall) otto.Value {
	return e.register(KeypressEvent, call.Argument(0))
}

// RunKeypress invokes the OnCollide handler function.
func (e *Events) RunKeypress(ev keybind.State) error {
	return e.run(KeypressEvent, ev)
}

// register a named event.
func (e *Events) register(name string, callback otto.Value) otto.Value {
	if !callback.IsFunction() {
		return otto.Value{} // TODO
	}

	e.lock.Lock()
	defer e.lock.Unlock()

	if _, ok := e.registry[name]; !ok {
		e.registry[name] = []otto.Value{}
	}

	e.registry[name] = append(e.registry[name], callback)
	return otto.Value{}
}

// Run an event handler. Returns an error only if there was a JavaScript error
// inside the function. If there are no event handlers, just returns nil.
func (e *Events) run(name string, args ...interface{}) error {
	e.lock.RLock()
	defer e.lock.RUnlock()

	if _, ok := e.registry[name]; !ok {
		return nil
	}

	for _, callback := range e.registry[name] {
		value, err := callback.Call(otto.Value{}, args...)
		if err != nil {
			return err
		}

		// If the event handler returned a boolean false, stop all other
		// callbacks and return the boolean.
		if value.IsBoolean() {
			if b, err := value.ToBoolean(); err == nil && b == false {
				return ErrReturnFalse
			}
		}
	}

	return nil
}
