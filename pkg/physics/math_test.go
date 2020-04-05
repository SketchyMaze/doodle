package physics_test

import (
	"testing"

	"git.kirsle.net/apps/doodle/pkg/physics"
)

func TestLerp(t *testing.T) {
	var tests = []struct {
		Inputs []float64
		Expect float64
	}{
		{
			Inputs: []float64{0, 1, 0.75},
			Expect: 0.75,
		},
		{
			Inputs: []float64{0, 100, 0.5},
			Expect: 50.0,
		},
		{
			Inputs: []float64{10, 75, 0.3},
			Expect: 29.5,
		},
		{
			Inputs: []float64{30, 2, 0.75},
			Expect: 9,
		},
	}

	for _, test := range tests {
		result := physics.Lerp(test.Inputs[0], test.Inputs[1], test.Inputs[2])
		if result != test.Expect {
			t.Errorf("Lerp(%+v) expected %f but got %f", test.Inputs, test.Expect, result)
		}
	}
}
