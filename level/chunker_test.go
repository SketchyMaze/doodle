package level_test

import (
	"testing"

	"git.kirsle.net/apps/doodle/level"
	"git.kirsle.net/apps/doodle/render"
)

func TestWorldSize(t *testing.T) {
	type TestCase struct {
		Size   int
		Points []render.Point
		Expect render.Rect
		Zero   render.Rect // expected WorldSizePositive
	}
	var tests = []TestCase{
		{
			Size: 1000,
			Points: []render.Point{
				render.NewPoint(0, 0),       // chunk 0,0
				render.NewPoint(512, 788),   // 0,0
				render.NewPoint(1002, 500),  // 1,0
				render.NewPoint(2005, 2006), // 2,2
				render.NewPoint(-5, -5),     // -1,-1
			},
			Expect: render.Rect{
				X: -1000,
				Y: -1000,
				W: 2999,
				H: 2999,
			},
			Zero: render.NewRect(3999, 3999),
		},
		{
			Size: 128,
			Points: []render.Point{
				render.NewPoint(5, 5),
			},
			Expect: render.Rect{
				X: 0,
				Y: 0,
				W: 127,
				H: 127,
			},
			Zero: render.NewRect(127, 127),
		},
		{
			Size: 200,
			Points: []render.Point{
				render.NewPoint(-6000, -38556),
				render.NewPoint(12345, 1288000),
			},
			Expect: render.Rect{
				X: -6000,
				Y: -38600,
				W: 12399,
				H: 1288199,
			},
			Zero: render.NewRect(18399, 1326799),
		},
	}
	for _, test := range tests {
		c := level.NewChunker(test.Size)
		sw := &level.Swatch{
			Name:  "solid",
			Color: render.Black,
		}

		for _, pt := range test.Points {
			c.Set(pt, sw)
		}

		size := c.WorldSize()
		if size != test.Expect {
			t.Errorf("WorldSize not as expected: %s <> %s", size, test.Expect)
		}

		zero := c.WorldSizePositive()
		if zero != test.Zero {
			t.Errorf("WorldSizePositive not as expected: %s <> %s", zero, test.Expect)
		}
	}

}
