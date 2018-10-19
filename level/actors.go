package level

import "git.kirsle.net/apps/doodle/render"

// ActorMap holds the doodad information by their ID in the level data.
type ActorMap map[string]*Actor

// Inflate assigns each actor its ID from the hash map for their self reference.
func (m ActorMap) Inflate() {
	for id, actor := range m {
		actor.id = id
	}
}

// Actor is an instance of a Doodad in the level.
type Actor struct {
	id       string       // NOTE: read only, use ID() to access.
	Filename string       `json:"filename"` // like "exit.doodad"
	Point    render.Point `json:"point"`
}

// ID returns the actor's ID.
func (a *Actor) ID() string {
	return a.id
}
