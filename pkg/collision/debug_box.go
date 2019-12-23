package collision

import (
	"fmt"

	"git.kirsle.net/go/render"
)

// CollisionBox holds all of the coordinate pairs to draw the collision box
// around a doodad.
type CollisionBox struct {
	Top    [2]render.Point
	Bottom [2]render.Point
	Left   [2]render.Point
	Right  [2]render.Point
}

// NewBox creates a collision box from the Top Left and Bottom Right points.
func NewBox(topLeft, bottomRight render.Point) CollisionBox {
	return GetCollisionBox(render.Rect{
		X: topLeft.X,
		Y: topLeft.Y,
		W: bottomRight.X - topLeft.X,
		H: bottomRight.Y - topLeft.Y,
	})
}

// GetCollisionBox returns a CollisionBox with the four coordinates.
func GetCollisionBox(box render.Rect) CollisionBox {
	return CollisionBox{
		Top: [2]render.Point{
			{
				X: box.X,
				Y: box.Y,
			},
			{
				X: box.X + box.W,
				Y: box.Y,
			},
		},
		Bottom: [2]render.Point{
			{
				X: box.X,
				Y: box.Y + box.H,
			},
			{
				X: box.X + box.W,
				Y: box.Y + box.H,
			},
		},
		Left: [2]render.Point{
			{
				X: box.X,
				Y: box.Y + box.H - 1,
			},
			{
				X: box.X,
				Y: box.Y + 1,
			},
		},
		Right: [2]render.Point{
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

// String prints the bounds of the collision box in absolute terms.
func (c CollisionBox) String() string {
	return fmt.Sprintf("CollisionBox<%s:%s>",
		c.TopLeft(),
		c.BottomRight(),
	)
}

// TopLeft returns the point at the top left.
func (c CollisionBox) TopLeft() render.Point {
	return render.NewPoint(c.Top[0].X, c.Top[0].Y)
}

// BottomRight returns the point at the bottom right.
func (c CollisionBox) BottomRight() render.Point {
	return render.NewPoint(c.Bottom[1].X, c.Bottom[1].Y)
}
