package draw_test

import (
	"fmt"
	"testing"

	"git.kirsle.net/SketchyMaze/doodle/lib/draw"
	"git.kirsle.net/go/render"
)

func TestLine(t *testing.T) {
	type task struct {
		X1     int32
		X2     int32
		Y1     int32
		Y2     int32
		Expect []render.Point
	}
	toString := func(t task) string {
		return fmt.Sprintf("Line<%d,%d -> %d,%d>",
			t.X1, t.Y1,
			t.X2, t.Y2,
		)
	}

	var tasks = []task{
		task{
			X1: 0,
			Y1: 0,
			X2: 0,
			Y2: 10,
			Expect: []render.Point{
				{X: 0, Y: 0},
				{X: 0, Y: 1},
				{X: 0, Y: 2},
				{X: 0, Y: 3},
				{X: 0, Y: 4},
				{X: 0, Y: 5},
				{X: 0, Y: 6},
				{X: 0, Y: 7},
				{X: 0, Y: 8},
				{X: 0, Y: 9},
				{X: 0, Y: 10},
			},
		},
		task{
			X1: 10,
			Y1: 10,
			X2: 15,
			Y2: 15,
			Expect: []render.Point{
				{X: 10, Y: 10},
				{X: 11, Y: 11},
				{X: 12, Y: 12},
				{X: 13, Y: 13},
				{X: 14, Y: 14},
				{X: 15, Y: 15},
			},
		},
	}
	for _, test := range tasks {
		gen := draw.Line(test.X1, test.Y1, test.X2, test.Y2)
		var i int
		for point := range gen {
			if i >= len(test.Expect) {
				t.Errorf("%s: Got more pixels back than expected: %s",
					toString(test),
					point,
				)
				break
			}

			expect := test.Expect[i]
			if expect != point {
				t.Errorf("%s: at index %d I got %s but expected %s",
					toString(test),
					i,
					point,
					expect,
				)
			}

			i++
		}
	}
}
