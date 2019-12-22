package doodads

import (
	"git.kirsle.net/apps/doodle/lib/render"
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
func NewDrawing(id string, doodad *Doodad) Drawing {
	if id == "" {
		id = uuid.Must(uuid.NewRandom()).String()
	}
	return Drawing{
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

// Velocity returns the Drawing's velocity.
func (d *Drawing) Velocity() render.Point {
	return d.velocity
}

// SetVelocity to set the speed.
func (d *Drawing) SetVelocity(v render.Point) {
	d.velocity = v
}

// Acceleration returns the Drawing's acceleration.
func (d *Drawing) Acceleration() int {
	return d.accel
}

// SetAcceleration to set the acceleration.
func (d *Drawing) SetAcceleration(v int) {
	d.accel = v
}

// Size returns the Drawing's size.
func (d *Drawing) Size() render.Rect {
	return d.size
}

// Grounded returns whether the Drawing is standing on solid ground.
func (d *Drawing) Grounded() bool {
	return d.grounded
}

// SetGrounded sets the grounded state.
func (d *Drawing) SetGrounded(v bool) {
	d.grounded = v
}

// // SetHitbox sets the actor's elected hitbox.
// func (d *Drawing) SetHitbox(x, y, w, h int) {
// 	d.hitbox = render.Rect{
// 		X: int32(x),
// 		Y: int32(y),
// 		W: int32(w),
// 		H: int32(h),
// 	}
// }
//
// // Hitbox returns the actor's elected hitbox.
// func (d *Drawing) Hitbox() render.Rect {
// 	return d.hitbox
// }

// MoveBy a relative value.
func (d *Drawing) MoveBy(by render.Point) {
	d.point.Add(by)
}

// MoveTo an absolute world value.
func (d *Drawing) MoveTo(to render.Point) {
	d.point = to
}

// Draw the drawing.
func (d *Drawing) Draw(e render.Engine) {

}
