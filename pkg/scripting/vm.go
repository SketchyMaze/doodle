package scripting

import (
	"errors"
	"fmt"
	"sync"

	"git.kirsle.net/SketchyMaze/doodle/pkg/log"
	"github.com/dop251/goja"
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
	Inbound     chan Message
	Outbound    []chan Message
	stop        chan bool
	subscribe   map[string][]goja.Value // Subscribed message handlers by name.
	muSubscribe sync.RWMutex
	muPublish   sync.Mutex // serialize PubSub publishes

	vm *goja.Runtime

	// setTimeout and setInterval variables.
	timerLastID int // becomes 1 when first timer is set
	timers      map[int]*Timer
}

// NewVM creates a new JavaScript VM.
func NewVM(name string) *VM {
	vm := &VM{
		Name:   name,
		vm:     goja.New(),
		timers: map[int]*Timer{},

		// Pub/sub structs.
		Inbound:   make(chan Message, 100),
		Outbound:  []chan Message{},
		stop:      make(chan bool, 1),
		subscribe: map[string][]goja.Value{},
	}
	vm.Events = NewEvents(vm)
	return vm
}

// Run code in the VM.
func (vm *VM) Run(src string) (goja.Value, error) {
	v, err := vm.vm.RunString(src)
	return v, err
}

// Set a value in the VM.
func (vm *VM) Set(name string, v interface{}) error {
	return vm.vm.Set(name, v)
}

// Get a value from the VM.
func (vm *VM) Get(name string) goja.Value {
	return vm.vm.Get(name)
}

// RegisterLevelHooks registers accessors to the level hooks
// and Doodad API for Play Mode.
func (vm *VM) RegisterLevelHooks() error {
	bindings := NewJSProxy(vm)
	for name, v := range bindings {
		err := vm.vm.Set(name, v)
		if err != nil {
			return fmt.Errorf("RegisterLevelHooks(%s): %s",
				name, err,
			)
		}
	}
	return nil
}

// Main calls the main function of the script.
func (vm *VM) Main() error {
	function, ok := goja.AssertFunction(vm.vm.Get("main"))
	if !ok {
		return errors.New("didn't find function main()")
	}

	// Catch panics.
	defer func() {
		if err := recover(); err != nil {
			log.Error("Panic caught in JavaScript VM: %s", err)
		}
	}()

	_, err := function(goja.Undefined())
	return err
}
