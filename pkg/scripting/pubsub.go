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
			// Send the message to all actor VMs, in deterministic (sorted by
			// actor ID) order.
			for _, id := range s.sortedIDs() {
				toVM := s.scripts[id]
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

/*
DrainInbound synchronously processes every PubSub message currently queued on
the VM's Inbound channel, invoking any matching Message.Subscribe handlers.

This is called once per tick by the Supervisor, for each VM in turn, so that
doodad scripts never receive or handle PubSub messages concurrently with each
other (or with the goja.Runtime of the VM being used elsewhere on the main
goroutine, e.g. for OnCollide/OnUse handlers). goja.Runtime is not safe for
concurrent use, and running this on a per-VM background goroutine (the old
approach) allowed a burst of cross-linked messages -- e.g. several doodads
like linked totems all colliding in the same tick -- to invoke JS callbacks
on the same VM from multiple goroutines at once, which could corrupt VM state
or deadlock on the channel sends in Message.Publish/Broadcast.
*/
func (vm *VM) DrainInbound() {
	defer func() {
		if err := recover(); err != nil {
			exceptions.FormatAndCatch(vm.vm, "DrainInbound(%s): %s", vm.Name, err)
		}
	}()

	for {
		select {
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
		default:
			// No more messages queued right now.
			return
		}
	}
}
