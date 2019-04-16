package collision_test

import (
	"testing"

	"git.kirsle.net/apps/doodle/lib/render"
	"git.kirsle.net/apps/doodle/pkg/collision"
)

func TestActorCollision(t *testing.T) {
	boxes := []render.Rect{
		// 0: intersects with 1
		render.Rect{
			X: 0,
			Y: 0,
			W: 100,
			H: 100,
		},

		// 1: intersects with 0
		render.Rect{
			X: 90,
			Y: 10,
			W: 100,
			H: 100,
		},

		// 2: no intersection
		render.Rect{
			X: 200,
			Y: 200,
			W: 32,
			H: 32,
		},

		// 3: intersects with 4
		render.Rect{
			X: 233,
			Y: 200,
			W: 32,
			H: 32,
		},

		// 4: intersects with 3
		render.Rect{
			X: 240,
			Y: 200,
			W: 32,
			H: 32,
		},

		// 5: completely contains 6 and intersects 7.
		render.Rect{
			X: 300,
			Y: 300,
			W: 1000,
			H: 600,
		},
		render.Rect{
			X: 450,
			Y: 500,
			W: 42,
			H: 42,
		},
		render.Rect{
			X: 1200,
			Y: 350,
			W: 512,
			H: 512,
		},
	}

	assert := func(i int, result collision.IndexTuple, expectA, expectB int) {
		if result[0] != expectA || result[1] != expectB {
			t.Errorf(
				"unexpected collision at index %d of BetweenBoxes() generator\n"+
					"expected: (%d,%d)\n"+
					" but got: (%d,%d)",
				i,
				expectA, expectB,
				result[0], result[1],
			)
		}
	}

	var i int
	for overlap := range collision.BetweenBoxes(boxes) {
		a, b := overlap[0], overlap[1]

		// Ensure expected collisions happened.
		switch i {
		case 0:
			assert(i, overlap, 0, 1)
		case 1:
			assert(i, overlap, 3, 4)
		case 2:
			assert(i, overlap, 5, 6)
		case 3:
			assert(i, overlap, 5, 7)
		default:
			t.Errorf("got unexpected collision result, index %d, tuple (%d,%d)",
				i, a, b,
			)
		}

		i++
	}
}
