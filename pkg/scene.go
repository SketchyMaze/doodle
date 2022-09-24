package doodle

import (
	"git.kirsle.net/SketchyMaze/doodle/pkg/gamepad"
	"git.kirsle.net/SketchyMaze/doodle/pkg/log"
	"git.kirsle.net/go/render/event"
)

// Scene is an abstraction for a game mode in Doodle. The app points to one
// scene at a time and that scene has control over the main loop, and its own
// state information.
type Scene interface {
	Name() string
	Setup(*Doodle) error
	Destroy() error

	// Loop should update the scene's state but not draw anything.
	Loop(*Doodle, *event.State) error

	// Draw should use the scene's state to figure out what pixels need
	// to draw to the screen.
	Draw(*Doodle) error
}

// Goto a scene. First it unloads the current scene.
func (d *Doodle) Goto(scene Scene) error {
	// Inform the gamepad controller what scene.
	gamepad.SceneName = scene.Name()

	// Clear any debug labels.
	customDebugLabels = []debugLabel{}

	// Teardown existing scene.
	if d.Scene != nil {
		d.Scene.Destroy()
	}

	log.Info("Goto Scene: %s", scene.Name())
	d.Scene = scene
	return d.Scene.Setup(d)
}
