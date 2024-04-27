package scripting

import (
	"time"

	"git.kirsle.net/SketchyMaze/doodle/pkg/balance"
	"git.kirsle.net/SketchyMaze/doodle/pkg/shmem"
	"github.com/dop251/goja"
)

// Timer keeps track of delayed function calls for the scripting engine.
type Timer struct {
	id       int
	callback goja.Value
	ticks    uint64 // interval (milliseconds) converted into game ticks
	nextTick uint64 // next tick to trigger the callback
	repeat   bool   // for setInterval
}

/*
SetTimeout registers a callback function to be run after a while.

This is to be called by JavaScript running in the VM and has an API similar to
that found in web browsers.

The callback is a JavaScript function and the interval is in milliseconds,
with 1000 being 'one second.'

Returns the ID number of the timer in case you want to clear it. The underlying
Timer type is NOT exposed to JavaScript.
*/
func (vm *VM) SetTimeout(callback goja.Value, interval int) int {
	return vm.AddTimer(callback, interval, false)
}

/*
SetInterval registers a callback function to be run repeatedly.

Returns the ID number of the timer in case you want to clear it. The underlying
Timer type is NOT exposed to JavaScript.
*/
func (vm *VM) SetInterval(callback goja.Value, interval int) int {
	return vm.AddTimer(callback, interval, true)
}

/*
AddTimer loads timeouts and intervals into the VM's memory and returns the ID.
*/
func (vm *VM) AddTimer(callback goja.Value, interval int, repeat bool) int {
	// Get the next timer ID. The first timer has ID 1.
	vm.timerLastID++

	var (
		id    = vm.timerLastID
		ticks = float64(interval) * (float64(balance.TargetFPS) / 1000)
	)

	t := &Timer{
		id:       id,
		callback: callback,
		ticks:    uint64(ticks),
		repeat:   repeat,
	}
	t.Schedule()
	vm.timers[id] = t

	return id
}

// TickTimer checks if any timers are ready and calls their functions.
func (vm *VM) TickTimer(now time.Time) {
	if len(vm.timers) == 0 {
		return
	}

	// IDs of expired timeouts to clear.
	var clear []int

	for id, timer := range vm.timers {
		if shmem.Tick > timer.nextTick {
			if function, ok := goja.AssertFunction(timer.callback); ok {
				function(goja.Undefined())
			}

			if timer.repeat {
				timer.Schedule()
			} else {
				clear = append(clear, id)
			}
		}
	}

	// Clean up expired timers.
	if len(clear) > 0 {
		for _, id := range clear {
			delete(vm.timers, id)
		}
	}
}

/*
ClearTimer will clear both timeouts and intervals.

In the JavaScript VM this function is bound to clearTimeout() and clearInterval()
to expose an API like that seen in web browsers.
*/
func (vm *VM) ClearTimer(id int) {
	delete(vm.timers, id)
}

// Schedule the callback to be run in the future.
func (t *Timer) Schedule() {
	t.nextTick = shmem.Tick + t.ticks
}
