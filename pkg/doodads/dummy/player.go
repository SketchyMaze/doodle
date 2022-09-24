// Package dummy implements a dummy doodads.Drawing.
package dummy

import "git.kirsle.net/SketchyMaze/doodle/pkg/doodads"

// NewPlayer creates a dummy player object.
func NewPlayer() *Drawing {
	return &Drawing{
		Drawing: doodads.NewDrawing("PLAYER", doodads.New(32)),
	}
}
