package uix

import "git.kirsle.net/go/render"

// CollideEvent holds data sent to an actor's Collide handler.
type CollideEvent struct {
	Actor    *Actor
	Overlap  render.Rect
	InHitbox bool // If the two elected hitboxes are overlapping
	Settled  bool // Movement phase finished, actor script can fire actions
}
