package collision

import "git.kirsle.net/go/render"

// GetBoundingRect computes the full pairs of points for the bounding box of
// the actor.
//
// The X,Y coordinates are the position in the level of the actor,
// The W,H are the size of the actor's drawn box.
func GetBoundingRect(a Actor) render.Rect {
	var (
		P = a.Position()
		S = a.Size()
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
// the X,Y of their hitbox, if any. For example, their sprite size could be 64x32
// and their hitbox is the lower 0,32,32,32 half. This function would return the
// world coordinate of where their bounding box begins.
//
// The W,H of the rect is the W,H of their declared hitbox.
//
// If the actor has NOT declared its hitbox, this function returns exactly the
// same way as GetBoundingRect() does.
func GetBoundingRectHitbox(a Actor, hitbox render.Rect) render.Rect {
	rect := GetBoundingRect(a)
	if !hitbox.IsZero() {
		rect.X += hitbox.X
		rect.Y += hitbox.Y
		rect.W = hitbox.W
		rect.H = hitbox.H
	}
	return rect
}

// SizePlusHitbox adjusts an actor's canvas Size() to better fit the
// declared Hitbox by the actor's script.
func SizePlusHitbox(size render.Rect, hitbox render.Rect) render.Rect {
	size.X += hitbox.X
	size.Y += hitbox.Y
	size.W -= size.W - hitbox.W
	size.H -= size.H - hitbox.H
	return size
}
