package physics_test

import (
	"testing"

	"git.kirsle.net/apps/doodle/pkg/physics"
	"git.kirsle.net/go/render"
)

// Test converting points to vectors and back again.
func TestVectorPoint(t *testing.T) {
	var tests = []struct {
		In  render.Point
		Mid physics.Vector
		Add physics.Vector
		Out render.Point
	}{
		{
			In:  render.NewPoint(102, 102),
			Mid: physics.NewVector(102.0, 102.0),
			Out: render.NewPoint(102, 102),
		},
		{
			In:  render.NewPoint(64, 128),
			Mid: physics.NewVector(64.0, 128.0),
			Add: physics.NewVector(0.4, 0.6),
			Out: render.NewPoint(64, 129),
		},
	}

	for _, test := range tests {
		// Convert point to vector.
		v := physics.VectorFromPoint(test.In)
		if v != test.Mid {
			t.Errorf("Unexpected Vector from Point(%s): wanted %s, got %s",
				test.In, test.Mid, v,
			)
			continue
		}

		// Add other vector.
		v.Add(test.Add)

		// Verify output point rounded down correctly.
		out := v.ToPoint()
		if out != test.Out {
			t.Errorf("Unexpected output vector from Point(%s) -> V(%s) + V(%s): wanted %s, got %s",
				test.In, test.Mid, test.Add, test.Out, out,
			)
			continue
		}
	}
}
