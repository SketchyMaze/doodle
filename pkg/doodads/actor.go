package doodads

import (
	"git.kirsle.net/apps/doodle/lib/render"
	"git.kirsle.net/apps/doodle/pkg/level"
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

	// Implement the Draw function.
	Draw(render.Engine)
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

// ScanBoundingBox scans all of the pixels in a bounding box on the grid and
// returns if any of them intersect with level geometry.
func (c *Collide) ScanBoundingBox(box render.Rect, grid *level.Chunker) bool {
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
func (c *Collide) ScanGridLine(p1, p2 render.Point, grid *level.Chunker, side Side) {
	for point := range render.IterLine2(p1, p2) {
		if _, err := grid.Get(point); err == nil {
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
