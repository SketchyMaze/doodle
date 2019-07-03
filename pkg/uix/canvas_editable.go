package uix

import (
	"git.kirsle.net/apps/doodle/lib/events"
	"git.kirsle.net/apps/doodle/lib/render"
	"git.kirsle.net/apps/doodle/lib/ui"
	"git.kirsle.net/apps/doodle/pkg/drawtool"
	"git.kirsle.net/apps/doodle/pkg/level"
)

// loopEditable handles the Loop() part for editable canvases.
func (w *Canvas) loopEditable(ev *events.State) error {
	// Get the absolute position of the canvas on screen to accurately match
	// it up to mouse clicks.
	var (
		P      = ui.AbsolutePosition(w)
		cursor = render.Point{
			X: ev.CursorX.Now - P.X - w.Scroll.X,
			Y: ev.CursorY.Now - P.Y - w.Scroll.Y,
		}
	)

	switch w.Tool {
	case PencilTool:
		// If no swatch is active, do nothing with mouse clicks.
		if w.Palette.ActiveSwatch == nil {
			return nil
		}

		// Clicking? Log all the pixels while doing so.
		if ev.Button1.Now {
			// Initialize a new Stroke for this atomic drawing operation?
			if w.currentStroke == nil {
				w.currentStroke = drawtool.NewStroke(drawtool.Freehand, w.Palette.ActiveSwatch.Color)
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
			if len(w.pixelHistory) == 0 || w.pixelHistory[len(w.pixelHistory)-1] != pixel {
				if lastPixel != nil {
					// Draw the pixels in between.
					if lastPixel != pixel {
						for point := range render.IterLine(lastPixel.X, lastPixel.Y, pixel.X, pixel.Y) {
							w.currentStroke.AddPoint(point)
						}
					}
				}

				w.lastPixel = pixel
				w.pixelHistory = append(w.pixelHistory, pixel)

				// Save the pixel in the current stroke.
				w.currentStroke.AddPoint(render.Point{
					X: cursor.X,
					Y: cursor.Y,
				})
			}
		} else {
			// Mouse released, commit the points to the drawing.
			if w.currentStroke != nil {
				for _, pt := range w.currentStroke.Points {
					w.chunks.Set(pt, w.Palette.ActiveSwatch)
				}

				// Add the stroke to level history.
				if w.level != nil {
					w.level.UndoHistory.AddStroke(w.currentStroke)
				}

				w.RemoveStroke(w.currentStroke)
				w.currentStroke = nil
			}

			w.lastPixel = nil
		}
	case ActorTool:
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
				if ev.Button1.Read() {
					// Pop this canvas out for the drag/drop.
					if w.OnDragStart != nil {
						deleteActors = append(deleteActors, actor.Actor)
						w.OnDragStart(actor.Actor.Filename)
					}
					break
				} else if ev.Button2.Read() {
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
	case LinkTool:
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
				if ev.Button1.Read() {
					if err := w.LinkAdd(actor); err != nil {
						return err
					}
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
