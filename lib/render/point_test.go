package render_test

import (
	"strconv"
	"testing"

	"git.kirsle.net/apps/doodle/lib/render"
)

func TestPointInside(t *testing.T) {
	type testCase struct {
		rect       render.Rect
		p          render.Point
		shouldPass bool
	}
	tests := []testCase{
		testCase{
			rect: render.Rect{
				X: 0,
				Y: 0,
				W: 500,
				H: 500,
			},
			p:          render.NewPoint(128, 256),
			shouldPass: true,
		},
		testCase{
			rect: render.Rect{
				X: 100,
				Y: 80,
				W: 40,
				H: 60,
			},
			p:          render.NewPoint(128, 256),
			shouldPass: false,
		},
		testCase{
			// true values when debugging why Doodads weren't
			// considered inside the viewport.
			rect: render.Rect{
				X: 0,
				Y: -232,
				H: 874,
				W: 490,
			},
			p:          render.NewPoint(509, 260),
			shouldPass: false,
		},
	}

	for _, test := range tests {
		if test.p.Inside(test.rect) != test.shouldPass {
			t.Errorf("Failed: %s inside %s should be %s",
				test.p,
				test.rect,
				strconv.FormatBool(test.shouldPass),
			)
		}
	}
}
