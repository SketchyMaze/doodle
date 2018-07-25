package doodads

import (
	"git.kirsle.net/apps/doodle/render"
)

// PlayerID is the Doodad ID for the player character.
const PlayerID = "PLAYER"

// Player is a special doodad for the player character.
type Player struct {
	point    render.Point
	velocity render.Point
	size     render.Rect
	grounded bool
}

// NewPlayer creates the special Player Character doodad.
func NewPlayer() *Player {
	return &Player{
		point: render.Point{
			X: 100,
			Y: 100,
		},
		size: render.Rect{
			W: 32,
			H: 32,
		},
	}
}

// ID of the Player singleton.
func (p *Player) ID() string {
	return PlayerID
}

// Position of the player.
func (p *Player) Position() render.Point {
	return p.point
}

// MoveBy a relative delta position.
func (p *Player) MoveBy(by render.Point) {
	p.point.X += by.X
	p.point.Y += by.Y
}

// MoveTo an absolute position.
func (p *Player) MoveTo(to render.Point) {
	p.point = to
}

// Velocity returns the player's current velocity.
func (p *Player) Velocity() render.Point {
	return p.velocity
}

// Size returns the player's size.
func (p *Player) Size() render.Rect {
	return p.size
}

// Grounded returns if the player is grounded.
func (p *Player) Grounded() bool {
	return p.grounded
}

// SetGrounded sets if the player is grounded.
func (p *Player) SetGrounded(v bool) {
	p.grounded = v
}

// Draw the player sprite.
func (p *Player) Draw(e render.Engine) {
	e.DrawBox(render.Color{255, 255, 153, 255}, render.Rect{
		X: p.point.X,
		Y: p.point.Y,
		W: p.size.W,
		H: p.size.H,
	})
}
