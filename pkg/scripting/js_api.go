package scripting

import (
	"time"

	"git.kirsle.net/apps/doodle/pkg/log"
	"git.kirsle.net/apps/doodle/pkg/physics"
	"git.kirsle.net/apps/doodle/pkg/shmem"
	"git.kirsle.net/apps/doodle/pkg/sound"
	"git.kirsle.net/go/render"
)

// JSProxy offers a function API interface to expose to Doodad javascripts.
// These methods safely give the JS access to important attributes and functions
// without exposing unintended API surface area in the process.
type JSProxy map[string]interface{}

// NewJSProxy initializes the API structure for JavaScript binding.
func NewJSProxy(vm *VM) JSProxy {
	return JSProxy{
		// Console logging.
		"console": map[string]interface{}{
			"log":   log.Info,
			"debug": log.Debug,
			"warn":  log.Warn,
			"error": log.Error,
		},

		// Audio API.
		"Sound": map[string]interface{}{
			"Play": sound.PlaySound,
		},

		// Type constructors.
		"RGBA":   render.RGBA,
		"Point":  render.NewPoint,
		"Vector": physics.NewVector,

		// Useful types and functions.
		"Flash": shmem.Flash,
		"GetTick": func() uint64 {
			return shmem.Tick
		},
		"time": map[string]interface{}{
			"Now": time.Now,
			"Add": func(t time.Time, ms int64) time.Time {
				return t.Add(time.Duration(ms) * time.Millisecond)
			},
		},

		// Bindings into the VM.
		"Events":        vm.Events,
		"setTimeout":    vm.SetTimeout,
		"setInterval":   vm.SetInterval,
		"clearTimeout":  vm.ClearTimer,
		"clearInterval": vm.ClearTimer,

		// Self for an actor to inspect themselves.
		"Self": vm.Self,
	}
}
