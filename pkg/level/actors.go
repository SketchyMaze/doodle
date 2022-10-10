package level

import (
	"git.kirsle.net/go/render"
	"github.com/google/uuid"
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
		a.id = uuid.Must(uuid.NewUUID()).String()
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
	id       string             // NOTE: read only, use ID() to access.
	Filename string             `json:"filename"` // like "exit.doodad"
	Point    render.Point       `json:"point"`
	Links    []string           `json:"links,omitempty"` // IDs of linked actors
	Options  map[string]*Option `json:"options,omitempty"`
}

// NewActor initializes a level.Actor.
func NewActor(a Actor) *Actor {
	return &Actor{
		Filename: a.Filename,
		Point:    a.Point,
		Links:    []string{},
		Options:  map[string]*Option{},
	}
}

// ID returns the actor's ID.
func (a *Actor) ID() string {
	return a.id
}

// AddLink adds a linked Actor to an Actor. Add the linked actor by its ID.
func (a *Actor) AddLink(id string) {
	// Don't add a duplicate ID to the links array.
	for _, exist := range a.Links {
		if exist == id {
			return
		}
	}
	a.Links = append(a.Links, id)
}

// Unlink removes the linked actor's ID.
func (a *Actor) Unlink(id string) {
	var newLinks []string
	for _, exist := range a.Links {
		if exist == id {
			continue
		}
		newLinks = append(newLinks, exist)
	}
	a.Links = newLinks
}

// IsLinked checks if the actor is linked to the other actor's ID.
func (a *Actor) IsLinked(id string) bool {
	for _, exist := range a.Links {
		if exist == id {
			return true
		}
	}
	return false
}
