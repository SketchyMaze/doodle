package uix

import "git.kirsle.net/apps/doodle/lib/render"

// CollideEvent holds data sent to an actor's Collide handler.
type CollideEvent struct {
	Actor   *Actor
	Overlap render.Rect
}
