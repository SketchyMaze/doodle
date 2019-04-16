package level

import (
	"git.kirsle.net/apps/doodle/lib/render"
	uuid "github.com/satori/go.uuid"
)

// ActorMap holds the doodad information by their ID in the level data.
type ActorMap map[string]*Actor

// Inflate assigns each actor its ID from the hash map for their self reference.
func (m ActorMap) Inflate() {
	for id, actor := range m {
		actor.id = id
	}
}

// Add a new Actor to the map. If it doesn't already have an ID it will be
// given a random UUIDv4 ID.
func (m ActorMap) Add(a *Actor) {
	if a.id == "" {
		a.id = uuid.Must(uuid.NewV4()).String()
	}
	m[a.id] = a
}

// Remove an Actor from the map. The ID must be set at the very least, so to
// remove by ID just create an Actor{id: x}
func (m ActorMap) Remove(a *Actor) bool {
	if _, ok := m[a.id]; ok {
		delete(m, a.id)
		return true
	}
	return false
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