package uix

import (
	"git.kirsle.net/apps/doodle/pkg/drawtool"
	"git.kirsle.net/apps/doodle/pkg/level"
	"git.kirsle.net/go/render"
	"git.kirsle.net/go/render/event"
	"git.kirsle.net/go/ui"
)

// Modified returns whether the canvas has been modified since it was last
// loaded. Methods like Load and LoadFile will set modified to false, and
// commitStroke sets it to true.
func (w *Canvas) Modified() bool {
	return w.modified
}

// SetModified sets the modified bit on the canvas.
func (w *Canvas) SetModified(v bool) {
	w.modified = v
}

// commitStroke is the common function that applies a stroke the user is
// actively drawing onto the canvas. This is for Edit Mode.
func (w *Canvas) commitStroke(tool drawtool.Tool, addHistory bool) {
	if w.currentStroke == nil {
		// nothing to commit
		return
	}

	// Zoom the stroke coordinates (this modifies the pointer)
	zStroke := w.ZoomStroke(w.currentStroke)
	_ = zStroke

	// Mark the canvas as modified.
	w.modified = true

	var (
		deleting = w.currentStroke.Shape == drawtool.Eraser
		dedupe   = map[render.Point]interface{}{} // don't revisit the same point twice

		// Helper functions to set pixels on the level while storing the original
		// value of any pixel being replaced.
		set = func(pt render.Point, sw *level.Swatch) {
			// Take note of what pixel was originally here before we change it.
			if swatch, err := w.chunks.Get(pt); err == nil {
				if _, ok := dedupe[pt]; !ok {
					w.currentStroke.OriginalPoints[pt] = swatch
					dedupe[pt] = nil
				}
			}

			if deleting {
				w.chunks.Delete(pt)
			} else if sw != nil {
				w.chunks.Set(pt, sw)
			} else {
				panic("Canvas.commitStroke.set: current stroke has no level.Swatch in ExtraData")
			}
		}

		// Rects: read existing pixels first, then write new pixels
		readRect = func(rect render.Rect) {
			for pt := range w.chunks.IterViewport(rect) {
				point := pt.Point()
				if _, ok := dedupe[point]; !ok {
					w.currentStroke.OriginalPoints[pt.Point()] = pt.Swatch
					dedupe[point] = nil
				}
			}
		}
		setRect = func(rect render.Rect, sw *level.Swatch) {
			if deleting {
				w.chunks.DeleteRect(rect)
			} else if sw != nil {
				w.chunks.SetRect(rect, sw)
			} else {
				panic("Canvas.commitStroke.setRect: current stroke has no level.Swatch in ExtraData")
			}
		}
	)

	var swatch *level.Swatch
	if v, ok := w.currentStroke.ExtraData.(*level.Swatch); ok {
		swatch = v
	}

	if w.currentStroke.Thickness > 0 {
		// Eraser Tool only: record which pixels will be blown away by this.
		// This is SLOW for thick (rect-based) lines, but eraser tool must have it.
		if deleting {
			for rect := range w.currentStroke.IterThickPoints() {
				readRect(rect)
			}
		}
		for rect := range w.currentStroke.IterThickPoints() {
			setRect(rect, swatch)
		}
	} else {
		for pt := range w.currentStroke.IterPoints() {
			// note: set already records the original pixel if changing it.
			set(pt, swatch)
		}
	}

	// Add the stroke to level history.
	if w.level != nil && addHistory {
		w.level.UndoHistory.AddStroke(w.currentStroke)
	}

	w.RemoveStroke(w.currentStroke)
	w.currentStroke = nil

	w.lastPixel = nil
}

