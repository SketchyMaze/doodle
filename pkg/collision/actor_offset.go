package collision

import "git.kirsle.net/go/render"

// ActorOffset helps normalize an actor's Position and Hitbox for collision detection.
//
// It allows for an actor to have a Hitbox which is offset from the 0,0 coordinate
// in the top left corner. During gameplay, the actor's Position (top left corner of
// its *sprite*) is what the game tracks for movement, but if the actor's declared hitbox
// doesn't encompass the point 0,0 it used to lead to collision bugs.
//
// ActorOffset will take your original Actor, compute the offset between its Position
// and its Hitbox, and return a simplified Actor that pretends the hitbox begins at 0,0.
type ActorOffset struct {
	d      Actor
	offset render.Point
}

// NewActorOffset consumes the game's original Actor and returns one that simplifies
// the Hitbox boundary.
func NewActorOffset(d Actor) *ActorOffset {
	// Compute the offset from the actor's Position to its Hitbox.
	var (
		position = d.Position()
		hitbox   = d.Hitbox()
		delta    = render.Point{
			X: position.X + hitbox.X,
			Y: position.Y + hitbox.Y,
		}
		offset = render.Point{
			X: delta.X - position.X,
			Y: delta.Y - position.Y,
		}
	)
	return &ActorOffset{
		d:      d,
		offset: offset,
	}
}

// Offset returns the offset from the source actor's Position to their new one.
func (ao *ActorOffset) Offset() render.Point {
	return ao.offset
}

// Position will be the actor's original world position (of its sprite) plus the
// hitbox offset coordinate.
func (ao *ActorOffset) Position() render.Point {
	var P = ao.d.Position()
	return render.Point{
		X: P.X + ao.offset.X,
		Y: P.Y + ao.offset.Y,
	}
}

// Size is the same as your original Actor.
func (ao *ActorOffset) Size() render.Rect {
	return ao.d.Size()
}

// Hitbox returns the actor's original Hitbox but where the X,Y are locked to 0,0.
// The W,H of the hitbox is the same as original.
func (ao *ActorOffset) Hitbox() render.Rect {
	var HB = ao.d.Hitbox()
	return render.Rect{
		X: 0,
		Y: 0,
		H: HB.H,
		W: HB.W,
	}
}

// Grounded returns your original actor's value.
func (ao *ActorOffset) Grounded() bool {
	return ao.d.Grounded()
}

// SetGrounded sets the grounded state of your original actor.
func (ao *ActorOffset) SetGrounded(v bool) {
	ao.d.SetGrounded(v)
}
