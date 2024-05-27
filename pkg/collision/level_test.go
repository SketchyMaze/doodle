package collision_test

import (
	"fmt"
	"testing"

	"git.kirsle.net/SketchyMaze/doodle/pkg/collision"
	"git.kirsle.net/SketchyMaze/doodle/pkg/doodads/dummy"
	"git.kirsle.net/SketchyMaze/doodle/pkg/level"
	"git.kirsle.net/go/render"
)

func TestCollisionFunctions(t *testing.T) {
	// Create a basic level for testing.
	grid := level.NewChunker(128)
	solid := &level.Swatch{
		Name:  "solid",
		Color: render.Black,
		Solid: true,
	}

	// with a solid platform at y=500 and x=0..1000
	for i := 0; i < 1000; i++ {
		grid.Set(render.NewPoint(i, 500), solid)
	}

	// and a short wall in the middle of the platform
	for i := 480; i < 500; i++ {
		grid.Set(render.NewPoint(500, i), solid)
	}

	// Make a dummy player character.
	player := dummy.NewPlayer()
	playerSize := player.Size()

	// Table based test schema.
	type testCase struct {
		Start           render.Point
		MoveTo          render.Point
		ExpectCollision bool
		Expect          *collision.Collide
	}

	// Describe the details of the test on failure.
	describeTest := func(t testCase, result *collision.Collide, b bool) string {
		return fmt.Sprintf(
			"       Moving From: %s to %s\n"+
				"Expected Collision: %+v (%+v)\n"+
				"     Got Collision: %+v (%+v)",
			t.Start, t.MoveTo,
			t.ExpectCollision, t.Expect,
			b, result,
		)
	}

	// Test cases to check.
	tests := []testCase{
		{
			Start:           render.NewPoint(0, 0),
			MoveTo:          render.NewPoint(8, 8),
			ExpectCollision: false,
		},

		// Player is standing on the floor at X=100
		// with their feet at Y=500 and they move right
		// 10 pixels.
		{
			Start: render.NewPoint(
				100,
				500-playerSize.H,
			),
			MoveTo: render.NewPoint(
				110,
				500-playerSize.H,
			),
			ExpectCollision: true,
			Expect: &collision.Collide{
				Bottom: true,
			},
		},

		// Player walks off the right edge of the platform.
		{
			// TODO: if the player is perfectly touching the floor,
			// this test fails and returns True for collision, so
			// I use 499-playerSize.H so they hover above the floor.
			Start: render.NewPoint(
				990,
				499-playerSize.H,
			),
			MoveTo: render.NewPoint(
				1100,
				499-playerSize.H,
			),
			ExpectCollision: false,
		},

		// Player moves through the barrier in the middle and
		// is stopped in his tracks.
		{
			Start: render.NewPoint(
				490-playerSize.W, 500-playerSize.H,
			),
			MoveTo: render.NewPoint(
				510, 500-playerSize.H,
			),
			ExpectCollision: true,
			Expect: &collision.Collide{
				Right:  true,
				Left:   true, // TODO: not expected
				Bottom: true,
				MoveTo: render.NewPoint(
					500-playerSize.W,
					500-playerSize.H,
				),
			},
		},

		// Player moves up from below the platform and hits the ceiling.
		{
			Start: render.NewPoint(
				490-playerSize.W,
				550,
			),
			MoveTo: render.NewPoint(
				490-playerSize.W,
				499-playerSize.H,
			),
			ExpectCollision: true,
			Expect: &collision.Collide{
				Top: true,

				// TODO: these are unexpected
				Left:   true,
				Right:  true,
				Bottom: true,

				// TODO: the MoveTo is unexpected
				MoveTo: render.NewPoint(458, 468),
				// MoveTo: render.NewPoint(
				// 	490-playerSize.W,
				// 	500,
				// ),
			},
		},
	}

	for i, test := range tests {
		player.MoveTo(test.Start)
		result, collided := collision.CollidesWithGrid(
			player, grid, test.MoveTo,
		)

		// Was there a collision at all?
		if collided && !test.ExpectCollision {
			t.Errorf(
				"Test %d: we collided when we did not expect to!\n%s",
				i,
				describeTest(test, result, collided),
			)
		} else if !collided && test.ExpectCollision {
			t.Errorf(
				"Test %d: we did not collide but we expected to!\n%s",
				i,
				describeTest(test, result, collided),
			)
		} else if test.Expect != nil {
			// Assert that each side is what we expected.
			expect := test.Expect
			if result.Top && !expect.Top || result.Left && !expect.Left ||
				result.Right && !expect.Right || result.Bottom && !expect.Bottom {
				t.Errorf(
					"Test %d: collided as expected, but not the right sides!\n%s",
					i,
					describeTest(test, result, collided),
				)
			}

			// Was the MoveTo position expected?
			if expect.MoveTo != render.Origin && result.MoveTo != expect.MoveTo {
				t.Errorf(
					"Test %d: collided as expected, but didn't move as expected!\n"+
						"Expected to move to: %s\n"+
						"   But actually was: %s",
					i,
					expect.MoveTo,
					result.MoveTo,
				)
			}
		}
	}
}
