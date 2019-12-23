package level_test

import (
	"fmt"
	"testing"

	"git.kirsle.net/go/render"
	"git.kirsle.net/apps/doodle/pkg/level"
)

// Test the high level Chunker.
func TestChunker(t *testing.T) {
	c := level.NewChunker(1000)

	// Test swatches.
	var (
		grey = &level.Swatch{
			Name:  "solid",
			Color: render.Grey,
		}
	)

	type testCase struct {
		name string
		run  func() error
	}
	tests := []testCase{
		testCase{
			name: "Access a pixel on the blank map and expect an error",
			run: func() error {
				p := render.NewPoint(65535, -214564545)
				_, err := c.Get(p)
				if err == nil {
					return fmt.Errorf("unexpected success getting point %s", p)
				}
				return nil
			},
		},

		testCase{
			name: "Set a pixel",
			run: func() error {
				// Set a point.
				p := render.NewPoint(100, 200)
				err := c.Set(p, grey)
				if err != nil {
					return fmt.Errorf("unexpected error getting point %s: %s", p, err)
				}
				return nil
			},
		},

		testCase{
			name: "Verify the set pixel",
			run: func() error {
				p := render.NewPoint(100, 200)
				px, err := c.Get(p)
				if err != nil {
					return err
				}
				if px != grey {
					return fmt.Errorf("pixel at %s not the expected color:\n"+
						"Expected: %s\n"+
						"     Got: %s",
						p,
						grey,
						px,
					)
				}
				return nil
			},
		},

		testCase{
			name: "Verify the neighboring pixel is unset",
			run: func() error {
				p := render.NewPoint(101, 200)
				_, err := c.Get(p)
				if err == nil {
					return fmt.Errorf("unexpected success getting point %s", p)
				}
				return nil
			},
		},

		testCase{
			name: "Delete the set pixel",
			run: func() error {
				p := render.NewPoint(100, 200)
				err := c.Delete(p)
				if err != nil {
					return err
				}
				return nil
			},
		},

		testCase{
			name: "Verify the deleted pixel is unset",
			run: func() error {
				p := render.NewPoint(100, 200)
				_, err := c.Get(p)
				if err == nil {
					return fmt.Errorf("unexpected success getting point %s", p)
				}
				return nil
			},
		},

		testCase{
			name: "Delete a pixel that didn't exist",
			run: func() error {
				p := render.NewPoint(-100, -100)
				err := c.Delete(p)
				if err == nil {
					return fmt.Errorf("unexpected success deleting point %s", p)
				}
				return nil
			},
		},
	}

	for _, test := range tests {
		if err := test.run(); err != nil {
			t.Errorf("Failed: %s\n%s", test.name, err)
		}
	}
}

// Test the map chunk accessor.
func TestMapAccessor(t *testing.T) {
	a := level.NewMapAccessor()
	_ = a

	// Test action types
	var (
		Get    = "Get"
		Set    = "Set"
		Delete = "Delete"
	)

	// Test swatches.
	var (
		red = &level.Swatch{
			Name:  "fire",
			Color: render.Red,
		}
	)

	type testCase struct {
		Action string
		P      render.Point
		S      *level.Swatch
		Expect *level.Swatch
		Err    bool // expect error
	}
	tests := []testCase{
		// Get a random point and expect to fail.
		testCase{
			Action: Get,
			P:      render.NewPoint(128, 128),
			Err:    true,
		},

		// Set a point.
		testCase{
			Action: Set,
			S:      red,
			P:      render.NewPoint(1024, 2048),
		},

		// Verify it exists.
		testCase{
			Action: Get,
			P:      render.NewPoint(1024, 2048),
			Expect: red,
		},

		// A neighboring point does not exist.
		testCase{
			Action: Get,
			P:      render.NewPoint(1025, 2050),
			Err:    true,
		},

		// Delete a pixel that doesn't exist.
		testCase{
			Action: Delete,
			P:      render.NewPoint(1987, 2006),
			Err:    true,
		},

		// Delete one that does.
		testCase{
			Action: Delete,
			P:      render.NewPoint(1024, 2048),
		},

		// Verify gone
		testCase{
			Action: Get,
			P:      render.NewPoint(1024, 2048),
			Err:    true,
		},
	}

	for _, test := range tests {
		var px *level.Swatch
		var err error
		switch test.Action {
		case Get:
			px, err = a.Get(test.P)
		case Set:
			err = a.Set(test.P, test.S)
		case Delete:
			err = a.Delete(test.P)
		}

		if err != nil && !test.Err {
			t.Errorf("unexpected error from %s %s: %s", test.Action, test.P, err)
			continue
		} else if err == nil && test.Err {
			t.Errorf("didn't get error when we expected from %s %s", test.Action, test.P)
			continue
		}

		if test.Action == Get {
			if px != test.Expect {
				t.Errorf("didn't get expected result\n"+
					"Expected: %s\n"+
					"     Got: %s\n",
					test.Expect,
					px,
				)
			}
		}
	}
}

// Test the ChunkCoordinate function.
func TestChunkCoordinates(t *testing.T) {
	c := level.NewChunker(1000)

	type testCase struct {
		In     render.Point
		Expect render.Point
	}
	tests := []testCase{
		testCase{
			In:     render.NewPoint(0, 0),
			Expect: render.NewPoint(0, 0),
		},
		testCase{
			In:     render.NewPoint(128, 128),
			Expect: render.NewPoint(0, 0),
		},
		testCase{
			In:     render.NewPoint(1024, 128),
			Expect: render.NewPoint(1, 0),
		},
		testCase{
			In:     render.NewPoint(3600, 1228),
			Expect: render.NewPoint(3, 1),
		},
		testCase{
			In:     render.NewPoint(-100, -1),
			Expect: render.NewPoint(-1, -1),
		},
		testCase{
			In:     render.NewPoint(-950, 100),
			Expect: render.NewPoint(-1, 0),
		},
		testCase{
			In:     render.NewPoint(-1001, -856),
			Expect: render.NewPoint(-2, -1),
		},
		testCase{
			In:     render.NewPoint(-3600, -4800),
			Expect: render.NewPoint(-4, -5),
		},
	}

	for _, test := range tests {
		actual := c.ChunkCoordinate(test.In)
		if actual != test.Expect {
			t.Errorf(
				"Failed ChunkCoordinate conversion:\n"+
					"   Input: %s\n"+
					"Expected: %s\n"+
					"     Got: %s",
				test.In,
				test.Expect,
				actual,
			)
		}
	}
}

func TestZeroChunkSize(t *testing.T) {
	c := &level.Chunker{}

	coord := c.ChunkCoordinate(render.NewPoint(1200, 3600))
	if !coord.IsZero() {
		t.Errorf("ChunkCoordinate didn't fail with a zero chunk size!")
	}
}
