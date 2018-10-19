package uix

import (
	"fmt"
	"strings"

	"git.kirsle.net/apps/doodle/balance"
	"git.kirsle.net/apps/doodle/doodads"
	"git.kirsle.net/apps/doodle/events"
	"git.kirsle.net/apps/doodle/level"
	"git.kirsle.net/apps/doodle/pkg/userdir"
	"git.kirsle.net/apps/doodle/render"
	"git.kirsle.net/apps/doodle/ui"
)

// Canvas is a custom ui.Widget that manages a single drawing.
type Canvas struct {
	ui.Frame
	Palette *level.Palette

	// Set to true to allow clicking to edit this canvas.
	Editable   bool
	Scrollable bool

	// Underlying chunk data for the drawing.
	chunks *level.Chunker

	// Actors to superimpose on top of the drawing.
	actor  *level.Actor // if this canvas IS an actor
	actors []*Actor

	// Tracking pixels while editing. TODO: get rid of pixelHistory?
	pixelHistory []*level.Pixel
	lastPixel    *level.Pixel

	// We inherit the ui.Widget which manages the width and height.
	Scroll render.Point // Scroll offset for which parts of canvas are visible.
}

// Actor is an instance of an actor with a Canvas attached.
type Actor struct {
	Actor  *level.Actor
	Canvas *Canvas
}

// NewCanvas initializes a Canvas widget.
//
// If editable is true, Scrollable is also set to true, which means the arrow
// keys will scroll the canvas viewport which is desirable in Edit Mode.
func NewCanvas(size int, editable bool) *Canvas {
	w := &Canvas{
		Editable:   editable,
		Scrollable: editable,
		Palette:    level.NewPalette(),
		chunks:     level.NewChunker(size),
		actors:     make([]*Actor, 0),
	}
	w.setup()
	w.IDFunc(func() string {
		var attrs []string

		if w.Editable {
			attrs = append(attrs, "editable")
		} else {
			attrs = append(attrs, "read-only")
		}

		if w.Scrollable {
			attrs = append(attrs, "scrollable")
		}

		return fmt.Sprintf("Canvas<%d; %s>", size, strings.Join(attrs, "; "))
	})
	return w
}

// Load initializes the Canvas using an existing Palette and Grid.
func (w *Canvas) Load(p *level.Palette, g *level.Chunker) {
	w.Palette = p
	w.chunks = g

	if len(w.Palette.Swatches) > 0 {
		w.SetSwatch(w.Palette.Swatches[0])
	}
}

// LoadLevel initializes a Canvas from a Level object.
func (w *Canvas) LoadLevel(level *level.Level) {
	w.Load(level.Palette, level.Chunker)
}

// LoadDoodad initializes a Canvas from a Doodad object.
func (w *Canvas) LoadDoodad(d *doodads.Doodad) {
	// TODO more safe
	w.Load(d.Palette, d.Layers[0].Chunker)
}

// InstallActors adds external Actors to the canvas to be superimposed on top
// of the drawing.
func (w *Canvas) InstallActors(actors level.ActorMap) error {
	w.actors = make([]*Actor, 0)
	for id, actor := range actors {
		log.Info("InstallActors: %s", id)

		doodad, err := doodads.LoadJSON(userdir.DoodadPath(actor.Filename))
		if err != nil {
			return fmt.Errorf("InstallActors: %s", err)
		}

		size := int32(doodad.Layers[0].Chunker.Size)
		can := NewCanvas(int(size), false)
		can.Name = id
		can.actor = actor
		can.LoadDoodad(doodad)
		can.Resize(render.NewRect(size, size))
		w.actors = append(w.actors, &Actor{
			Actor:  actor,
			Canvas: can,
		})
	}
	return nil
}

// SetSwatch changes the currently selected swatch for editing.
func (w *Canvas) SetSwatch(s *level.Swatch) {
	w.Palette.ActiveSwatch = s
}

