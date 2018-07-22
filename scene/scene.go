package scene

import (
	"errors"
	"fmt"

	"git.kirsle.net/apps/doodle/events"
)

// Scene is an interface for the top level of a game mode. The game points to
// one Scene at a time, and that Scene has majority control of the main loop,
// and maintains its own state local to that scene.
type Scene interface {
	String() string // the scene's name
	Setup() error
	Loop() error
	Destroy() error
}

// Manager is a type that provides context switching features to manage scenes.
type Manager struct {
	events *events.State
	scene  Scene
	ticks  uint64
}

// NewManager creates the new manager.
func NewManager(events *events.State) Manager {
	return Manager{
		events: events,
		scene:  nil,
	}
}

// Go to a new scene. This tears down the existing scene, sets up the new one,
// and switches control to the new scene.
func (m *Manager) Go(scene Scene) error {
	// Already running a scene?
	if m.scene != nil {
		if err := m.scene.Destroy(); err != nil {
			return fmt.Errorf("couldn't destroy scene %s: %s", m.scene, err)
		}
		m.scene = nil
	}

	// Initialize the new scene.
	m.scene = scene
	return m.scene.Setup()
}

// Loop the scene manager. This is the game's main loop which runs all the tasks
// that fall in the realm of the scene manager.
func (m *Manager) Loop() error {
	if m.scene == nil {
		return errors.New("no scene loaded")
	}

	// Poll for events.
	ev, err := m.events.Poll(m.ticks)
	if err != nil {
		log.Error("event poll error: %s", err)
		return err
	}
	_ = ev

	return m.scene.Loop()
}
