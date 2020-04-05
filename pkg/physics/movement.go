package physics

// Mover is a moving object.
type Mover struct {
	Acceleration float64
	Friction     float64
	// Gravity      Vector

	// // Position and previous frame's position.
	// Position    render.Point
	// OldPosition render.Point
	//
	// // Speed and previous frame's speed.
	// Speed    render.Point
	// OldSpeed render.Point
	MaxSpeed Vector
	//
	// // Object is on the ground and its grounded state last frame.
	// Grounded    bool
	// WasGrounded bool
}

// NewMover initializes state for a moving object.
func NewMover() *Mover {
	return &Mover{}
}

// // UpdatePhysics runs calculations on the mover's physics each frame.
// func (m *Mover) UpdatePhysics() {
// 	m.OldPosition = m.Position
// 	m.OldSpeed = m.Speed
// 	m.WasGrounded = m.Grounded
// }
