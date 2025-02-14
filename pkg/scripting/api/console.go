package api

import "github.com/dop251/goja"

/*
Console is exported to the global scope with `console.log` and friends available.

These functions have an API similar to that found in web browsers and node.js.
*/
type Console struct {
	Log   func(string, ...interface{}) `json:"log"`
	Debug func(string, ...interface{}) `json:"debug"`
	Warn  func(string, ...interface{}) `json:"warn"`
	Error func(string, ...interface{}) `json:"error"`
}

// ToMap converts the struct into a generic hash map.
func (g Console) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"log":   g.Log,
		"debug": g.Debug,
		"warn":  g.Warn,
		"error": g.Error,
	}
}

// Register the global functions in your JavaScript VM.
func (g Console) Register(vm *goja.Runtime) {
	for k, v := range g.ToMap() {
		vm.Set(k, v)
	}
}
