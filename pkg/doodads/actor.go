package doodads

import (
	"git.kirsle.net/apps/doodle/lib/render"
)

// Actor is a reusable run-time drawing component used in Doodle. Actors are an
// active instance of a Doodad which have a position, velocity, etc.
type Actor interface {
	ID() string

	// Position and velocity, not saved to disk.
	Position() render.Point
	Velocity() render.Point
	Size() render.Rect
	Grounded() bool
	SetGrounded(bool)

	// Movement commands.
	MoveBy(render.Point) // Add {X,Y} to current Position.
	MoveTo(render.Point) // Set current Position to {X,Y}.
}

// GetBoundingRect computes the full pairs of points for the collision box
// around a doodad actor.
func GetBoundingRect(d Actor) render.Rect {
	var (
		P = d.Position()
		S = d.Size()
	)
	return render.Rect{
		X: P.X,
		Y: P.Y,
		W: S.W,
		H: S.H,
	}
}