package doodle

import "git.kirsle.net/apps/doodle/events"

// Scene is an abstraction for a game mode in Doodle. The app points to one
// scene at a time and that scene has control over the main loop, and its own
// state information.
type Scene interface {
	Name() string
	Setup(*Doodle) error
	Loop(*Doodle, *events.State) error
}

// Goto a scene. First it unloads the current scene.
func (d *Doodle) Goto(scene Scene) error {
	// d.scene.Destroy()
	log.Info("Goto Scene")
	d.scene = scene
	return d.scene.Setup(d)
}
