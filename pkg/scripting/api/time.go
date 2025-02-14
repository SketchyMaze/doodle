// Package api defines the JavaScript API surface area for Doodad scripts.
package api

import (
	"time"

	"github.com/dop251/goja"
)

// Time functions exposed on the global scope like `time.Now()`.
//
// They are mainly a direct export of useful things from the Go `time` package.
type Time struct {
	Now   func() time.Time
	Since func(time.Time) time.Duration
	Add   func(t time.Time, milliseconds int64) time.Time

	// Multiples of time.Duration.
	Hour        time.Duration
	Minute      time.Duration
	Second      time.Duration
	Millisecond time.Duration
	Microsecond time.Duration
}

// NewTime creates an instanced global `time` object for your JavaScript VM.
func NewTime() Time {
	return Time{
		Now:   time.Now,
		Since: time.Since,
		Add: func(t time.Time, ms int64) time.Time {
			return t.Add(time.Duration(ms) * time.Millisecond)
		},
		Hour:        time.Hour,
		Minute:      time.Minute,
		Second:      time.Second,
		Millisecond: time.Millisecond,
		Microsecond: time.Microsecond,
	}
}

// ToMap converts the struct into a generic hash map.
func (g Time) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"Now":         g.Now,
		"Since":       g.Since,
		"Add":         g.Add,
		"Hour":        g.Hour,
		"Minute":      g.Minute,
		"Second":      g.Second,
		"Millisecond": g.Millisecond,
		"Microsecond": g.Microsecond,
	}
}

// Register the global functions in your JavaScript VM.
func (g Time) Register(vm *goja.Runtime) {
	for k, v := range g.ToMap() {
		vm.Set(k, v)
	}
}
