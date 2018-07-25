package render

import (
	"git.kirsle.net/apps/doodle/level"
)

// Grid is a 2D grid of pixels in X,Y notation.
type Grid map[level.Pixel]interface{}

// Exists returns true if the point exists on the grid.
func (g *Grid) Exists(p level.Pixel) bool {
	if _, ok := (*g)[p]; ok {
		return true
	}
	return false
}

// Draw the grid efficiently.
func (g *Grid) Draw(e Engine) {
	for pixel := range *g {
		e.DrawPoint(Black, Point{
			X: pixel.X,
			Y: pixel.Y,
		})
	}
}
