package uix

import (
	"fmt"
	"strings"

	"git.kirsle.net/go/render"
	"git.kirsle.net/go/ui"
	"git.kirsle.net/apps/doodle/pkg/balance"
	"git.kirsle.net/apps/doodle/pkg/log"
)

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

	// Draw the wallpaper.
	if w.wallpaper.Valid() {
		err := w.PresentWallpaper(e, p)
		if err != nil {
			log.Error(err.Error())
		}
	}

	// Get the chunks in the viewport and cache their textures.
	for coord := range w.chunks.IterViewportChunks(Viewport) {
		if chunk, ok := w.chunks.GetChunk(coord); ok {
			var tex render.Texturer
			if w.MaskColor != render.Invisible {
				tex = chunk.TextureMasked(e, w.MaskColor)
			} else {
				tex = chunk.Texture(e)
			}
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

			// TODO: all this shit is in TrimBox(), make it DRY

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
	w.presentStrokes(e)
	w.presentCursor(e)

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

		// Draw the actor's position details.
		// LP = Level Position, where the Actor starts at in the level data
		// WP = World Position, the Actor's current position in the level
		if w.actor != nil {
			rows = append(rows,
				fmt.Sprintf("LP=%s", w.actor.Actor.Point),
				fmt.Sprintf("WP=%s", w.actor.Position()),
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
