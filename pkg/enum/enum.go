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

// File extensions
const (
	LevelExt     = ".level"
	DoodadExt    = ".doodad"
	LevelPackExt = ".levelpack"
)

// Responsive breakpoints for mobile friendly UIs.
const (
	ScreenWidthXSmall = 400
	ScreenWidthSmall  = 600
	ScreenWidthMedium = 800
	ScreenWidthLarge  = 1000
)
