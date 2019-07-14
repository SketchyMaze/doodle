package render

import (
	"math"

	"git.kirsle.net/apps/doodle/pkg/log"
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

// IterEllipse is a generator that draws out the pixels of an ellipse.
func IterEllipse(rx, ry, xc, yc float32) chan Point {
	generator := make(chan Point)

	mkPoint := func(x, y float32) Point {
		return NewPoint(int32(x), int32(y))
	}

	go func() {
		var (
			dx float32
			dy float32
			d1 float32
			d2 float32
			x  float32
			y  = ry
		)

		d1 = (ry * ry) - (rx * rx * ry) + (0.25 * rx * rx)
		dx = 2 * ry * ry * x
		dy = 2 * rx * rx * y

		// For region 1
		for dx < dy {
			// Yields points based on 4-way symmetry.
			for _, point := range []Point{
				mkPoint(x+xc, y+yc),
				mkPoint(-x+xc, y+yc),
				mkPoint(x+xc, -y+yc),
				mkPoint(-x+xc, -y+yc),
			} {
				generator <- point
			}

			if d1 < 0 {
				x++
				dx = dx + (2 * ry * ry)
				d1 = d1 + dx + (ry * ry)
			} else {
				x++
				y--
				dx = dx + (2 * ry * ry)
				dy = dy - (2 * rx * rx)
				d1 = d1 + dx - dy + (ry * ry)
			}
		}

		d2 = ((ry * ry) + ((x + 0.5) * (x + 0.5))) +
			((rx * rx) * ((y - 1) * (y - 1))) -
			(rx * rx * ry * ry)

		// Region 2
		for y >= 0 {
			// Yields points based on 4-way symmetry.
			for _, point := range []Point{
				mkPoint(x+xc, y+yc),
				mkPoint(-x+xc, y+yc),
				mkPoint(x+xc, -y+yc),
				mkPoint(-x+xc, -y+yc),
			} {
				generator <- point
			}

			if d2 > 0 {
				y--
				dy = dy - (2 * rx * rx)
				d2 = d2 + (rx * rx) - dy
			} else {
				y--
				x++
				dx = dx + (2 * ry * ry)
				dy = dy - (2 * rx * rx)
				d2 = d2 + dx - dy + (rx * rx)
			}
		}

		close(generator)
	}()

	return generator
}

// IterEllipse2 iterates an Ellipse using two Points as the top-left and
// bottom-right corners of a rectangle that encompasses the ellipse.
func IterEllipse2(A, B Point) chan Point {
	var (
		// xc = float32(A.X+B.X) / 2
		// yc = float32(A.Y+B.Y) / 2
		xc = float32(B.X)
		yc = float32(B.Y)
		rx = float32(B.X - A.X)
		ry = float32(B.Y - A.Y)
	)

	if rx < 0 {
		rx = -rx
	}
	if ry < 0 {
		ry = -ry
	}
	log.Info("Ellipse btwn=%s-%s  radius=%f,%f at center %f,%f", A, B, rx, ry, xc, yc)
	return IterEllipse(rx, ry, xc, yc)
}