// setup common configs between both initializers of the canvas.
func (w *Canvas) setup() {
	w.SetBackground(render.White)

	// XXX: Debug code.
	if balance.DebugCanvasBorder != render.Invisible {
		w.Configure(ui.Config{
			BorderColor: balance.DebugCanvasBorder,
			BorderSize:  2,
			BorderStyle: ui.BorderSolid,
		})
	}

	w.Handle(ui.MouseOver, func(p render.Point) {
		w.SetBackground(render.Yellow)
	})
	w.Handle(ui.MouseOut, func(p render.Point) {
		w.SetBackground(render.SkyBlue)
	})
}

// Loop is called on the scene's event loop to handle mouse interaction with
// the canvas, i.e. to edit it.
func (w *Canvas) Loop(ev *events.State) error {
	// Get the absolute position of the canvas on screen to accurately match
	// it up to mouse clicks.
	var P = ui.AbsolutePosition(w)

	if w.Scrollable {
		// Arrow keys to scroll the view.
		scrollBy := render.Point{}
		if ev.Right.Now {
			scrollBy.X -= balance.CanvasScrollSpeed
		} else if ev.Left.Now {
			scrollBy.X += balance.CanvasScrollSpeed
		}
		if ev.Down.Now {
			scrollBy.Y -= balance.CanvasScrollSpeed
		} else if ev.Up.Now {
			scrollBy.Y += balance.CanvasScrollSpeed
		}
		if !scrollBy.IsZero() {
			w.ScrollBy(scrollBy)
		}
	}

	// Only care if the cursor is over our space.
	cursor := render.NewPoint(ev.CursorX.Now, ev.CursorY.Now)
	if !cursor.Inside(ui.AbsoluteRect(w)) {
		return nil
	}

	// If no swatch is active, do nothing with mouse clicks.
	if w.Palette.ActiveSwatch == nil {
		return nil
	}

	// Clicking? Log all the pixels while doing so.
	if ev.Button1.Now {
		lastPixel := w.lastPixel
		cursor := render.Point{
			X: ev.CursorX.Now - P.X - w.Scroll.X,
			Y: ev.CursorY.Now - P.Y - w.Scroll.Y,
		}
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

	return nil
}

// Viewport returns a rect containing the viewable drawing coordinates in this
// canvas. The X,Y values are the scroll offset (top left) and the W,H values
// are the scroll offset plus the width/height of the Canvas widget.
//
// The Viewport rect are the Absolute World Coordinates of the drawing that are
// visible inside the Canvas. The X,Y is the top left World Coordinate and the
// W,H are the bottom right World Coordinate, making this rect an absolute
// slice of the world. For a normal rect with a relative width and height,
// use ViewportRelative().
//
// The rect X,Y are the negative Scroll Value.
// The rect W,H are the Canvas widget size minus the Scroll Value.
func (w *Canvas) Viewport() render.Rect {
	var S = w.Size()
	return render.Rect{
		X: -w.Scroll.X,
		Y: -w.Scroll.Y,
		W: S.W - w.Scroll.X,
		H: S.H - w.Scroll.Y,
	}
}

// ViewportRelative returns a relative viewport where the Width and Height
// values are zero-relative: so you can use it with point.Inside(viewport)
// to see if a World Index point should be visible on screen.
//
// The rect X,Y are the negative Scroll Value
// The rect W,H are the Canvas widget size.
func (w *Canvas) ViewportRelative() render.Rect {
	var S = w.Size()
	return render.Rect{
		X: -w.Scroll.X,
		Y: -w.Scroll.Y,
		W: S.W,
		H: S.H,
	}
}

// Chunker returns the underlying Chunker object.
func (w *Canvas) Chunker() *level.Chunker {
	return w.chunks
}

// ScrollTo sets the viewport scroll position.
func (w *Canvas) ScrollTo(to render.Point) {
	w.Scroll.X = to.X
	w.Scroll.Y = to.Y
}

// ScrollBy adjusts the viewport scroll position.
func (w *Canvas) ScrollBy(by render.Point) {
	w.Scroll.Add(by)
}

// Compute the canvas.
func (w *Canvas) Compute(e render.Engine) {

}

// Present the canvas.
func (w *Canvas) Present(e render.Engine, p render.Point) {
	var (
		S        = w.Size()
		Viewport = w.Viewport()
	)
	// w.MoveTo(p) // TODO: when uncommented the canvas will creep down the Workspace frame in EditorMode
	w.DrawBox(e, p)
	e.DrawBox(w.Background(), render.Rect{
		X: p.X + w.BoxThickness(1),
		Y: p.Y + w.BoxThickness(1),
		W: S.W - w.BoxThickness(2),
		H: S.H - w.BoxThickness(2),
	})

	// Get the chunks in the viewport and cache their textures.
	for coord := range w.chunks.IterViewportChunks(Viewport) {
		if chunk, ok := w.chunks.GetChunk(coord); ok {
			tex := chunk.Texture(e, w.Name+coord.String())
			src := render.Rect{
				W: tex.Size().W,
				H: tex.Size().H,
			}

			// If the source bitmap is already bigger than the Canvas widget
			// into which it will render, cap the source width and height.
			// This is especially useful for Doodad buttons because the drawing
			// is bigger than the button.
			if src.W > S.W {
				src.W = S.W
			}
			if src.H > S.H {
				src.H = S.H
			}

			dst := render.Rect{
				X: p.X + w.Scroll.X + w.BoxThickness(1) + (coord.X * int32(chunk.Size)),
				Y: p.Y + w.Scroll.Y + w.BoxThickness(1) + (coord.Y * int32(chunk.Size)),

				// src.W and src.H will be AT MOST the full width and height of
				// a Canvas widget. Subtract the scroll offset to keep it bounded
				// visually on its right and bottom sides.
				W: src.W,
				H: src.H,
			}

			// If the destination width will cause it to overflow the widget
			// box, trim off the right edge of the destination rect.
			//
			// Keep in mind we're dealing with chunks here, and a chunk is
			// a small part of the image. Example:
			// - Canvas is 800x600 (S.W=800  S.H=600)
			// - Chunk wants to render at 790,0 width 100,100 or whatever
			//   dst={790, 0, 100, 100}
			// - Chunk box would exceed 800px width (X=790 + W=100 == 890)
			// - Find the delta how much it exceeds as negative (800 - 890 == -90)
			// - Lower the Source and Dest rects by that delta size so they
			//   stay proportional and don't scale or anything dumb.
			if dst.X+src.W > p.X+S.W {
				// NOTE: delta is a negative number,
				// so it will subtract from the width.
				delta := (p.X + S.W - w.BoxThickness(1)) - (dst.W + dst.X)
				src.W += delta
				dst.W += delta
			}
			if dst.Y+src.H > p.Y+S.H {
				// NOTE: delta is a negative number
				delta := (p.Y + S.H - w.BoxThickness(1)) - (dst.H + dst.Y)
				src.H += delta
				dst.H += delta
			}

			// The same for the top left edge, so the drawings don't overlap
			// menu bars or left side toolbars.
			// - Canvas was placed 80px from the left of the screen.
			//   Canvas.MoveTo(80, 0)
			// - A texture wants to draw at 60, 0 which would cause it to
			//   overlap 20 pixels into the left toolbar. It needs to be cropped.
			// - The delta is: p.X=80 - dst.X=60 == 20
			// - Set destination X to p.X to constrain it there: 20
			// - Subtract the delta from destination W so we don't scale it.
			// - Add 20 to X of the source: the left edge of source is not visible
			if dst.X < p.X {
				// NOTE: delta is a positive number,
				// so it will add to the destination coordinates.
				delta := p.X - dst.X
				dst.X = p.X + w.BoxThickness(1)
				dst.W -= delta
				src.X += delta
			}
			if dst.Y < p.Y {
				delta := p.Y - dst.Y
				dst.Y = p.Y + w.BoxThickness(1)
				dst.H -= delta
				src.Y += delta
			}

			// Trim the destination width so it doesn't overlap the Canvas border.
			if dst.W >= S.W-w.BoxThickness(1) {
				dst.W = S.W - w.BoxThickness(1)
			}

			e.Copy(tex, src, dst)
		}
	}

	w.drawActors(e, p)

	// XXX: Debug, show label in canvas corner.
	if balance.DebugCanvasLabel {
		rows := []string{
			w.Name,

			// XXX: debug options, uncomment for more details

			// Size of the canvas
			// fmt.Sprintf("S=%d,%d", S.W, S.H),

			// Viewport of the canvas
			// fmt.Sprintf("V=%d,%d:%d,%d",
			// 	Viewport.X, Viewport.Y,
			// 	Viewport.W, Viewport.H,
			// ),
		}
		if w.actor != nil {
			rows = append(rows,
				fmt.Sprintf("WP=%s", w.actor.Point),
			)
		}
		label := ui.NewLabel(ui.Label{
			Text: strings.Join(rows, "\n"),
			Font: render.Text{
				FontFilename: balance.ShellFontFilename,
				Size:         balance.ShellFontSizeSmall,
				Color:        render.White,
			},
		})
		label.SetBackground(render.RGBA(0, 0, 50, 150))
		label.Compute(e)
		label.Present(e, render.Point{
			X: p.X + S.W - label.Size().W - w.BoxThickness(1),
			Y: p.Y + w.BoxThickness(1),
		})
	}
}

// drawActors superimposes the actors on top of the drawing.
func (w *Canvas) drawActors(e render.Engine, p render.Point) {
	var (
		Viewport = w.ViewportRelative()
		S        = w.Size()
	)

	// See if each Actor is in range of the Viewport.
	for _, a := range w.actors {
		var (
			actor      = a.Actor     // Static Actor instance from Level file, DO NOT CHANGE
			can        = a.Canvas    // Canvas widget that draws the actor
			actorPoint = actor.Point // XXX TODO: DO NOT CHANGE
			actorSize  = can.Size()
		)

		// Create a box of World Coordinates that this actor occupies. The
		// Actor X,Y from level data is already a World Coordinate;
		// accomodate for the size of the Actor.
		actorBox := render.Rect{
			X: actorPoint.X,
			Y: actorPoint.Y,
			W: actorSize.W,
			H: actorSize.H,
		}

		// Is any part of the actor visible?
		if !Viewport.Intersects(actorBox) {
			continue // not visible on screen
		}

		drawAt := render.Point{
			X: p.X + w.Scroll.X + actorPoint.X + w.BoxThickness(1),
			Y: p.Y + w.Scroll.Y + actorPoint.Y + w.BoxThickness(1),
		}
		resizeTo := actorSize

		// XXX TODO: when an Actor hits the left or top edge and shrinks,
		// scrolling to offset that shrink is currently hard to solve.
		scrollTo := render.Origin

		// Handle cropping and scaling if this Actor's canvas can't be
		// completely visible within the parent.
		if drawAt.X+resizeTo.W > p.X+S.W {
			// Hitting the right edge, shrunk the width now.
			delta := (drawAt.X + resizeTo.W) - (p.X + S.W)
			resizeTo.W -= delta
		} else if drawAt.X < p.X {
			// Hitting the left edge. Cap the X coord and shrink the width.
			delta := p.X - drawAt.X // positive number
			drawAt.X = p.X
			// scrollTo.X -= delta // TODO
			resizeTo.W -= delta
		}

		if drawAt.Y+resizeTo.H > p.Y+S.H {
			// Hitting the bottom edge, shrink the height.
			delta := (drawAt.Y + resizeTo.H) - (p.Y + S.H)
			resizeTo.H -= delta
		} else if drawAt.Y < p.Y {
			// Hitting the top edge. Cap the Y coord and shrink the height.
			delta := p.Y - drawAt.Y
			drawAt.Y = p.Y
			// scrollTo.Y -= delta // TODO
			resizeTo.H -= delta
		}

		if resizeTo != actorSize {
			can.Resize(resizeTo)
			can.ScrollTo(scrollTo)
		}
		can.Present(e, drawAt)

		// Clean up the canvas size and offset.
		can.Resize(actorSize) // restore original size in case cropped
		can.ScrollTo(render.Origin)
	}
}
