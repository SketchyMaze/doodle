package scripting

import (
	"git.kirsle.net/apps/doodle/pkg/log"
	"github.com/robertkrimen/otto"
)

// Message holds data being published from one script VM with information sent
// to the linked VMs.
type Message struct {
	Name string
	Args []interface{}
}

/*
RegisterPublishHooks adds the pub/sub hooks to a JavaScript VM.

This adds the global methods `Message.Subscribe(name, func)` and
`Message.Publish(name, args)` to the JavaScript VM's scope.
*/
func RegisterPublishHooks(vm *VM) {
	// Goroutine to watch the VM's inbound channel and invoke Subscribe handlers
	// for any matching messages received.
	go func() {
		for msg := range vm.Inbound {
			log.Warn("vm: %s  msg: %+v", vm.Name, msg)
			if _, ok := vm.subscribe[msg.Name]; ok {
				for _, callback := range vm.subscribe[msg.Name] {
					callback.Call(otto.Value{}, msg.Args...)
				}
			}
		}
	}()

	// Register the Message.Subscribe and Message.Publish functions.
	vm.vm.Set("Message", map[string]interface{}{
		"Subscribe": func(name string, callback otto.Value) {
			log.Error("SUBSCRIBE: %s", name)
			if !callback.IsFunction() {
				log.Error("SUBSCRIBE(%s): callback is not a function", name)
				return
			}
			if _, ok := vm.subscribe[name]; !ok {
				vm.subscribe[name] = []otto.Value{}
			}

			vm.subscribe[name] = append(vm.subscribe[name], callback)
		},

		"Publish": func(name string, v ...interface{}) {
			log.Error("PUBLISH: %s %+v", name, v)
			for _, channel := range vm.Outbound {
				channel <- Message{
					Name: name,
					Args: v,
				}
			}
		},
	})
}