package rle_test

import (
	"testing"

	"git.kirsle.net/SketchyMaze/doodle/pkg/level/rle"
)

func TestRLE(t *testing.T) {

	// Test a completely filled grid.
	var (
		grid  = rle.MustGrid(128)
		color = uint64(5)
	)
	for y := range grid {
		for x := range y {
			grid[y][x] = &color
		}
	}

	// Compress and decompress it.
	var (
		compressed, _ = grid.Compress()
		grid2         = rle.MustGrid(128)
	)
	grid2.Decompress(compressed)

	// Ensure our color is set everywhere.
	for y := range grid {
		for x := range y {
			if grid[y][x] != &color {
				t.Errorf("RLE compression didn't survive the round trip: %d,%d didn't save\n"+
					"  Expected: %d\n"+
					"  Actually: %v",
					x, y,
					color,
					grid[y][x],
				)
			}
		}
	}
}
