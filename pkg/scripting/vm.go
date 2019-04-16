package scripting

import (
	"errors"
	"fmt"

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
}

// NewVM creates a new JavaScript VM.
func NewVM(name string) *VM {
	vm := &VM{
		Name:   name,
		Events: NewEvents(),
		vm:     otto.New(),
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
		"Self":   vm.Self,
		"Events": vm.Events,
	}
	for name, v := range bindings {
		err := vm.vm.Set(name, v)
		if err != nil {
			return fmt.Errorf("RegisterLevelHooks(%s): %s",
				name, err,
			)
		}
	}
	vm.vm.Run(`console = {}; console.log = log.Info;`)
	return nil
}

// Main calls the main function of the script.
func (vm *VM) Main() error {
	function, err := vm.vm.Get("main")
	if err != nil {
		return err
	}

	if !function.IsFunction() {
		return errors.New("main is not a function")
	}

	_, err = function.Call(otto.Value{})
	return err
}
