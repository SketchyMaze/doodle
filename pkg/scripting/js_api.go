package scripting

import (
	"fmt"
	"time"

	"git.kirsle.net/SketchyMaze/doodle/pkg/log"
	"git.kirsle.net/SketchyMaze/doodle/pkg/physics"
	"git.kirsle.net/SketchyMaze/doodle/pkg/shmem"
	"git.kirsle.net/SketchyMaze/doodle/pkg/sound"
	"git.kirsle.net/go/render"
)

// SEE ALSO: uix/scripting.go for more global functions

// JSProxy offers a function API interface to expose to Doodad javascripts.
// These methods safely give the JS access to important attributes and functions
// without exposing unintended API surface area in the process.
type JSProxy map[string]interface{}

// ProxyLog wraps a console.log function to inject the script's name.
func ProxyLog(vm *VM, fn func(string, ...interface{})) func(string, ...interface{}) {
	var prefix = fmt.Sprintf("[%s] ", vm.Name)
	return func(msg string, v ...interface{}) {
		fn(prefix+msg, v...)
	}
}

// NewJSProxy initializes the API structure for JavaScript binding.
func NewJSProxy(vm *VM) JSProxy {

	return JSProxy{
		// Console logging.
		"console": map[string]interface{}{
			"log":   ProxyLog(vm, log.Info),
			"debug": ProxyLog(vm, log.Debug),
			"warn":  ProxyLog(vm, log.Warn),
			"error": ProxyLog(vm, log.Error),
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
			"Now":   time.Now,
			"Since": time.Since,
			"Add": func(t time.Time, ms int64) time.Time {
				return t.Add(time.Duration(ms) * time.Millisecond)
			},
			"Hour":        time.Hour,
			"Minute":      time.Minute,
			"Second":      time.Second,
			"Millisecond": time.Millisecond,
			"Microsecond": time.Microsecond,
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
