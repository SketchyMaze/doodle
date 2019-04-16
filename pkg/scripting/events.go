package scripting

import (
	"github.com/robertkrimen/otto"
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
	callback := call.Argument(0)
	if !callback.IsFunction() {
		return otto.Value{} // TODO
	}

	if _, ok := e.registry[CollideEvent]; !ok {
		e.registry[CollideEvent] = []otto.Value{}
	}

	e.registry[CollideEvent] = append(e.registry[CollideEvent], callback)
	return otto.Value{}
}

// RunCollide invokes the OnCollide handler function.
func (e *Events) RunCollide() error {
	if _, ok := e.registry[CollideEvent]; !ok {
		return nil
	}

	for _, callback := range e.registry[CollideEvent] {
		_, err := callback.Call(otto.Value{}, "test argument")
		if err != nil {
			return err
		}
	}

	return nil
}

// Event name constants.
const (
	CollideEvent = "collide"
)
