package render

import (
	"fmt"
)

// Pixel TODO: not a global
// TODO get rid of this ugly thing.
type Pixel struct {
	Start bool
	X     int32
	Y     int32
	DX    int32
	DY    int32
}

func (p Pixel) String() string {
	return fmt.Sprintf("(%d,%d) delta (%d,%d)",
		p.X, p.Y,
		p.DX, p.DY,
	)
}

// Grid is a 2D grid of pixels in X,Y notation.
type Grid map[Pixel]interface{}

// Exists returns true if the point exists on the grid.
func (g *Grid) Exists(p Pixel) bool {
	if _, ok := (*g)[p]; ok {
		return true
	}
	return false
}

// Draw the grid efficiently.
func (g *Grid) Draw(e Engine) {
	for pixel := range *g {
		if pixel.DX == 0 && pixel.DY == 0 {
			e.DrawPoint(Black, Point{
				X: pixel.X,
				Y: pixel.Y,
			})
		} else {
			for point := range IterLine(pixel.X, pixel.Y, pixel.DX, pixel.DY) {
				e.DrawPoint(Black, point)
			}
		}
	}
}
