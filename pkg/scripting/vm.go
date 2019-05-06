package scripting

import (
	"fmt"
	"reflect"
	"time"

	"git.kirsle.net/apps/doodle/lib/render"
	"git.kirsle.net/apps/doodle/pkg/log"
	"github.com/robertkrimen/otto"
)

// VM manages a single isolated JavaScript VM.
type VM struct {
	Name string

	// Globals available to the scripts.
	Events *Events
	Self   interface{}

	vm *otto.Otto

	// setTimeout and setInterval variables.
	timerLastID int // becomes 1 when first timer is set
	timers      map[int]*Timer
}

// NewVM creates a new JavaScript VM.
func NewVM(name string) *VM {
	vm := &VM{
		Name:   name,
		Events: NewEvents(),
		vm:     otto.New(),
		timers: map[int]*Timer{},
	}
	return vm
}

// Run code in the VM.
func (vm *VM) Run(src interface{}) (otto.Value, error) {
	v, err := vm.vm.Run(src)
	return v, err
}

// Set a value in the VM.
func (vm *VM) Set(name string, v interface{}) error {
	return vm.vm.Set(name, v)
}

// RegisterLevelHooks registers accessors to the level hooks
// and Doodad API for Play Mode.
func (vm *VM) RegisterLevelHooks() error {
	bindings := map[string]interface{}{
		"log":    log.Logger,
		"RGBA":   render.RGBA,
		"Point":  render.NewPoint,
		"Self":   vm.Self, // i.e., the uix.Actor object
		"Events": vm.Events,

		"TypeOf": reflect.TypeOf,
		"time": map[string]interface{}{
			"Now": time.Now,
			"Add": func(t time.Time, ms int64) time.Time {
				return t.Add(time.Duration(ms) * time.Millisecond)
			},
		},

		// Timer functions with APIs similar to the web browsers.
		"setTimeout":    vm.SetTimeout,
		"setInterval":   vm.SetInterval,
		"clearTimeout":  vm.ClearTimer,
		"clearInterval": vm.ClearTimer,
	}
	for name, v := range bindings {
		err := vm.vm.Set(name, v)
		if err != nil {
			return fmt.Errorf("RegisterLevelHooks(%s): %s",
				name, err,
			)
		}
	}

	// Alias the console.log functions to the logger.
	vm.vm.Run(`
		console = {};
		console.log = log.Info;
		console.debug = log.Debug;
		console.warn = log.Warn;
		console.error = log.Error;
	`)
	return nil
}

// Main calls the main function of the script.
func (vm *VM) Main() error {
	function, err := vm.vm.Get("main")
	if err != nil {
		return err
	}

	if !function.IsFunction() {
		return nil
	}

	// Catch panics.
	defer func() {
		if err := recover(); err != nil {
			log.Error("Panic caught in JavaScript VM: %s", err)
		}
	}()

	_, err = function.Call(otto.Value{})
	return err
}
