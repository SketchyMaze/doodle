package physics

import (
	"fmt"
	"math"

	"git.kirsle.net/go/render"
)

// Vector holds floating point values on an X and Y coordinate.
type Vector struct {
	X float64
	Y float64
}

// NewVector creates a Vector from X and Y values.
func NewVector(x, y float64) Vector {
	return Vector{
		X: x,
		Y: y,
	}
}

// VectorFromPoint converts a render.Point into a vector.
func VectorFromPoint(p render.Point) Vector {
	return Vector{
		X: float64(p.X),
		Y: float64(p.Y),
	}
}

// IsZero returns if the vector is zero.
func (v Vector) IsZero() bool {
	return v.X == 0 && v.Y == 0
}

// Add to the vector.
func (v *Vector) Add(other Vector) {
	v.X += other.X
	v.Y += other.Y
}

// ToPoint converts the vector into a render.Point with integer coordinates.
func (v Vector) ToPoint() render.Point {
	return render.Point{
		X: int(math.Round(v.X)),
		Y: int(math.Round(v.Y)),
	}
}

// String encoding of the vector.
func (v Vector) String() string {
	return fmt.Sprintf("%f,%f", v.X, v.Y)
}
