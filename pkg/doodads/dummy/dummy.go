// Package dummy implements a dummy doodads.Drawing.
package dummy

import (
	"git.kirsle.net/apps/doodle/pkg/doodads"
	"git.kirsle.net/go/render"
)

// Drawing is a dummy doodads.Drawing that has no data.
type Drawing struct {
	Drawing *doodads.Drawing
}

// NewDrawing creates a new dummy drawing.
func NewDrawing(id string, doodad *doodads.Doodad) *Drawing {
	return &Drawing{
		Drawing: doodads.NewDrawing(id, doodad),
	}
}

// Size returns the size of the underlying doodads.Drawing.
func (d *Drawing) Size() render.Rect {
	return d.Drawing.Size()
}

// MoveTo changes the drawing's position.
func (d *Drawing) MoveTo(to render.Point) {
	d.Drawing.MoveTo(to)
}

// Grounded satisfies the collision.Actor interface.
func (d *Drawing) Grounded() bool {
	return false
}

// SetGrounded satisfies the collision.Actor interface.
func (d *Drawing) SetGrounded(v bool) {}

// Position satisfies the collision.Actor interface.
func (d *Drawing) Position() render.Point {
	return render.Point{}
}

// Hitbox satisfies the collision.Actor interface.
func (d *Drawing) Hitbox() render.Rect {
	return render.Rect{}
}
