package scripting

import (
	"fmt"
	"reflect"
	"time"

	"git.kirsle.net/apps/doodle/lib/render"
	"git.kirsle.net/apps/doodle/pkg/log"
	"git.kirsle.net/apps/doodle/pkg/shmem"
	"github.com/robertkrimen/otto"
)

// VM manages a single isolated JavaScript VM.
type VM struct {
	Name string

	// Globals available to the scripts.
	Events *Events
	Self   interface{}

	// Channels for inbound and outbound PubSub messages.
	// Each VM has a single Inbound channel that watches for received messages
	//     and invokes the Message.Subscribe() handlers for relevant ones.
	// Each VM also has an array of Outbound channels which map to the Inbound
	//     channel of the VMs it is linked to, for pushing out Message.Publish()
	//     messages.
	Inbound   chan Message
	Outbound  []chan Message
	subscribe map[string][]otto.Value // Subscribed message handlers by name.

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

		// Pub/sub structs.
		Inbound:   make(chan Message),
		Outbound:  []chan Message{},
		subscribe: map[string][]otto.Value{},
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
		"Flash":  shmem.Flash,
		"RGBA":   render.RGBA,
		"Point":  render.NewPoint,
		"Self":   vm.Self, // i.e., the uix.Actor object
		"Events": vm.Events,
		"GetTick": func() uint64 {
			return shmem.Tick
		},

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
