package render_test

import (
	"strconv"
	"testing"

	"git.kirsle.net/apps/doodle/lib/render"
)

func TestIntersection(t *testing.T) {
	newRect := func(x, y, w, h int) render.Rect {
		return render.Rect{
			X: int32(x),
			Y: int32(y),
			W: int32(w),
			H: int32(h),
		}
	}

	type TestCase struct {
		A      render.Rect
		B      render.Rect
		Expect bool
	}
	var tests = []TestCase{
		{
			A:      newRect(0, 0, 1000, 1000),
			B:      newRect(200, 200, 100, 100),
			Expect: true,
		},
		{
			A:      newRect(200, 200, 100, 100),
			B:      newRect(0, 0, 1000, 1000),
			Expect: true,
		},
		{
			A:      newRect(0, 0, 100, 100),
			B:      newRect(100, 0, 100, 100),
			Expect: true,
		},
		{
			A:      newRect(0, 0, 99, 99),
			B:      newRect(100, 0, 99, 99),
			Expect: false,
		},
		{
			// Real coords of a test doodad!
			A:      newRect(183, 256, 283, 356),
			B:      newRect(0, -232, 874, 490),
			Expect: true,
		},
		{
			A:      newRect(183, 256, 283, 356),
			B:      newRect(0, -240, 874, 490),
			Expect: false, // XXX: must be true
		},
		{
			A:      newRect(0, 30, 9, 62),
			B:      newRect(16, 0, 32, 64),
			Expect: false,
		},
		{
			A:      newRect(0, 30, 11, 62),
			B:      newRect(7, 4, 17, 28),
			Expect: false,
		},
	}

	for _, test := range tests {
		actual := test.A.Intersects(test.B)
		if actual != test.Expect {
			t.Errorf(
				"%s collision with %s: expected %s, got %s",
				test.A,
				test.B,
				strconv.FormatBool(test.Expect),
				strconv.FormatBool(actual),
			)
		}
	}
}
