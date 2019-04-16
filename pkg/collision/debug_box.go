package collision

import "git.kirsle.net/apps/doodle/lib/render"

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
