package collision

import "git.kirsle.net/go/render"

// MockActor implements the Actor interface for unit testing.
type MockActor struct {
	P  render.Point
	S  render.Rect
	HB render.Rect
	G  bool
}

func (actor *MockActor) Position() render.Point {
	return actor.P
}

func (actor *MockActor) Size() render.Rect {
	return actor.S
}

func (actor *MockActor) Hitbox() render.Rect {
	return actor.HB
}

func (actor *MockActor) Grounded() bool {
	return actor.G
}

func (actor *MockActor) SetGrounded(v bool) {
	actor.G = v
}
