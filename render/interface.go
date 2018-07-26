package render

import (
	"fmt"
	"math"

	"git.kirsle.net/apps/doodle/events"
)

// Engine is the interface for the rendering engine, keeping SDL-specific stuff
// far away from the core of Doodle.
type Engine interface {
	Setup() error

	// Poll for events like keypresses and mouse clicks.
	Poll() (*events.State, error)
	GetTicks() uint32

	// Present presents the current state to the screen.
	Present() error

	// Clear the full canvas and set this color.
	Clear(Color)
	DrawPoint(Color, Point)
	DrawLine(Color, Point, Point)
	DrawRect(Color, Rect)
	DrawBox(Color, Rect)
	DrawText(Text, Point) error
	ComputeTextRect(Text) (Rect, error)

	// Delay for a moment using the render engine's delay method,
	// implemented by sdl.Delay(uint32)
	Delay(uint32)

	// Tasks that the Setup function should defer until tear-down.
	Teardown()

	Loop() error // maybe?
}

// Color holds an RGBA color value.
type Color struct {
	Red   uint8
	Green uint8
	Blue  uint8
	Alpha uint8
}

// RGBA creates a new Color.
func RGBA(r, g, b, a uint8) Color {
	return Color{
		Red:   r,
		Green: g,
		Blue:  b,
		Alpha: a,
	}
}

func (c Color) String() string {
	return fmt.Sprintf(
		"Color<#%02x%02x%02x>",
		c.Red, c.Green, c.Blue,
	)
}

// Add a relative color value to the color.
func (c Color) Add(r, g, b, a int32) Color {
	var (
		R = int32(c.Red) + r
		G = int32(c.Green) + g
		B = int32(c.Blue) + b
		A = int32(c.Alpha) + a
	)

	cap8 := func(v int32) uint8 {
		if v > 255 {
			v = 255
		} else if v < 0 {
			v = 0
		}
		return uint8(v)
	}

	return Color{
		Red:   cap8(R),
		Green: cap8(G),
		Blue:  cap8(B),
		Alpha: cap8(A),
	}
}

// Point holds an X,Y coordinate value.
type Point struct {
	X int32
	Y int32
}

// NewPoint makes a new Point at an X,Y coordinate.
func NewPoint(x, y int32) Point {
	return Point{
		X: x,
		Y: y,
	}
}

func (p Point) String() string {
	return fmt.Sprintf("Point<%d,%d>", p.X, p.Y)
}

// Rect has a coordinate and a width and height.
type Rect struct {
	X int32
	Y int32
	W int32
	H int32
}

func (r Rect) String() string {
	return fmt.Sprintf("Rect<%d,%d,%d,%d>",
		r.X, r.Y, r.W, r.H,
	)
}

// Text holds information for drawing text.
type Text struct {
	Text   string
	Size   int
	Color  Color
	Stroke Color // Stroke color (if not zero)
	Shadow Color // Drop shadow color (if not zero)
}

func (t Text) String() string {
	return fmt.Sprintf("Text<%s>", t.Text)
}

// Common color names.
var (
	Invisible  = Color{}
	White      = RGBA(255, 255, 255, 255)
	Grey       = RGBA(153, 153, 153, 255)
	Black      = RGBA(0, 0, 0, 255)
	SkyBlue    = RGBA(0, 153, 255, 255)
	Blue       = RGBA(0, 0, 255, 255)
	DarkBlue   = RGBA(0, 0, 153, 255)
	Red        = RGBA(255, 0, 0, 255)
	DarkRed    = RGBA(153, 0, 0, 255)
	Green      = RGBA(0, 255, 0, 255)
	DarkGreen  = RGBA(0, 153, 0, 255)
	Cyan       = RGBA(0, 255, 255, 255)
	DarkCyan   = RGBA(0, 153, 153, 255)
	Yellow     = RGBA(255, 255, 0, 255)
	DarkYellow = RGBA(153, 153, 0, 255)
	Magenta    = RGBA(255, 0, 255, 255)
	Purple     = RGBA(153, 0, 153, 255)
	Pink       = RGBA(255, 153, 255, 255)
)

// IterLine is a generator that returns the X,Y coordinates to draw a line.
// https://en.wikipedia.org/wiki/Digital_differential_analyzer_(graphics_algorithm)
func IterLine(x1, y1, x2, y2 int32) chan Point {
	generator := make(chan Point)

	go func() {
		var (
			dx = float64(x2 - x1)
			dy = float64(y2 - y1)
		)
		var step float64
		if math.Abs(dx) >= math.Abs(dy) {
			step = math.Abs(dx)
		} else {
			step = math.Abs(dy)
		}

		dx = dx / step
		dy = dy / step
		x := float64(x1)
		y := float64(y1)
		for i := 0; i <= int(step); i++ {
			generator <- Point{
				X: int32(x),
				Y: int32(y),
			}
			x += dx
			y += dy
		}

		close(generator)
	}()

	return generator
}

// IterLine2 works with two Points rather than four coordinates.
func IterLine2(p1 Point, p2 Point) chan Point {
	return IterLine(p1.X, p1.Y, p2.X, p2.Y)
}
