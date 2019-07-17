package render

import (
	"math"
)

// IterLine is a generator that returns the X,Y coordinates to draw a line.
// https://en.wikipedia.org/wiki/Digital_differential_analyzer_(graphics_algorithm)
func IterLine(p1 Point, p2 Point) chan Point {
	var (
		x1 = p1.X
		y1 = p1.Y
		x2 = p2.X
		y2 = p2.Y
	)
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
			dedupe = map[Point]interface{}{}
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
			for pt := range IterLine(edge.A, edge.B) {
				if _, ok := dedupe[pt]; !ok {
					generator <- pt
					dedupe[pt] = nil
				}
			}
		}

		close(generator)
	}()

	return generator
}

// IterEllipse iterates an Ellipse using two Points as the top-left and
// bottom-right corners of a rectangle that encompasses the ellipse.
func IterEllipse(A, B Point) chan Point {
	var (
		width  = AbsInt32(B.X - A.X)
		height = AbsInt32(B.Y - A.Y)
		radius = NewPoint(width/2, height/2)
		center = NewPoint(AbsInt32(B.X-radius.X), AbsInt32(B.Y-radius.Y))
	)

	return MidpointEllipse(center, radius)
}
