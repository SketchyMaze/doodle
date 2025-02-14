package api

import (
	"git.kirsle.net/go/render"
	"github.com/dop251/goja"
)

// GameplayLevelControl
type GameplayLevelControl struct {
	// EndLevel ends the game in a "win" condition.
	EndLevel func()

	// FailLevel ends the game in a "fail" condition, with the message shown to the player.
	FailLevel func(message string)

	// SetCheckpoint updates the player's respawn position to the given point.
	SetCheckpoint func(render.Point)
}

// ToMap converts the struct into a generic hash map.
func (g GameplayLevelControl) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"EndLevel":      g.EndLevel,
		"FailLevel":     g.FailLevel,
		"SetCheckpoint": g.SetCheckpoint,
	}
}

// Register the global functions in your JavaScript VM.
func (g GameplayLevelControl) Register(vm *goja.Runtime) {
	for k, v := range g.ToMap() {
		vm.Set(k, v)
	}
}
