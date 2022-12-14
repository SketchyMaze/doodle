package draw

import (
	"math"

	"git.kirsle.net/go/render"
)

// Line is a generator that returns the X,Y coordinates to draw a line.
// https://en.wikipedia.org/wiki/Digital_differential_analyzer_(graphics_algorithm)
func Line(x1, y1, x2, y2 int32) chan render.Point {
	generator := make(chan render.Point)

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
			generator <- render.Point{
				X: int(x),
				Y: int(y),
			}
			x += dx
			y += dy
		}

		close(generator)
	}()

	return generator
}
