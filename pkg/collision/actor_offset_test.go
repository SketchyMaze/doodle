package collision_test

import (
	"testing"

	"git.kirsle.net/SketchyMaze/doodle/pkg/collision"
	"git.kirsle.net/go/render"
)

func TestActorOffset(t *testing.T) {
	type testCase struct {
		Actor       *collision.MockActor
		Offset      render.Point
		ExpectPoint render.Point
	}

	var tests = []testCase{
		// Simple case where the hitbox == the size.
		{
			Actor: &collision.MockActor{
				P:  render.NewPoint(10, 10),
				S:  render.NewRect(32, 32),
				HB: render.NewRect(32, 32),
			},
			ExpectPoint: render.NewPoint(10, 10),
		},

		// Bottom heavy actor
		{
			Actor: &collision.MockActor{
				P: render.NewPoint(11, 22),
				S: render.NewRect(32, 64),
				HB: render.Rect{
					X: 0,
					Y: 32,
					W: 32,
					H: 32,
				},
			},
			ExpectPoint: render.NewPoint(11, 22+32),
		},
	}

	for i, test := range tests {
		offset := collision.NewActorOffset(test.Actor)

		actualPoint := offset.Position()
		if actualPoint != test.ExpectPoint {
			t.Errorf("Test #%d: Position() expected to be %s but was %s",
				i,
				test.ExpectPoint,
				actualPoint,
			)
		}
	}
}
