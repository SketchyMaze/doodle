package doodads

import (
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

	// Movement commands.
	MoveBy(render.Point) // Add {X,Y} to current Position.
	MoveTo(render.Point) // Set current Position to {X,Y}.

	// Implement the Draw function.
	Draw(render.Engine)
}

// Collide describes how a collision occurred.
type Collide struct {
	X      int32
	Y      int32
	W      int32
	H      int32
	Top    bool
	Left   bool
	Right  bool
	Bottom bool
}

// CollidesWithGrid checks if a Doodad collides with level geometry.
func CollidesWithGrid(d Doodad, grid *render.Grid) (Collide, bool) {
	var (
		P        = d.Position()
		S        = d.Size()
		topLeft  = P
		topRight = render.Point{
			X: P.X + S.W,
			Y: P.Y,
		}
		bottomLeft = render.Point{
			X: P.X,
			Y: P.Y + S.H,
		}
		bottomRight = render.Point{
			X: bottomLeft.X + S.W,
			Y: P.Y + S.H,
		}
	)

	// Bottom edge.
	for point := range render.IterLine2(bottomLeft, bottomRight) {
		if grid.Exists(render.Pixel{
			X: point.X,
			Y: point.Y,
		}) {
			return Collide{
				Bottom: true,
				X:      point.X,
				Y:      point.Y,
			}, true
		}
	}

	// Top edge.
	for point := range render.IterLine2(topLeft, topRight) {
		if grid.Exists(render.Pixel{
			X: point.X,
			Y: point.Y,
		}) {
			return Collide{
				Top: true,
				X:   point.X,
				Y:   point.Y,
			}, true
		}
	}

	for point := range render.IterLine2(topLeft, bottomLeft) {
		if grid.Exists(render.Pixel{
			X: point.X,
			Y: point.Y,
		}) {
			return Collide{
				Left: true,
				X:    point.X,
				Y:    point.Y,
			}, true
		}
	}

	for point := range render.IterLine2(topRight, bottomRight) {
		if grid.Exists(render.Pixel{
			X: point.X,
			Y: point.Y,
		}) {
			return Collide{
				Right: true,
				X:     point.X,
				Y:     point.Y,
			}, true
		}
	}

	return Collide{}, false
}
