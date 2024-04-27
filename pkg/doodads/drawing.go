package doodads

import (
	"git.kirsle.net/go/render"
	"github.com/google/uuid"
)

// Drawing is a Doodad Actor that is based on drawings made inside the game.
type Drawing struct {
	Doodad *Doodad

	id       string
	point    render.Point
	velocity render.Point
	accel    int
	size     render.Rect
	hitbox   render.Rect
	grounded bool
}

// NewDrawing creates a Drawing actor based on a Doodad drawing. If you pass
// an empty ID string, it will make a random UUIDv4 ID.
func NewDrawing(id string, doodad *Doodad) *Drawing {
	if id == "" {
		id = uuid.Must(uuid.NewUUID()).String()
	}
	return &Drawing{
		id:     id,
		Doodad: doodad,
		size:   doodad.Rect(),
	}
}

// ID to get the Drawing ID.
func (d *Drawing) ID() string {
	return d.id
}

// Position returns the Drawing's position.
func (d *Drawing) Position() render.Point {
	return d.point
}

// Size returns the Drawing's size.
func (d *Drawing) Size() render.Rect {
	return d.size
}

// MoveTo an absolute world value.
//
// NOTE: used only by unit test.
func (d *Drawing) MoveTo(to render.Point) {
	d.point = to
}