// loopEditable handles the Loop() part for editable canvases.
func (w *Canvas) loopEditable(ev *event.State) error {
	// Get the absolute position of the canvas on screen to accurately match
	// it up to mouse clicks.
	var (
		P      = ui.AbsolutePosition(w)
		cursor = render.Point{
			X: ev.CursorX - P.X - w.Scroll.X,
			Y: ev.CursorY - P.Y - w.Scroll.Y,
		}
	)

	// If the actual cursor is not over the actual Canvas UI element, don't
	// pay any attention to clicks. I added this when I saw you were able to
	// accidentally draw (with large brush size) when clicking on the Palette
	// panel and not the drawing itself.
	if !w.IsCursorOver() {
		return nil
	}

	switch w.Tool {
	case drawtool.PencilTool:
		// If no swatch is active, do nothing with mouse clicks.
		if w.Palette.ActiveSwatch == nil {
			return nil
		}

		// Clicking? Log all the pixels while doing so.
		if ev.Button1 {
			// Initialize a new Stroke for this atomic drawing operation?
			if w.currentStroke == nil {
				w.currentStroke = drawtool.NewStroke(drawtool.Freehand, w.Palette.ActiveSwatch.Color)
				w.currentStroke.Pattern = w.Palette.ActiveSwatch.Pattern
				w.currentStroke.Thickness = w.BrushSize
				w.currentStroke.ExtraData = w.Palette.ActiveSwatch
				w.AddStroke(w.currentStroke)
			}

			lastPixel := w.lastPixel
			pixel := &level.Pixel{
				X:      cursor.X,
				Y:      cursor.Y,
				Swatch: w.Palette.ActiveSwatch,
			}

			// If the user is holding the mouse down over one spot and not
			// moving, don't do anything. The pixel has already been set and
			// needless writes to the map cause needless cache rewrites etc.
			if lastPixel != nil {
				if pixel.X == lastPixel.X && pixel.Y == lastPixel.Y {
					break
				}
			}

			// Append unique new pixels.
			if lastPixel != nil || lastPixel != pixel {
				// Draw the pixels in between.
				if lastPixel != nil && lastPixel != pixel {
					for point := range render.IterLine(lastPixel.Point(), pixel.Point()) {
						w.currentStroke.AddPoint(point)
					}
				}

				w.lastPixel = pixel

				// Save the pixel in the current stroke.
				w.currentStroke.AddPoint(render.Point{
					X: cursor.X,
					Y: cursor.Y,
				})
			}
		} else {
			w.commitStroke(w.Tool, true)
		}
	case drawtool.LineTool:
		// If no swatch is active, do nothing with mouse clicks.
		if w.Palette.ActiveSwatch == nil {
			return nil
		}

		// Clicking? Log all the pixels while doing so.
		if ev.Button1 {
			// Initialize a new Stroke for this atomic drawing operation?
			if w.currentStroke == nil {
				w.currentStroke = drawtool.NewStroke(drawtool.Line, w.Palette.ActiveSwatch.Color)
				w.currentStroke.Pattern = w.Palette.ActiveSwatch.Pattern
				w.currentStroke.Thickness = w.BrushSize
				w.currentStroke.ExtraData = w.Palette.ActiveSwatch
				w.currentStroke.PointA = render.NewPoint(cursor.X, cursor.Y)
				w.AddStroke(w.currentStroke)
			}

			w.currentStroke.PointB = render.NewPoint(cursor.X, cursor.Y)
		} else {
			w.commitStroke(w.Tool, true)
		}
	case drawtool.RectTool:
		// If no swatch is active, do nothing with mouse clicks.
		if w.Palette.ActiveSwatch == nil {
			return nil
		}

		// Clicking? Log all the pixels while doing so.
		if ev.Button1 {
			// Initialize a new Stroke for this atomic drawing operation?
			if w.currentStroke == nil {
				w.currentStroke = drawtool.NewStroke(drawtool.Rectangle, w.Palette.ActiveSwatch.Color)
				w.currentStroke.Pattern = w.Palette.ActiveSwatch.Pattern
				w.currentStroke.Thickness = w.BrushSize
				w.currentStroke.ExtraData = w.Palette.ActiveSwatch
				w.currentStroke.PointA = render.NewPoint(cursor.X, cursor.Y)
				w.AddStroke(w.currentStroke)
			}

			w.currentStroke.PointB = render.NewPoint(cursor.X, cursor.Y)
		} else {
			w.commitStroke(w.Tool, true)
		}
	case drawtool.EllipseTool:
		if w.Palette.ActiveSwatch == nil {
			return nil
		}

		if ev.Button1 {
			if w.currentStroke == nil {
				w.currentStroke = drawtool.NewStroke(drawtool.Ellipse, w.Palette.ActiveSwatch.Color)
				w.currentStroke.Pattern = w.Palette.ActiveSwatch.Pattern
				w.currentStroke.Thickness = w.BrushSize
				w.currentStroke.ExtraData = w.Palette.ActiveSwatch
				w.currentStroke.PointA = render.NewPoint(cursor.X, cursor.Y)
				w.AddStroke(w.currentStroke)
			}

			w.currentStroke.PointB = render.NewPoint(cursor.X, cursor.Y)
		} else {
			w.commitStroke(w.Tool, true)
		}
	case drawtool.EraserTool:
		// Clicking? Log all the pixels while doing so.
		if ev.Button1 {
			// Initialize a new Stroke for this atomic drawing operation?
			if w.currentStroke == nil {
				// The color is white, will look like white-out that covers the
				// wallpaper during the stroke.
				w.currentStroke = drawtool.NewStroke(drawtool.Eraser, render.White)
				w.currentStroke.Thickness = w.BrushSize
				w.AddStroke(w.currentStroke)
			}

			lastPixel := w.lastPixel
			pixel := &level.Pixel{
				X:      cursor.X,
				Y:      cursor.Y,
				Swatch: w.Palette.ActiveSwatch,
			}

			// If the user is holding the mouse down over one spot and not
			// moving, don't do anything. The pixel has already been set and
			// needless writes to the map cause needless cache rewrites etc.
			if lastPixel != nil {
				if pixel.X == lastPixel.X && pixel.Y == lastPixel.Y {
					break
				}
			}

			// Append unique new pixels.
			if lastPixel == nil || lastPixel != pixel {
				if lastPixel != nil && lastPixel != pixel {
					for point := range render.IterLine(lastPixel.Point(), pixel.Point()) {
						w.currentStroke.AddPoint(point)
					}
				}

				w.lastPixel = pixel
				w.currentStroke.AddPoint(render.Point{
					X: cursor.X,
					Y: cursor.Y,
				})
			}
		} else {
			w.commitStroke(w.Tool, true)
		}
	case drawtool.ActorTool:
		// See if any of the actors are below the mouse cursor.
		var WP = w.WorldIndexAt(cursor)

		var deleteActors = []*level.Actor{}
		for _, actor := range w.actors {
			box := render.Rect{
				X: actor.Actor.Point.X - P.X - w.Scroll.X,
				Y: actor.Actor.Point.Y - P.Y - w.Scroll.Y,
				W: actor.Canvas.Size().W,
				H: actor.Canvas.Size().H,
			}

			if WP.Inside(box) {
				actor.Canvas.Configure(ui.Config{
					BorderSize:  1,
					BorderColor: render.RGBA(255, 153, 0, 255),
					BorderStyle: ui.BorderSolid,
					Background:  render.White, // TODO: cuz the border draws a bgcolor
				})

				// Check for a mouse down event to begin dragging this
				// canvas around.
				if ev.Button1 {
					// Pop this canvas out for the drag/drop.
					if w.OnDragStart != nil {
						deleteActors = append(deleteActors, actor.Actor)
						w.OnDragStart(actor.Actor)
					}
					break
				} else if ev.Button3 {
					// Right click to delete an actor.
					deleteActors = append(deleteActors, actor.Actor)
				}
			} else {
				actor.Canvas.SetBorderSize(0)
				actor.Canvas.SetBackground(render.RGBA(0, 0, 1, 0)) // TODO
			}
		}

		// Change in actor count?
		if len(deleteActors) > 0 && w.OnDeleteActors != nil {
			w.OnDeleteActors(deleteActors)
		}
	case drawtool.LinkTool:
		// See if any of the actors are below the mouse cursor.
		var WP = w.WorldIndexAt(cursor)

		for _, actor := range w.actors {
			// Permanently color the actor if it's the current subject of the
			// Link Tool (after 1st click, until 2nd click of other actor)
			if w.linkFirst == actor {
				actor.Canvas.Configure(ui.Config{
					Background: render.RGBA(255, 153, 255, 153),
				})
				continue
			}

			box := render.Rect{
				X: actor.Actor.Point.X - P.X - w.Scroll.X,
				Y: actor.Actor.Point.Y - P.Y - w.Scroll.Y,
				W: actor.Canvas.Size().W,
				H: actor.Canvas.Size().H,
			}

			if WP.Inside(box) {
				actor.Canvas.Configure(ui.Config{
					BorderSize:  1,
					BorderColor: render.RGBA(255, 153, 255, 255),
					BorderStyle: ui.BorderSolid,
					Background:  render.White, // TODO: cuz the border draws a bgcolor
				})

				// Click handler to start linking this actor.
				if ev.Button1 {
					if err := w.LinkAdd(actor); err != nil {
						return err
					}

					// TODO: reset the Button1 state so we don't finish a
					// link and then LinkAdd the clicked doodad immediately
					// (causing link chaining)
					ev.Button1 = false
					break
				}
			} else {
				actor.Canvas.SetBorderSize(0)
				actor.Canvas.SetBackground(render.RGBA(0, 0, 1, 0)) // TODO
			}
		}
	}

	return nil

}
