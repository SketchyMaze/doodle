package doodads

import (
	"git.kirsle.net/apps/doodle/level"
	"git.kirsle.net/apps/doodle/render"
)

// Collide describes how a collision occurred.
type Collide struct {
	Top         bool
	TopPoint    render.Point
	TopPixel    *level.Swatch
	Left        bool
	LeftPoint   render.Point
	LeftPixel   *level.Swatch
	Right       bool
	RightPoint  render.Point
	RightPixel  *level.Swatch
	Bottom      bool
	BottomPoint render.Point
	BottomPixel *level.Swatch
	MoveTo      render.Point
}

// Reset a Collide struct flipping all the bools off, but keeping MoveTo.
func (c *Collide) Reset() {
	c.Top = false
	c.Left = false
	c.Right = false
	c.Bottom = false
}

// CollisionBox holds all of the coordinate pairs to draw the collision box
// around a doodad.
type CollisionBox struct {
	Top    []render.Point
	Bottom []render.Point
	Left   []render.Point
	Right  []render.Point
}

// GetCollisionBox returns a CollisionBox with the four coordinates.
func GetCollisionBox(box render.Rect) CollisionBox {
	return CollisionBox{
		Top: []render.Point{
			{
				X: box.X,
				Y: box.Y,
			},
			{
				X: box.X + box.W,
				Y: box.Y,
			},
		},
		Bottom: []render.Point{
			{
				X: box.X,
				Y: box.Y + box.H,
			},
			{
				X: box.X + box.W,
				Y: box.Y + box.H,
			},
		},
		Left: []render.Point{
			{
				X: box.X,
				Y: box.Y + box.H - 1,
			},
			{
				X: box.X,
				Y: box.Y + 1,
			},
		},
		Right: []render.Point{
			{
				X: box.X + box.W,
				Y: box.Y + box.H - 1,
			},
			{
				X: box.X + box.W,
				Y: box.Y + 1,
			},
		},
	}
}

// Side of the collision box (top, bottom, left, right)
type Side uint8

// Options for the Side type.
const (
	Top Side = iota
	Bottom
	Left
	Right
)

// CollidesWithGrid checks if a Doodad collides with level geometry.
func CollidesWithGrid(d Actor, grid *level.Chunker, target render.Point) (*Collide, bool) {
	var (
		P = d.Position()
		S = d.Size()

		result = &Collide{
			MoveTo: P,
		}
		ceiling   bool  // Has hit a ceiling?
		capHeight int32 // Stop vertical movement thru a ceiling
		capLeft   int32 // Stop movement thru a wall
		capRight  int32
		hitLeft   bool // Has hit an obstacle on the left
		hitRight  bool // or right
		hitFloor  bool
		capFloor  int32
	)

	// Test all of the bounding boxes for a collision with level geometry.
	if ok := result.ScanBoundingBox(GetBoundingRect(d), grid); ok {
		// We've already collided! Try to wiggle free.
		if result.Bottom {
			if !d.Grounded() {
				d.SetGrounded(true)
			} else {
				// result.Bottom = false
			}
		} else {
			d.SetGrounded(false)
		}
		if result.Top {
			// Never seen it touch the top.
		}
		if result.Left {
			P.X++
		}
		if result.Right {
			P.X--
		}
	}

	// If grounded, cap our Y position.
	if d.Grounded() {
		if !result.Bottom {
			// We've fallen off a ledge.
			d.SetGrounded(false)
		} else if target.Y < P.Y {
			// We're moving upward.
			d.SetGrounded(false)
		} else {
			// Cap our downward motion to our current position.
			target.Y = P.Y
		}
	}

	// Cap our horizontal movement if we're touching walls.
	if (result.Left && target.X < P.X) || (result.Right && target.X > P.X) {
		// If the step is short enough, try and jump up.
		height := P.Y + S.H
		if result.Left && target.X < P.X {
			height -= result.LeftPoint.Y
		} else {
			height -= result.RightPoint.Y
		}
		if height <= 8 {
			target.Y -= height
			if target.X < P.X {
				target.X-- // push along to the left
			} else if target.X > P.X {
				target.X++ // push along to the right
			}
		} else {
			target.X = P.X
		}
	}

	// Cap our vertical movement if we're touching ceilings.
	if ceiling {
		// The existing box intersects a ceiling, this will almost never
		// happen because gravity will always pull you away at the last frame.
		// But if we do somehow get here, may as well cap it where it's at.
		capHeight = P.Y
	}

	// Trace a line from where we are to where we wanna go.
	result.Reset()
	result.MoveTo = P
	for point := range render.IterLine2(P, target) {
		if has := result.ScanBoundingBox(render.Rect{
			X: point.X,
			Y: point.Y,
			W: S.W,
			H: S.H,
		}, grid); has {
			if result.Bottom {
				if !hitFloor {
					hitFloor = true
					capFloor = result.BottomPoint.Y - S.H
				}
				d.SetGrounded(true)
			}

			if result.Top && !ceiling {
				// This is a newly discovered ceiling.
				ceiling = true
				capHeight = result.TopPoint.Y
			}

			if result.Left && !hitLeft {
				hitLeft = true
				capLeft = result.LeftPoint.X
			}
			if result.Right && !hitRight {
				hitRight = true
				capRight = result.RightPoint.X - S.W
			}
		}

		// So far so good, keep following the MoveTo to
		// the last good point before a collision.
		result.MoveTo = point

	}

	// If they hit the roof, cap them to the roof.
	if ceiling && result.MoveTo.Y < capHeight {
		result.Top = true
		result.MoveTo.Y = capHeight
	}
	if hitFloor && result.MoveTo.Y > capFloor {
		result.Bottom = true
		result.MoveTo.Y = capFloor
	}
	if hitLeft {
		result.Left = true
		result.MoveTo.X = capLeft
	}
	if hitRight {
		result.Right = true
		result.MoveTo.X = capRight
	}

	return result, result.IsColliding()
}

// IsColliding returns whether any sort of collision has occurred.
func (c *Collide) IsColliding() bool {
	return c.Top || c.Bottom || c.Left || c.Right
}
