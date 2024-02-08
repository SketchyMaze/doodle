package scripting

import (
	"git.kirsle.net/SketchyMaze/doodle/pkg/log"
	"git.kirsle.net/SketchyMaze/doodle/pkg/scripting/exceptions"
	"github.com/dop251/goja"
)

// Message holds data being published from one script VM with information sent
// to the linked VMs.
type Message struct {
	Name     string
	SenderID string
	Args     []goja.Value
}

/*
RegisterPublishHooks adds the pub/sub hooks to a JavaScript VM.

This adds the global methods `Message.Subscribe(name, func)` and
`Message.Publish(name, args)` to the JavaScript VM's scope.
*/
func RegisterPublishHooks(s *Supervisor, vm *VM) {
	// Goroutine to watch the VM's inbound channel and invoke Subscribe handlers
	// for any matching messages received.
	go func() {
		// Catch any exceptions raised by the JavaScript VM.
		defer func() {
			if err := recover(); err != nil {
				exceptions.FormatAndCatch(vm.vm, "RegisterPublishHooks(%s): %s: %s", vm.Name, err)
			}
		}()

		// Watch the Inbound channel for PubSub messages and the stop channel for Teardown.
		for {
			select {
			case <-vm.stop:
				log.Debug("JavaScript VM %s stopping PubSub goroutine", vm.Name)
				return
			case msg := <-vm.Inbound:
				vm.muSubscribe.Lock()

				if _, ok := vm.subscribe[msg.Name]; ok {
					for _, callback := range vm.subscribe[msg.Name] {
						log.Debug("PubSub: %s receives from %s: %s", vm.Name, msg.SenderID, msg.Name)
						if function, ok := goja.AssertFunction(callback); ok {
							function(goja.Undefined(), msg.Args...)
						}
					}
				}

				vm.muSubscribe.Unlock()
			}
		}
	}()

	// Register the Message.Subscribe and Message.Publish functions.
	vm.vm.Set("Message", map[string]interface{}{
		"Subscribe": func(name string, callback goja.Value) {
			vm.muSubscribe.Lock()
			defer vm.muSubscribe.Unlock()

			if _, ok := goja.AssertFunction(callback); !ok {
				log.Error("SUBSCRIBE(%s): callback is not a function", name)
				return
			}
			if _, ok := vm.subscribe[name]; !ok {
				vm.subscribe[name] = []goja.Value{}
			}

			vm.subscribe[name] = append(vm.subscribe[name], callback)
		},

		"Publish": func(name string, v ...goja.Value) {
			vm.muPublish.Lock()
			for _, channel := range vm.Outbound {
				channel <- Message{
					Name:     name,
					SenderID: vm.Name,
					Args:     v,
				}
			}
			vm.muPublish.Unlock()
		},

		"Broadcast": func(name string, v ...goja.Value) {
			// Send the message to all actor VMs.
			for _, toVM := range s.scripts {
				if toVM == nil {
					continue
				}

				if vm.Name == toVM.Name {
					log.Debug("Broadcast(%s): skip to vm '%s' cuz it is the sender", name, toVM.Name)
					continue
				}

				toVM.Inbound <- Message{
					Name:     name,
					SenderID: vm.Name,
					Args:     v,
				}
			}
		},
	})
}
