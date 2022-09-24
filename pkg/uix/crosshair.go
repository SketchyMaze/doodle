package uix

import (
	"git.kirsle.net/SketchyMaze/doodle/pkg/shmem"
	"git.kirsle.net/go/render"
	"git.kirsle.net/go/ui"
)

/*
Functions for the Crosshair feature of the game.

NOT dependent on Canvas!
*/

type Crosshair struct {
	LengthPct float64 // between 0 and 1
	Widget    ui.Widget
}

func NewCrosshair(child ui.Widget, length float64) *Crosshair {
	return &Crosshair{
		LengthPct: length,
		Widget:    child,
	}
}

// DrawCrosshair renders a crosshair on the screen. It appears while the mouse cursor is
// over the child widget and draws within the bounds of the child widget.
//
// The lengthPct is an integer ranged 0 to 100 to be a percentage length of the crosshair.
// A value of zero will not draw anything and just return.
func DrawCrosshair(e render.Engine, child ui.Widget, color render.Color, lengthPct int) {
	// Get our window boundaries based on our widget.
	var (
		Position = ui.AbsolutePosition(child)
		Size     = child.Size()
		Cursor   = shmem.Cursor
		VertLine = []render.Point{
			{X: Cursor.X, Y: Position.Y},
			{X: Cursor.X, Y: Position.Y + Size.H},
		}
		HozLine = []render.Point{
			{X: Position.X, Y: Cursor.Y},
			{X: Position.X + Size.W, Y: Cursor.Y},
		}
	)

	if lengthPct > 100 {
		lengthPct = 100
	} else if lengthPct <= 0 {
		return
	}

	// Mouse outside our box.
	if Cursor.X < Position.X || Cursor.X > Position.X+Size.W ||
		Cursor.Y < Position.Y || Cursor.Y > Position.Y+Size.H {
		return
	}

	e.DrawLine(
		color,
		VertLine[0],
		VertLine[1],
	)
	e.DrawLine(
		color,
		HozLine[0],
		HozLine[1],
	)
}
