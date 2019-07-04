package render

import (
	"fmt"
	"image"
	"math"

	"git.kirsle.net/apps/doodle/lib/events"
)

// Engine is the interface for the rendering engine, keeping SDL-specific stuff
// far away from the core of Doodle.
type Engine interface {
	Setup() error

	// Poll for events like keypresses and mouse clicks.
	Poll() (*events.State, error)
	GetTicks() uint32
	WindowSize() (w, h int)

	// Present presents the current state to the screen.
	Present() error

	// Clear the full canvas and set this color.
	Clear(Color)
	SetTitle(string)
	DrawPoint(Color, Point)
	DrawLine(Color, Point, Point)
	DrawRect(Color, Rect)
	DrawBox(Color, Rect)
	DrawText(Text, Point) error
	ComputeTextRect(Text) (Rect, error)

	// Texture caching.
	StoreTexture(name string, img image.Image) (Texturer, error)
	LoadTexture(name string) (Texturer, error)
	Copy(t Texturer, src, dst Rect)

	// Delay for a moment using the render engine's delay method,
	// implemented by sdl.Delay(uint32)
	Delay(uint32)

	// Tasks that the Setup function should defer until tear-down.
	Teardown()

	Loop() error // maybe?
}

// Texturer is a stored image texture used by the rendering engine while
// abstracting away its inner workings.
type Texturer interface {
	Size() Rect
}

// Rect has a coordinate and a width and height.
type Rect struct {
	X int32
	Y int32
	W int32
	H int32
}

// NewRect creates a rectangle of size `width` and `height`. The X,Y values
// are initialized to zero.
func NewRect(width, height int32) Rect {
	return Rect{
		W: width,
		H: height,
	}
}

func (r Rect) String() string {
	return fmt.Sprintf("Rect<%d,%d,%d,%d>",
		r.X, r.Y, r.W, r.H,
	)
}

// Point returns the rectangle's X,Y values as a Point.
func (r Rect) Point() Point {
	return Point{
		X: r.X,
		Y: r.Y,
	}
}

// Bigger returns if the given rect is larger than the current one.
func (r Rect) Bigger(other Rect) bool {
	// TODO: don't know why this is !
	return !(other.X < r.X || // Lefter
		other.Y < r.Y || // Higher
		other.W > r.W || // Wider
		other.H > r.H) // Taller
}

// Intersects with the other rectangle in any way.
func (r Rect) Intersects(other Rect) bool {
	// Do a bidirectional compare.
	compare := func(a, b Rect) bool {
		var corners = []Point{
			NewPoint(b.X, b.Y),
			NewPoint(b.X, b.Y+b.H),
			NewPoint(b.X+b.W, b.Y),
			NewPoint(b.X+b.W, b.Y+b.H),
		}
		for _, pt := range corners {
			if pt.Inside(a) {
				return true
			}
		}
		return false
	}

	return compare(r, other) || compare(other, r) || false
}

// IsZero returns if the Rect is uninitialized.
func (r Rect) IsZero() bool {
	return r.X == 0 && r.Y == 0 && r.W == 0 && r.H == 0
}

// Add another rect.
func (r Rect) Add(other Rect) Rect {
	return Rect{
		X: r.X + other.X,
		Y: r.Y + other.Y,
		W: r.W + other.W,
		H: r.H + other.H,
	}
}

// Add a point to move the rect.
func (r Rect) AddPoint(other Point) Rect {
	return Rect{
		X: r.X + other.X,
		Y: r.Y + other.Y,
		W: r.W,
		H: r.H,
	}
}

// SubtractPoint is the inverse of AddPoint. Use this only if you need to invert
// the Point being added.
//
// This does r.X - other.X, r.Y - other.Y and keeps the width/height the same.
func (r Rect) SubtractPoint(other Point) Rect {
	return Rect{
		X: r.X - other.X,
		Y: r.Y - other.Y,
		W: r.W,
		H: r.H,
	}
}

// Text holds information for drawing text.
type Text struct {
	Text         string
	Size         int
	Color        Color
	Padding      int32
	PadX         int32
	PadY         int32
	Stroke       Color  // Stroke color (if not zero)
	Shadow       Color  // Drop shadow color (if not zero)
	FontFilename string // Path to *.ttf file on disk
}

func (t Text) String() string {
	return fmt.Sprintf(`Text<"%s" %dpx %s>`, t.Text, t.Size, t.Color)
}

// IsZero returns if the Text is the zero value.
func (t Text) IsZero() bool {
	return t.Text == "" && t.Size == 0 && t.Color == Invisible && t.Padding == 0 && t.Stroke == Invisible && t.Shadow == Invisible
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

// IterRect loops through all the points forming a rectangle between the
// top-left point and the bottom-right point.
func IterRect(p1, p2 Point) chan Point {
	generator := make(chan Point)

	go func() {
		var (
			TopLeft     = p1
			BottomRight = p2
			TopRight    = Point{
				X: BottomRight.X,
				Y: TopLeft.Y,
			}
			BottomLeft = Point{
				X: TopLeft.X,
				Y: BottomRight.Y,
			}
		)

		// Trace all four edges and yield it.
		var edges = []struct {
			A Point
			B Point
		}{
			{TopLeft, TopRight},
			{TopLeft, BottomLeft},
			{BottomLeft, BottomRight},
			{TopRight, BottomRight},
		}
		for _, edge := range edges {
			for pt := range IterLine2(edge.A, edge.B) {
				generator <- pt
			}
		}

		close(generator)
	}()

	return generator
}
