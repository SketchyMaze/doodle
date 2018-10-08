package doodle

import "git.kirsle.net/apps/doodle/uix"

// TODO: build flags to not include this in production builds.
// This adds accessors for private variables from the dev console.

// GetDrawing returns the uix.Canvas
func (w *EditorScene) GetDrawing() *uix.Canvas {
	return w.UI.Canvas
}
