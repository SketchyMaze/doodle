package render

import (
	"fmt"
	"strconv"
	"strings"
)

// Point holds an X,Y coordinate value.
type Point struct {
	X int32
	Y int32
}

// Common points.
var (
	Origin Point
)

// NewPoint makes a new Point at an X,Y coordinate.
func NewPoint(x, y int32) Point {
	return Point{
		X: x,
		Y: y,
	}
}

func (p Point) String() string {
	return fmt.Sprintf("%d,%d", p.X, p.Y)
}

// ParsePoint to parse a point from its string representation.
func ParsePoint(v string) (Point, error) {
	halves := strings.Split(v, ",")
	if len(halves) != 2 {
		return Point{}, fmt.Errorf("'%s': not a valid coordinate string", v)
	}
	x, errX := strconv.Atoi(halves[0])
	y, errY := strconv.Atoi(halves[1])
	if errX != nil || errY != nil {
		return Point{}, fmt.Errorf("invalid coordinate string (X: %v; Y: %v)",
			errX,
			errY,
		)
	}
	return Point{
		X: int32(x),
		Y: int32(y),
	}, nil
}

// IsZero returns if the point is the zero value.
func (p Point) IsZero() bool {
	return p.X == 0 && p.Y == 0
}

// Inside returns whether the Point falls inside the rect.
//
// NOTICE: the W and H are zero-relative, so a 100x100 box at coordinate
// X,Y would still have W,H of 100.
func (p Point) Inside(r Rect) bool {
	var (
		x1 = r.X
		y1 = r.Y
		x2 = r.X + r.W
		y2 = r.Y + r.H
	)
	return ((p.X >= x1 && p.X <= x2) &&
		(p.Y >= y1 && p.Y <= y2))
}

// Add (or subtract) the other point to your current point.
func (p *Point) Add(other Point) {
	p.X += other.X
	p.Y += other.Y
}

// MarshalText to convert the point into text so that a render.Point may be used
// as a map key and serialized to JSON.
func (p *Point) MarshalText() ([]byte, error) {
	return []byte(fmt.Sprintf("%d,%d", p.X, p.Y)), nil
}

// UnmarshalText to restore it from text.
func (p *Point) UnmarshalText(b []byte) error {
	halves := strings.Split(strings.Trim(string(b), `"`), ",")
	if len(halves) != 2 {
		return fmt.Errorf("'%s': not a valid coordinate string", b)
	}

	x, errX := strconv.Atoi(halves[0])
	y, errY := strconv.Atoi(halves[1])
	if errX != nil || errY != nil {
		return fmt.Errorf("Point.UnmarshalJSON: Atoi errors (X=%s Y=%s)",
			errX,
			errY,
		)
	}

	p.X = int32(x)
	p.Y = int32(y)
	return nil
}
