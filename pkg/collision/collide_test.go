package collision_test

import (
	"testing"

	"git.kirsle.net/apps/doodle/pkg/collision"
	"git.kirsle.net/go/render"
)

func TestBetweenBoxes(t *testing.T) {
	mkrect := func(x, y, w, h int) render.Rect {
		return render.Rect{
			X: x,
			Y: y,
			W: w,
			H: h,
		}
	}
	table := []struct {
		A      render.Rect
		B      render.Rect
		Expect bool
	}{
		{
			A:      mkrect(0, 0, 32, 32),
			B:      mkrect(32, 0, 32, 32),
			Expect: true,
		},
		{
			A:      mkrect(0, 0, 32, 32),
			B:      mkrect(100, 100, 40, 40),
			Expect: false,
		},
		{
			A:      mkrect(100, 100, 50, 50),
			B:      mkrect(80, 110, 100, 30),
			Expect: true,
		},
	}

	var actual bool
	for i, test := range table {
		actual = false
		for range collision.BetweenBoxes([]render.Rect{test.A, test.B}) {
			actual = true
			break
		}

		if test.Expect != actual {
			t.Errorf(
				"Test %d BetweenBoxes: %s cmp %s\n"+
					"Expected: %+v\n"+
					"Actually: %+v",
				i, test.A, test.B, test.Expect, actual,
			)
		}
	}
}
