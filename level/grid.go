package level

import (
	"git.kirsle.net/apps/doodle/render"
)

// Grid is a 2D grid of pixels in X,Y notation.
type Grid map[*Pixel]interface{}

// Exists returns true if the point exists on the grid.
func (g *Grid) Exists(p *Pixel) bool {
	if _, ok := (*g)[p]; ok {
		return true
	}
	return false
}

// Draw the grid efficiently.
func (g *Grid) Draw(e render.Engine) {
	for pixel := range *g {
		color := pixel.Swatch.Color
		e.DrawPoint(color, render.Point{
			X: pixel.X,
			Y: pixel.Y,
		})
	}
}
