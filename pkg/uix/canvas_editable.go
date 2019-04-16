package uix

import (
	"git.kirsle.net/apps/doodle/lib/events"
	"git.kirsle.net/apps/doodle/lib/render"
	"git.kirsle.net/apps/doodle/lib/ui"
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
			lastPixel := w.lastPixel
			pixel := &level.Pixel{
				X:      cursor.X,
				Y:      cursor.Y,
				Swatch: w.Palette.ActiveSwatch,
			}

			// Append unique new pixels.
			if len(w.pixelHistory) == 0 || w.pixelHistory[len(w.pixelHistory)-1] != pixel {
				if lastPixel != nil {
					// Draw the pixels in between.
					if lastPixel != pixel {
						for point := range render.IterLine(lastPixel.X, lastPixel.Y, pixel.X, pixel.Y) {
							w.chunks.Set(point, lastPixel.Swatch)
						}
					}
				}

				w.lastPixel = pixel
				w.pixelHistory = append(w.pixelHistory, pixel)

				// Save in the pixel canvas map.
				w.chunks.Set(cursor, pixel.Swatch)
			}
		} else {
			w.lastPixel = nil
		}
	case ActorTool:
		// See if any of the actors are below the mouse cursor.
		var WP = w.WorldIndexAt(cursor)
		_ = WP

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
	}

	return nil

}