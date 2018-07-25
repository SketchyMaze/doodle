package doodads

import (
	"fmt"

	"git.kirsle.net/apps/doodle/level"
	"git.kirsle.net/apps/doodle/render"
)

// Doodad is a reusable drawing component used in Doodle. Doodads are buttons,
// doors, switches, the player characters themselves, anything that isn't a part
// of the level geometry.
type Doodad interface {
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

	// Implement the Draw function.
	Draw(render.Engine)
}

// Collide describes how a collision occurred.
type Collide struct {
	Top         bool
	TopPoint    render.Point
	Left        bool
	LeftPoint   render.Point
	Right       bool
	RightPoint  render.Point
	Bottom      bool
	BottomPoint render.Point
	MoveTo      render.Point
}

// CollisionBox holds all of the coordinate pairs to draw the collision box
// around a doodad.
type CollisionBox struct {
	Top    []render.Point
	Bottom []render.Point
	Left   []render.Point
	Right  []render.Point
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
func CollidesWithGrid(d Doodad, grid *render.Grid, target render.Point) (*Collide, bool) {
	var (
		P = d.Position()
		S = d.Size()

		result = &Collide{
			MoveTo: P,
		}
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
			P.Y++
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
		relPoint := P.Y + S.H
		if result.Left && target.X < P.X {
			relPoint -= result.LeftPoint.Y
		} else {
			relPoint -= result.RightPoint.Y
		}
		fmt.Printf("Touched a wall at %d pixels height (P=%s)\n", relPoint, P)
		if S.H-relPoint > S.H-8 {
			target.Y -= 12
			if target.X < P.X {
				target.X-- // push along to the left
			} else if target.X > P.X {
				target.X++ // push along to the right
			}
		} else {
			target.X = P.X
		}
	}

	// Trace a line from where we are to where we wanna go.
	result.MoveTo = P
	for point := range render.IterLine2(P, target) {
		if ok := result.ScanBoundingBox(render.Rect{
			X: point.X,
			Y: point.Y,
			W: S.W,
			H: S.H,
		}, grid); ok {
			if d.Grounded() {
				if !result.Bottom {
					d.SetGrounded(false)
				}
			} else if result.Bottom {
				d.SetGrounded(true)
			}
		}
		result.MoveTo = point
	}

	return result, result.IsColliding()
}

// IsColliding returns whether any sort of collision has occurred.
func (c *Collide) IsColliding() bool {
	return c.Top || c.Bottom || c.Left || c.Right
}

// GetCollisionBox computes the full pairs of points for the collision box
// around a doodad.
func GetBoundingRect(d Doodad) render.Rect {
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
				Y: box.Y + 1,
			},
			{
				X: box.X,
				Y: box.Y + box.H - 1,
			},
		},
		Right: []render.Point{
			{
				X: box.X + box.W,
				Y: box.Y + 1,
			},
			{
				X: box.X + box.W,
				Y: box.Y + box.H - 1,
			},
		},
	}
}

// ScanBoundingBox scans all of the pixels in a bounding box on the grid and
// returns if any of them intersect with level geometry.
func (c *Collide) ScanBoundingBox(box render.Rect, grid *render.Grid) bool {
	col := GetCollisionBox(box)

	c.ScanGridLine(col.Top[0], col.Top[1], grid, Top)
	c.ScanGridLine(col.Bottom[0], col.Bottom[1], grid, Bottom)
	c.ScanGridLine(col.Left[0], col.Left[1], grid, Left)
	c.ScanGridLine(col.Right[0], col.Right[1], grid, Right)
	return c.IsColliding()
}

// ScanGridLine scans all of the pixels between p1 and p2 on the grid and tests
// for any pixels to be set, implying a collision between level geometry and the
// bounding boxes of the doodad.
func (c *Collide) ScanGridLine(p1, p2 render.Point, grid *render.Grid, side Side) {
	for point := range render.IterLine2(p1, p2) {
		if grid.Exists(level.Pixel{
			X: point.X,
			Y: point.Y,
		}) {
			// A hit!
			switch side {
			case Top:
				c.Top = true
				c.TopPoint = point
			case Bottom:
				c.Bottom = true
				c.BottomPoint = point
			case Left:
				c.Left = true
				c.LeftPoint = point
			case Right:
				c.Right = true
				c.RightPoint = point
			}
		}
	}
}
