package scripting

import (
	"errors"
	"fmt"
	"sync"

	"git.kirsle.net/SketchyMaze/doodle/pkg/keybind"
	"git.kirsle.net/SketchyMaze/doodle/pkg/log"
	"github.com/dop251/goja"
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
	runtime  *goja.Runtime
	registry map[string][]goja.Value
	lock     sync.RWMutex
}

// NewEvents initializes the Events API.
func NewEvents(runtime *goja.Runtime) *Events {
	return &Events{
		runtime:  runtime,
		registry: map[string][]goja.Value{},
	}
}

// OnCollide fires when another actor collides with yours.
func (e *Events) OnCollide(call goja.FunctionCall) goja.Value {
	return e.register(CollideEvent, call.Argument(0))
}

// RunCollide invokes the OnCollide handler function.
func (e *Events) RunCollide(v interface{}) error {
	return e.run(CollideEvent, v)
}

// OnUse fires when another actor collides with yours.
func (e *Events) OnUse(call goja.FunctionCall) goja.Value {
	return e.register(UseEvent, call.Argument(0))
}

// RunUse invokes the OnUse handler function.
func (e *Events) RunUse(v interface{}) error {
	return e.run(UseEvent, v)
}

// OnLeave fires when another actor stops colliding with yours.
func (e *Events) OnLeave(call goja.FunctionCall) goja.Value {
	return e.register(LeaveEvent, call.Argument(0))
}

// RunLeave invokes the OnLeave handler function.
func (e *Events) RunLeave(v interface{}) error {
	return e.run(LeaveEvent, v)
}

// OnKeypress fires when another actor collides with yours.
func (e *Events) OnKeypress(call goja.FunctionCall) goja.Value {
	return e.register(KeypressEvent, call.Argument(0))
}

// RunKeypress invokes the OnCollide handler function.
func (e *Events) RunKeypress(ev keybind.State) error {
	return e.run(KeypressEvent, e.runtime.ToValue(ev))
}

// register a named event.
func (e *Events) register(name string, callback goja.Value) goja.Value {
	e.lock.Lock()
	defer e.lock.Unlock()

	if _, ok := e.registry[name]; !ok {
		e.registry[name] = []goja.Value{}
	}

	e.registry[name] = append(e.registry[name], callback)
	return goja.Undefined()
}

// Run an event handler. Returns an error only if there was a JavaScript error
// inside the function. If there are no event handlers, just returns nil.
func (e *Events) run(name string, args ...interface{}) error {
	e.lock.RLock()
	defer e.lock.RUnlock()

	defer func() {
		if err := recover(); err != nil {
			// TODO EXCEPTIONS: I once saw a "runtime error: index out of range [-1]"
			// from an OnCollide handler between azu-white and thief that was crashing
			// the app, report this upstream nicely to the user.
			log.Error("PANIC: JS %s handler: %s", name, err)
		}
	}()

	if _, ok := e.registry[name]; !ok {
		return nil
	}

	var params = make([]goja.Value, len(args))
	for i, v := range args {
		params[i] = e.runtime.ToValue(v)
	}

	for _, callback := range e.registry[name] {
		function, ok := goja.AssertFunction(callback)
		if !ok {
			return fmt.Errorf("failed to callback %s: %s", name, callback)
		}

		value, err := function(goja.Undefined(), params...)
		if err != nil {
			// TODO EXCEPTIONS: this err is useful like
			// `ReferenceError: playerSpeed is not defined at <eval>:173:9(93)`
			// but wherever we're returning the err to isn't handling it!
			log.Error("Scripting error on %s: %s", name, err)
			return err
		}

		// If the event handler returned a boolean false, stop all other
		// callbacks and return the boolean.
		if b, ok := value.Export().(bool); ok && !b {
			return ErrReturnFalse
		}
	}

	return nil
}
