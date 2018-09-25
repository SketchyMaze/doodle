// Package enum defines all the little enum types used throughout Doodle.
package enum

// DrawingType tells the EditorScene whether the currently open drawing is
// a Level or a Doodad.
type DrawingType int

// EditorType values.
const (
	LevelDrawing DrawingType = iota
	DoodadDrawing
)
