package scripting

import (
	"time"

	"github.com/robertkrimen/otto"
)

// Timer keeps track of delayed function calls for the scripting engine.
type Timer struct {
	id       int
	callback otto.Value
	interval time.Duration // milliseconds delay for timeout
	next     time.Time     // scheduled time for next invocation
	repeat   bool          // for setInterval
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
func (vm *VM) SetTimeout(callback otto.Value, interval int) int {
	return vm.AddTimer(callback, interval, false)
}

/*
SetInterval registers a callback function to be run repeatedly.

Returns the ID number of the timer in case you want to clear it. The underlying
Timer type is NOT exposed to JavaScript.
*/
func (vm *VM) SetInterval(callback otto.Value, interval int) int {
	return vm.AddTimer(callback, interval, true)
}

/*
AddTimer loads timeouts and intervals into the VM's memory and returns the ID.
*/
func (vm *VM) AddTimer(callback otto.Value, interval int, repeat bool) int {
	// Get the next timer ID. The first timer has ID 1.
	vm.timerLastID++
	id := vm.timerLastID

	t := &Timer{
		id:       id,
		callback: callback,
		interval: time.Duration(interval),
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
		if now.After(timer.next) {
			timer.callback.Call(otto.Value{})
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
	t.next = time.Now().Add(t.interval * time.Millisecond)
}
