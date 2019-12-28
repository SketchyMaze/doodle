package uix

import (
	"git.kirsle.net/go/render"
	"git.kirsle.net/go/ui"
	"git.kirsle.net/apps/doodle/pkg/shmem"
)

// IsCursorOver returns true if the mouse cursor is physically over top
// of the canvas's widget space.
func (w *Canvas) IsCursorOver() bool {
	var (
		P = ui.AbsolutePosition(w)
		S = w.Size()
	)
	return shmem.Cursor.Inside(render.Rect{
		X: P.X,
		Y: P.Y,
		W: S.W,
		H: S.H,
	})
}

// presentCursor draws something at the mouse cursor on the Canvas.
//
// This is currently used in Edit Mode when you're drawing a shape with a thick
// brush size, and draws a "preview rect" under the cursor of how big a click
// will be at that size.
func (w *Canvas) presentCursor(e render.Engine) {
	if !w.IsCursorOver() {
		return
	}

	// Are we editing with a thick brush?
	if w.BrushSize > 0 {
		var r = int32(w.BrushSize)
		rect := render.Rect{
			X: shmem.Cursor.X - r,
			Y: shmem.Cursor.Y - r,
			W: r * 2,
			H: r * 2,
		}
		e.DrawRect(render.Black, rect)
		rect.X++
		rect.Y++
		rect.W -= 2
		rect.H -= 2
		e.DrawRect(render.RGBA(153, 153, 153, 153), rect)
	}
}
