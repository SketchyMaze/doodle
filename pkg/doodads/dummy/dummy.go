// Package dummy implements a dummy doodads.Drawing.
package dummy

import "git.kirsle.net/apps/doodle/pkg/doodads"

// Drawing is a dummy doodads.Drawing that has no data.
type Drawing struct {
	doodads.Drawing
}

// NewDrawing creates a new dummy drawing.
func NewDrawing(id string, doodad *doodads.Doodad) *Drawing {
	return &Drawing{
		Drawing: doodads.NewDrawing(id, doodad),
	}
}
