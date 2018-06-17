package types

import "fmt"

// Point is a 2D point in space.
type Point struct {
	X int32 `json:"x"`
	Y int32 `json:"y"`
}

func (p Point) String() string {
	return fmt.Sprintf("(%d,%d)", p.X, p.Y)
}
