package level_test

import (
	"fmt"
	"testing"

	"git.kirsle.net/apps/doodle/pkg/level"
	"git.kirsle.net/go/render"
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

func TestViewportChunks(t *testing.T) {
	// Initialize a 100 chunk image with 5x5 chunks.
	var ChunkSize int = 100
	var Offset int = 50
	c := level.NewChunker(int(ChunkSize))
	sw := &level.Swatch{
		Name:  "solid",
		Color: render.Black,
	}

	// The 5x5 chunks are expected to be (diagonally)
	//    -2,-2
	//    -1,-1
	//    0,0
	//    1,1
	//    2,2
	// The chunk size is 100px so place a single pixel in each
	// 100px quadrant.
	fmt.Printf("size=%d  offset=%d\n", ChunkSize, Offset)
	for x := -2; x <= 2; x++ {
		for y := -2; y <= 2; y++ {
			point := render.NewPoint(
				x*ChunkSize+Offset,
				y*ChunkSize+Offset,
			)
			fmt.Printf("in chunk: %d,%d  set pt: %s\n",
				x, y, point,
			)
			c.Set(point, sw)
		}
	}

	// Sanity check the test canvas was created correctly.
	worldSize := c.WorldSize()
	expectSize := render.Rect{
		X: -200,
		Y: -200,
		W: 299,
		H: 299,
	}
	if worldSize != expectSize {
		t.Errorf(
			"Test canvas world size wasn't as expected:\n"+
				"Expected: %s\n"+
				"  Actual: %s\n",
			expectSize,
			worldSize,
		)
	}
	if len(c.Chunks) != 25 {
		t.Errorf(
			"Test canvas chunk count wasn't as expected:\n"+
				"Expected: 25\n"+
				"  Actual: %d\n",
			len(c.Chunks),
		)
	}

	type TestCase struct {
		Viewport render.Rect
		Expect   map[render.Point]interface{}
	}
	var tests = []TestCase{
		{
			Viewport: render.Rect{X: -10000, Y: -10000, W: 10000, H: 10000},
			Expect: map[render.Point]interface{}{
				render.NewPoint(-2, -2): nil,
				render.NewPoint(-2, -1): nil,
				render.NewPoint(-2, 0):  nil,
				render.NewPoint(-2, 1):  nil,
				render.NewPoint(-2, 2):  nil,
				render.NewPoint(-1, -2): nil,
				render.NewPoint(-1, -1): nil,
				render.NewPoint(-1, 0):  nil,
				render.NewPoint(-1, 1):  nil,
				render.NewPoint(-1, 2):  nil,
				render.NewPoint(0, -2):  nil,
				render.NewPoint(0, -1):  nil,
				render.NewPoint(0, 0):   nil,
				render.NewPoint(0, 1):   nil,
				render.NewPoint(0, 2):   nil,
				render.NewPoint(1, -2):  nil,
				render.NewPoint(1, -1):  nil,
				render.NewPoint(1, 0):   nil,
				render.NewPoint(1, 1):   nil,
				render.NewPoint(1, 2):   nil,
				render.NewPoint(2, -2):  nil,
				render.NewPoint(2, -1):  nil,
				render.NewPoint(2, 0):   nil,
				render.NewPoint(2, 1):   nil,
				render.NewPoint(2, 2):   nil,
			},
		},
		{
			Viewport: render.Rect{X: 0, Y: 0, W: 200, H: 200},
			Expect: map[render.Point]interface{}{
				render.NewPoint(0, 0): nil,
				render.NewPoint(0, 1): nil,
				render.NewPoint(1, 0): nil,
				render.NewPoint(1, 1): nil,
			},
		},
		// {
		// 	Viewport: render.Rect{X: -5, Y: 0, W: 200, H: 200},
		// 	Expect: map[render.Point]interface{}{
		// 		render.NewPoint(-1, 0): nil,
		// 		render.NewPoint(0, 0):  nil,
		// 		render.NewPoint(1, 1):  nil,
		// 	},
		// },
	}

	for _, test := range tests {
		chunks := []render.Point{}
		for chunk := range c.IterViewportChunks(test.Viewport) {
			chunks = append(chunks, chunk)
		}

		if len(chunks) != len(test.Expect) {
			t.Errorf("%s: chunk count mismatch: expected %d, got %d",
				test.Viewport,
				len(test.Expect),
				len(chunks),
			)
		}

		for _, actual := range chunks {
			if _, ok := test.Expect[actual]; !ok {
				t.Errorf("%s: got chunk coord %d but did not expect to",
					test.Viewport,
					actual,
				)
			}
			delete(test.Expect, actual)
		}

		if len(test.Expect) > 0 {
			t.Errorf("%s: failed to see these coords: %+v",
				test.Viewport,
				test.Expect,
			)
		}
	}
}
