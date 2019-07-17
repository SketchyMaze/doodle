package render

// MidpointEllipse implements an ellipse plotting algorithm.
func MidpointEllipse(center, radius Point) chan Point {
	yield := make(chan Point)
	go func() {

		var (
			pos   = NewPoint(radius.X, 0)
			delta = NewPoint(
				2*radius.Y*radius.Y*pos.X,
				2*radius.X*radius.X*pos.Y,
			)
			err = radius.X*radius.X -
				radius.Y*radius.Y*radius.X +
				(radius.Y*radius.Y)/4
		)

		for delta.Y < delta.X {
			yield <- NewPoint(center.X+pos.X, center.Y+pos.Y)
			yield <- NewPoint(center.X+pos.X, center.Y-pos.Y)
			yield <- NewPoint(center.X-pos.X, center.Y+pos.Y)
			yield <- NewPoint(center.X-pos.X, center.Y-pos.Y)

			pos.Y++

			if err < 0 {
				delta.Y += 2 * radius.X * radius.X
				err += delta.Y + radius.X*radius.X
			} else {
				pos.X--
				delta.Y += 2 * radius.X * radius.X
				delta.X -= 2 * radius.Y * radius.Y
				err += delta.Y - delta.X + radius.X*radius.X
			}
		}

		err = radius.X*radius.X*(pos.Y*pos.Y+pos.Y) +
			radius.Y*radius.Y*(pos.X-1)*(pos.X-1) -
			radius.Y*radius.Y*radius.X*radius.X

		for pos.X >= 0 {
			yield <- NewPoint(center.X+pos.X, center.Y+pos.Y)
			yield <- NewPoint(center.X+pos.X, center.Y-pos.Y)
			yield <- NewPoint(center.X-pos.X, center.Y+pos.Y)
			yield <- NewPoint(center.X-pos.X, center.Y-pos.Y)

			pos.X--

			if err > 0 {
				delta.X -= 2 * radius.Y * radius.Y
				err += radius.Y*radius.Y - delta.X
			} else {
				pos.Y++
				delta.Y += 2 * radius.X * radius.X
				delta.X -= 2 * radius.Y * radius.Y
				err += delta.Y - delta.X + radius.Y*radius.Y
			}
		}

		close(yield)
	}()
	return yield
}
