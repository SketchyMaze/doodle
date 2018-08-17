package render_test

import (
	"strconv"
	"testing"

	"git.kirsle.net/apps/doodle/render"
)

func TestPointInside(t *testing.T) {
	var p = render.Point{
		X: 128,
		Y: 256,
	}

	type testCase struct {
		rect       render.Rect
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
			shouldPass: true,
		},
		testCase{
			rect: render.Rect{
				X: 100,
				Y: 80,
				W: 40,
				H: 60,
			},
			shouldPass: false,
		},
	}

	for _, test := range tests {
		if p.Inside(test.rect) != test.shouldPass {
			t.Errorf("Failed: %s inside %s should %s",
				p,
				test.rect,
				strconv.FormatBool(test.shouldPass),
			)
		}
	}
}
