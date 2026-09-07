package collision_test

import (
	"testing"

	"git.kirsle.net/SketchyMaze/doodle/pkg/collision"
	"git.kirsle.net/SketchyMaze/doodle/pkg/doodads/dummy"
	"git.kirsle.net/SketchyMaze/doodle/pkg/level"
	"git.kirsle.net/go/render"
)

// benchmarkLevel builds a level similar in shape to TestCollisionFunctions:
// a long solid floor with a short wall/ceiling obstacle in the middle, so a
// walking/jumping doodad exercises floor, wall and ceiling collisions.
func benchmarkLevel() (*level.Chunker, *dummy.Drawing) {
	grid := level.NewChunker(128)
	solid := &level.Swatch{
		Name:  "solid",
		Color: render.Black,
		Solid: true,
	}

	for i := 0; i < 1000; i++ {
		grid.Set(render.NewPoint(i, 500), solid)
	}
	for i := 480; i < 500; i++ {
		grid.Set(render.NewPoint(500, i), solid)
	}

	player := dummy.NewPlayer()
	return grid, player
}

// BenchmarkBoxCollidesWithGrid simulates a doodad walking back and forth
// along the floor at a few pixels per tick, which is the steady-state
// workload BoxCollidesWithGrid runs for every moving doodad in a level,
// every tick, during real gameplay.
func BenchmarkBoxCollidesWithGrid(b *testing.B) {
	grid, player := benchmarkLevel()
	size := player.Size()

	start := render.NewPoint(100, 500-size.H)
	player.MoveTo(start)

	const step = 4 // pixels/tick, a typical walking speed
	direction := 1

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pos := player.Position()
		target := render.NewPoint(pos.X+step*direction, pos.Y)

		result, _ := collision.CollidesWithGrid(player, grid, target)
		player.MoveTo(result.MoveTo)

		// Bounce between x=100 and x=400 so the doodad stays on the
		// unobstructed part of the floor and keeps moving every tick.
		if player.Position().X > 400 {
			direction = -1
		} else if player.Position().X < 100 {
			direction = 1
		}
	}
}

// BenchmarkScanBoundingBox isolates the per-pixel edge-scanning hot path
// itself (four edges of a hitbox against level geometry), without the
// movement/physics bookkeeping BoxCollidesWithGrid layers on top.
func BenchmarkScanBoundingBox(b *testing.B) {
	grid, _ := benchmarkLevel()
	box := render.Rect{X: 100, Y: 500 - 56, W: 56, H: 56}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var result collision.Collide
		result.ScanBoundingBox(box, grid)
	}
}
