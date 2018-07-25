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

func (c Color) String() string {
	return fmt.Sprintf(
		"Color<#%02x%02x%02x>",
		c.Red, c.Green, c.Blue,
	)
}

// Point holds an X,Y coordinate value.
type Point struct {
	X int32
	Y int32
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
	Invisible = Color{}
	White     = Color{255, 255, 255, 255}
	Grey      = Color{153, 153, 153, 255}
	Black     = Color{0, 0, 0, 255}
	SkyBlue   = Color{0, 153, 255, 255}
	Blue      = Color{0, 0, 255, 255}
	Red       = Color{255, 0, 0, 255}
	Green     = Color{0, 255, 0, 255}
	Cyan      = Color{0, 255, 255, 255}
	Yellow    = Color{255, 255, 0, 255}
	Magenta   = Color{255, 0, 255, 255}
	Pink      = Color{255, 153, 255, 255}
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
