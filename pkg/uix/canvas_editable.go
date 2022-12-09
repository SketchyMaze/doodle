package uix

import (
	"git.kirsle.net/SketchyMaze/doodle/pkg/balance"
	"git.kirsle.net/SketchyMaze/doodle/pkg/drawtool"
	"git.kirsle.net/SketchyMaze/doodle/pkg/keybind"
	"git.kirsle.net/SketchyMaze/doodle/pkg/level"
	"git.kirsle.net/SketchyMaze/doodle/pkg/log"
	"git.kirsle.net/SketchyMaze/doodle/pkg/shmem"
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

	// Zoom the stroke coordinates (this modifies the pointer).
	// Note: all the points on the stroke were mouse cursor coordinates on the screen.
	w.currentStroke = w.ZoomStroke(w.currentStroke)

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
	if addHistory {
		w.strokeToHistory(w.currentStroke)
	}

	w.RemoveStroke(w.currentStroke)
	w.currentStroke = nil

	w.lastPixel = nil
}

// Add a recently drawn stroke to the UndoHistory.
func (w *Canvas) strokeToHistory(stroke *drawtool.Stroke) {
	if w.level != nil {
		w.level.UndoHistory.AddStroke(stroke)
	} else if w.doodad != nil {
		if w.doodad.UndoHistory == nil {
			// HACK: if UndoHistory was not initialized properly.
			w.doodad.UndoHistory = drawtool.NewHistory(balance.UndoHistory)
		}
		w.doodad.UndoHistory.AddStroke(stroke)
	}
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
	case drawtool.PanTool:
		// Pan tool = click to pan the level.
		var delta render.Point
		if keybind.LeftClick(ev) || keybind.MiddleClick(ev) {
			if !w.scrollDragging {
				w.scrollDragging = true
				w.scrollStartAt = shmem.Cursor
				w.scrollWasAt = w.Scroll
			} else {
				delta = shmem.Cursor.Compare(w.scrollStartAt)
				w.Scroll = w.scrollWasAt
				w.Scroll.Subtract(delta)

				// TODO: if I don't call this, the user is able to (temporarily!)
				// pan outside the level boundaries before it snaps-back when they
				// release. But the normal middle-click to pan code doesn't let
				// them do this.. investigate why later.
				w.loopConstrainScroll()
			}
		} else {
			if w.scrollDragging {
				w.scrollDragging = false
			}
		}

		// All the Pan tool to still interact with the Settings button on mouse-over
		// of an actor. On touch devices it's difficult to access an actor's settings
		// without accidentally dragging the actor, so the Pan Tool allows safe access.
		// NOTE: code copied from Actor Tool but with delete and drag/drop hooks removed.
		var WP = w.WorldIndexAt(cursor)
		for _, actor := range w.actors {

			// Compute the bounding box on screen where this doodad
			// visually appears.
			var scrollBias = render.Point{
				X: w.Scroll.X,
				Y: w.Scroll.Y,
			}
			if w.Zoom != 0 {
				scrollBias.X = w.ZoomDivide(scrollBias.X)
				scrollBias.Y = w.ZoomDivide(scrollBias.Y)
			}
			box := render.Rect{
				X: actor.Actor.Point.X - scrollBias.X - w.ZoomDivide(P.X),
				Y: actor.Actor.Point.Y - scrollBias.Y - w.ZoomDivide(P.Y),
				W: actor.Canvas.Size().W,
				H: actor.Canvas.Size().H,
			}

			// Mouse hover?
			if WP.Inside(box) {
				actor.Canvas.Configure(ui.Config{
					BorderSize:  1,
					BorderColor: render.RGBA(153, 153, 153, 255),
					BorderStyle: ui.BorderSolid,
					Background:  render.White, // TODO: cuz the border draws a bgcolor
				})

				// Show doodad buttons.
				actor.Canvas.ShowDoodadButtons = true

				// Check for a mouse down event to begin dragging this
				// canvas around.
				if keybind.LeftClick(ev) && delta == render.Origin {
					// Did they click onto the doodad buttons?
					if shmem.Cursor.Inside(actor.Canvas.doodadButtonRect()) {
						keybind.ClearLeftClick(ev)
						if w.OnDoodadConfig != nil {
							w.OnDoodadConfig(actor)
						} else {
							log.Error("OnDoodadConfig: handler not defined for parent canvas")
						}
						return nil
					}

					break
				}
			} else {
				actor.Canvas.SetBorderSize(0)
				actor.Canvas.SetBackground(render.RGBA(0, 0, 1, 0)) // TODO
				actor.Canvas.ShowDoodadButtons = false
			}
		}

	case drawtool.PencilTool:
		// If no swatch is active, do nothing with mouse clicks.
		if w.Palette.ActiveSwatch == nil {
			return nil
		}

		// Clicking? Log all the pixels while doing so.
		if keybind.LeftClick(ev) {
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
		if keybind.LeftClick(ev) {
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
		if keybind.LeftClick(ev) {
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

		if keybind.LeftClick(ev) {
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

	case drawtool.TextTool:
		// The Text Tool popup should initialize this for us, if somehow not
		// initialized skip this tool processing.
		if w.Palette.ActiveSwatch == nil || drawtool.TT.IsZero() {
			return nil
		}

		// Do we need to create the Label?
		if drawtool.TT.Label == nil {
			drawtool.TT.Label = ui.NewLabel(ui.Label{
				Text: drawtool.TT.Message,
				Font: render.Text{
					FontFilename: drawtool.TT.Font,
					Size:         drawtool.TT.Size,
					Color:        w.Palette.ActiveSwatch.Color,
				},
			})
		}

		// Do we need to update the color of the label?
		if drawtool.TT.Label.Font.Color != w.Palette.ActiveSwatch.Color {
			drawtool.TT.Label.Font.Color = w.Palette.ActiveSwatch.Color
		}

		// NOTE: Canvas.presentStrokes() will handle drawing the font preview
		// at the cursor location while the TextTool is active.

		// On mouse click, commit the text to the drawing.
		if keybind.LeftClick(ev) {
			if stroke, err := drawtool.TT.ToStroke(shmem.CurrentRenderEngine, w.Palette.ActiveSwatch.Color, cursor); err != nil {
				shmem.FlashError("Text Tool error: %s", err)
				return nil
			} else {
				w.currentStroke = stroke
				w.currentStroke.ExtraData = w.Palette.ActiveSwatch
				w.commitStroke(drawtool.PencilTool, true)
			}

			keybind.ClearLeftClick(ev)
		}

	case drawtool.FloodTool:
		if w.Palette.ActiveSwatch == nil {
			return nil
		}

		// Click to activate.
		if keybind.LeftClick(ev) {
			var (
				chunker = w.Chunker()
				stroke  = drawtool.NewStroke(drawtool.Freehand, w.Palette.ActiveSwatch.Color)
			)

			// Set some max boundaries to prevent runaway infinite loops, e.g. if user
			// clicked the wide open void the flood fill would never finish!
			limit := balance.FloodToolLimit

			// Get the original color at this location.
			// Error cases can include: no chunk at this spot, or no pixel at this spot.
			// Treat these as just a null color and proceed anyway, user should be able
			// to flood fill blank areas of their level.
			baseColor, err := chunker.Get(cursor)
			if err != nil {
				limit = balance.FloodToolVoidLimit
				log.Warn("FloodTool: couldn't get base color at %s: %s (got %s)", cursor, err)
			}

			// If no change, do nothing.
			if baseColor == w.Palette.ActiveSwatch {
				break
			}

			// The flood fill algorithm.
			queue := []render.Point{cursor}
			for len(queue) > 0 {
				node := queue[0]
				queue = queue[1:]

				colorAt, _ := chunker.Get(node)
				if colorAt != baseColor {
					continue
				}

				// For Undo history, store the original color at this point.
				if colorAt != nil {
					stroke.OriginalPoints[node] = colorAt
				}

				// Add the neighboring pixels.
				for _, neighbor := range []render.Point{
					{X: node.X - 1, Y: node.Y},
					{X: node.X + 1, Y: node.Y},
					{X: node.X, Y: node.Y - 1},
					{X: node.X, Y: node.Y + 1},
				} {
					// Only if not too far from the origin!
					if render.AbsInt(neighbor.X-cursor.X) <= limit && render.AbsInt(neighbor.Y-cursor.Y) <= limit {
						queue = append(queue, neighbor)
					}
				}

				stroke.AddPoint(node)
				err = chunker.Set(node, w.Palette.ActiveSwatch)
				if err != nil {
					log.Error("FloodTool: error setting %s to %s: %s", node, w.Palette.ActiveSwatch, err)
				}
			}

			w.strokeToHistory(stroke)
			keybind.ClearLeftClick(ev)
		}

	case drawtool.EraserTool:
		// Clicking? Log all the pixels while doing so.
		if keybind.LeftClick(ev) {
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

		var deleteActors = []*Actor{}
		for _, actor := range w.actors {

			// Compute the bounding box on screen where this doodad
			// visually appears.
			var scrollBias = render.Point{
				X: w.Scroll.X,
				Y: w.Scroll.Y,
			}
			if w.Zoom != 0 {
				scrollBias.X = w.ZoomDivide(scrollBias.X)
				scrollBias.Y = w.ZoomDivide(scrollBias.Y)
			}
			box := render.Rect{
				X: actor.Actor.Point.X - scrollBias.X - w.ZoomDivide(P.X),
				Y: actor.Actor.Point.Y - scrollBias.Y - w.ZoomDivide(P.Y),
				W: actor.Canvas.Size().W,
				H: actor.Canvas.Size().H,
			}

			// Mouse hover?
			if WP.Inside(box) {
				actor.Canvas.Configure(ui.Config{
					BorderSize:  1,
					BorderColor: render.RGBA(255, 153, 0, 255),
					BorderStyle: ui.BorderSolid,
					Background:  render.White, // TODO: cuz the border draws a bgcolor
				})

				// Show doodad buttons.
				actor.Canvas.ShowDoodadButtons = true

				// Check for a mouse down event to begin dragging this
				// canvas around.
				if keybind.LeftClick(ev) {
					// Did they click onto the doodad buttons?
					if shmem.Cursor.Inside(actor.Canvas.doodadButtonRect()) {
						keybind.ClearLeftClick(ev)
						if w.OnDoodadConfig != nil {
							w.OnDoodadConfig(actor)
						} else {
							log.Error("OnDoodadConfig: handler not defined for parent canvas")
						}
						return nil
					}

					// Pop this canvas out for the drag/drop.
					if w.OnDragStart != nil {
						deleteActors = append(deleteActors, actor)
						w.OnDragStart(actor.Actor)
					}
					break
				} else if ev.Button3 {
					// Right click to delete an actor.
					deleteActors = append(deleteActors, actor)
				}
			} else {
				actor.Canvas.SetBorderSize(0)
				actor.Canvas.SetBackground(render.RGBA(0, 0, 1, 0)) // TODO
				actor.Canvas.ShowDoodadButtons = false
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
			// Compute the bounding box on screen where this doodad
			// visually appears.
			var scrollBias = render.Point{
				X: w.Scroll.X,
				Y: w.Scroll.Y,
			}
			if w.Zoom != 0 {
				scrollBias.X = w.ZoomDivide(scrollBias.X)
				scrollBias.Y = w.ZoomDivide(scrollBias.Y)
			}
			box := render.Rect{
				X: actor.Actor.Point.X - scrollBias.X - w.ZoomDivide(P.X),
				Y: actor.Actor.Point.Y - scrollBias.Y - w.ZoomDivide(P.Y),
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
				if keybind.LeftClick(ev) {
					if err := w.LinkAdd(actor); err != nil {
						return err
					}

					// TODO: reset the Button1 state so we don't finish a
					// link and then LinkAdd the clicked doodad immediately
					// (causing link chaining)
					keybind.ClearLeftClick(ev)
					break
				}
			} else {
				actor.Canvas.SetBorderSize(0)
				actor.Canvas.SetBackground(render.RGBA(0, 0, 1, 0)) // TODO
			}

			// Permanently color the actor if it's the current subject of the
			// Link Tool (after 1st click, until 2nd click of other actor)
			if w.linkFirst == actor {
				actor.Canvas.Configure(ui.Config{
					Background: render.RGBA(255, 153, 255, 153),
				})
			}
		}
	}

	return nil

}
