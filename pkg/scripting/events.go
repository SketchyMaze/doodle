package scripting

import (
	"git.kirsle.net/apps/doodle/lib/events"
	"github.com/robertkrimen/otto"
)

// Event name constants.
const (
	CollideEvent = "OnCollide" // another doodad collides with us
	EnterEvent   = "OnEnter"   // a doodad is fully inside us
	LeaveEvent   = "OnLeave"   // a doodad no longer collides with us

	// Controllable (player character) doodad events
	KeypressEvent = "OnKeypress" // i.e. arrow keys
)

// Events API for Doodad scripts.
type Events struct {
	registry map[string][]otto.Value
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
func (e *Events) RunCollide() error {
	return e.run(CollideEvent)
}

// OnKeypress fires when another actor collides with yours.
func (e *Events) OnKeypress(call otto.FunctionCall) otto.Value {
	return e.register(KeypressEvent, call.Argument(0))
}

// RunKeypress invokes the OnCollide handler function.
func (e *Events) RunKeypress(ev *events.State) error {
	return e.run(KeypressEvent, ev)
}

// register a named event.
func (e *Events) register(name string, callback otto.Value) otto.Value {
	if !callback.IsFunction() {
		return otto.Value{} // TODO
	}

	if _, ok := e.registry[name]; !ok {
		e.registry[name] = []otto.Value{}
	}

	e.registry[name] = append(e.registry[name], callback)
	return otto.Value{}
}

// Run an event handler. Returns an error only if there was a JavaScript error
// inside the function. If there are no event handlers, just returns nil.
func (e *Events) run(name string, args ...interface{}) error {
	if _, ok := e.registry[name]; !ok {
		return nil
	}

	for _, callback := range e.registry[name] {
		_, err := callback.Call(otto.Value{}, args...)
		if err != nil {
			return err
		}
	}

	return nil
}
