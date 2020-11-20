package uix

import (
	"git.kirsle.net/apps/doodle/pkg/drawtool"
	"git.kirsle.net/go/render"
)

// Functions related to the Zoom Tool to magnify the size of the canvas.

/*
GetZoomMultiplier parses the .Zoom integer and returns a multiplier.

Examples:

	Zoom = 0: neutral (100% scale, 1x)
	Zoom = 1: 2x zoom
	Zoom = 2: 4x zoom
	Zoom = 3: 8x zoom
	Zoom = -1: 0.5x zoom
	Zoom = -2: 0.25x zoom
*/
func (w *Canvas) GetZoomMultiplier() float64 {
	// Get and bounds cap the zoom setting.
	if w.Zoom < -2 {
		w.Zoom = -2
	} else if w.Zoom > 3 {
		w.Zoom = 3
	}

	// Return the multipliers.
	switch w.Zoom {
	case -2:
		return 0.25
	case -1:
		return 0.5
	case 0:
		return 1
	case 1:
		return 1.5
	case 2:
		return 2
	case 3:
		return 2.5
	default:
		return 1
	}
}

/*
ZoomMultiply multiplies a width or height value by the Zoom Multiplier and
returns the modified integer.

Usage is like:

	// when building a render.Rect destination box.
	dest.W *= ZoomMultiply(dest.W)
	dest.H *= ZoomMultiply(dest.H)
*/
func (w *Canvas) ZoomMultiply(value int) int {
	return int(float64(value) * w.GetZoomMultiplier())
}

/*
ZoomStroke adjusts a drawn stroke on the canvas to account for the zoom level.

Returns a copy Stroke value without changing the original.
*/
func (w *Canvas) ZoomStroke(stroke *drawtool.Stroke) drawtool.Stroke {
	copy := drawtool.Stroke{
		ID:             stroke.ID,
		Shape:          stroke.Shape,
		Color:          stroke.Color,
		Thickness:      stroke.Thickness,
		ExtraData:      stroke.ExtraData,
		PointA:         stroke.PointA,
		PointB:         stroke.PointB,
		Points:         stroke.Points,
		OriginalPoints: stroke.OriginalPoints,
	}
	return copy

	// Multiply all coordinates in this stroke, which should be World
	// Coordinates in the level data, by the zoom multiplier.
	adjust := func(p render.Point) render.Point {
		p.X = w.ZoomMultiply(p.X)
		p.Y = w.ZoomMultiply(p.Y)
		return p
	}

	copy.PointA = adjust(copy.PointA)
	copy.PointB = adjust(copy.PointB)
	for i := range copy.Points {
		copy.Points[i] = adjust(copy.Points[i])
	}

	return copy
}
