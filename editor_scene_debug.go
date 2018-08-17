package doodle

import "git.kirsle.net/apps/doodle/level"

// TODO: build flags to not include this in production builds.
// This adds accessors for private variables from the dev console.

// GetDrawing returns the level.Canvas
func (w *EditorScene) GetDrawing() *level.Canvas {
	return w.drawing
}
