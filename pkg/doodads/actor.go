package doodads

import (
	"git.kirsle.net/go/render"
)

// Actor is a reusable run-time drawing component used in Doodle. Actors are an
// active instance of a Doodad which have a position, velocity, etc.
type Actor interface {
	ID() string

	// Position and velocity, not saved to disk.
	Position() render.Point // DEPRECATED
	Velocity() render.Point // DEPRECATED for uix.Actor
	Size() render.Rect
	Grounded() bool
	SetGrounded(bool)

	// Actor's elected hitbox set by their script.
	// SetHitbox(x, y, w, h int)
	// Hitbox() render.Rect

	// Movement commands.
	MoveBy(render.Point) // Add {X,Y} to current Position.
	MoveTo(render.Point) // Set current Position to {X,Y}.
}
