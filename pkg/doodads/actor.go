package doodads

import (
	"git.kirsle.net/go/render"
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

	// Actor's elected hitbox set by their script.
	// SetHitbox(x, y, w, h int)
	// Hitbox() render.Rect

	// Movement commands.
	MoveBy(render.Point) // Add {X,Y} to current Position.
	MoveTo(render.Point) // Set current Position to {X,Y}.
}

// GetBoundingRect computes the full pairs of points for the bounding box of
// the actor.
//
// The X,Y coordinates are the position in the level of the actor,
// The W,H are the size of the actor's drawn box.
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

// GetBoundingRectHitbox returns the bounding rect of the Actor taking into
// account their self-declared collision hitbox.
//
// The rect returned has the X,Y coordinate set to the actor's position, plus
// the X,Y of their hitbox, if any.
//
// The W,H of the rect is the W,H of their declared hitbox.
//
// If the actor has NOT declared its hitbox, this function returns exactly the
// same way as GetBoundingRect() does.
func GetBoundingRectHitbox(d Actor, hitbox render.Rect) render.Rect {
	rect := GetBoundingRect(d)
	if !hitbox.IsZero() {
		rect.X += hitbox.X
		rect.Y += hitbox.Y
		rect.W = hitbox.W
		rect.H = hitbox.H
	}
	return rect
}

// GetBoundingRectWithHitbox is like GetBoundingRect but adjusts it for the
// relative hitbox of the actor.
// func GetBoundingRectWithHitbox(d Actor, hitbox render.Rect) render.Rect {
// 	rect := GetBoundingRect(d)
// 	rect.W = hitbox.W
// 	rect.H = hitbox.H
// 	return rect
// }
