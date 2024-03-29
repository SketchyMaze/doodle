package uix

import (
	"fmt"
	"strings"

	"git.kirsle.net/SketchyMaze/doodle/pkg/balance"
	"git.kirsle.net/SketchyMaze/doodle/pkg/log"
	"git.kirsle.net/SketchyMaze/doodle/pkg/sprites"
	"git.kirsle.net/go/render"
	"git.kirsle.net/go/ui"
)

// Present the canvas.
func (w *Canvas) Present(e render.Engine, p render.Point) {
	var (
		S        = w.Size()
		Viewport = w.Viewport()
		// Bezel    = render.NewRect(
		// 	p.X+w.Scroll.X+w.BoxThickness(1),
		// 	p.Y+w.Scroll.Y+w.BoxThickness(1),
		// )
		// zoomMultiplier = int(w.GetZoomMultiplier())
	)
	// w.MoveTo(p) // TODO: when uncommented the canvas will creep down the Workspace frame in EditorMode
	w.DrawBox(e, p)
	e.DrawBox(w.Background(), render.Rect{
		X: p.X + w.BoxThickness(1),
		Y: p.Y + w.BoxThickness(1),
		W: S.W - w.BoxThickness(2),
		H: S.H - w.BoxThickness(2),
	})

	// If we are an Actor canvas as part of a Level, get the absolute position of
	// the parent (Level) canvas so we can compare where the Actor is drawn on-screen
	// and detect if we are at the Top or Left edges of the parent, to crop and adjust
	// the texture accordingly.
	var ParentPosition render.Point
	if w.parent != nil {
		ParentPosition = ui.AbsolutePosition(w.parent)
	}

	// Draw the wallpaper.
	if w.wallpaper.Valid() {
		err := w.PresentWallpaper(e, p)
		if err != nil {
			log.Error(err.Error())
		}
	}

	// Scale the viewport to account for zoom level.
	if w.Zoom != 0 {
		// Zoomed out (level go tiny)
		// TODO: seems unstable as shit on Zoom In??
		Viewport.W = w.ZoomDivide(Viewport.W)
		Viewport.H = w.ZoomDivide(Viewport.H)
		if w.Zoom != 0 {
			Viewport.X = w.ZoomDivide(Viewport.X)
			Viewport.Y = w.ZoomDivide(Viewport.Y)
		}
	}

	// Disappearing chunks issue:
	// When at the top-left corner of a bounded level,
	// And Zoom=2 (200% zoom), Level Chunk Size=128,
	// When you scroll down exactly -128 pixels, the whole row
	// of chunks along the top edge of the viewport unload.
	// At -256, the next row of chunks unloads.
	// At -383, another row - it's creeping down the page with
	//    the top 1/3 of the level editor showing blank wallpaper
	// At -768, about 3/4 the level editor is blank
	//
	// It must think the upper 128px of chunks had left the
	// viewport when in fact the bottom half of them was still
	// in view, not respecting the 2X zoom level.
	//
	// Viewport is like: Rect<128,0,1058,721>
	// Level chunks would be:
	//   (0, 0) = (0,0, 127,127)      chunk A
	//   (1, 0) = (128,128, 255,255)  chunk B
	// At 2x zoom, Chunk A is still half on screen at -128 scroll
	// if w.Zoom > 0 {
	// 	Viewport.X = w.ZoomDivide(w.chunks.Size)
	// 	Viewport.Y = w.ZoomDivide(w.chunks.Size)
	// }
	// Seems resolved now?

	// Get the chunks in the viewport and cache their textures.
	for coord := range w.chunks.IterViewportChunks(Viewport) {
		if chunk, ok := w.chunks.GetChunk(coord); ok {
			var tex render.Texturer
			if w.MaskColor != render.Invisible {
				tex = chunk.TextureMasked(e, w.MaskColor)
			} else {
				tex = chunk.Texture(e)
			}

			// Zoom in the texture.
			var (
				texSize     = tex.Size()
				texSizeOrig = texSize
			)

			if w.Zoom != 0 {
				texSize.W = w.ZoomMultiply(texSize.W)
				texSize.H = w.ZoomMultiply(texSize.H)
			}
			src := render.Rect{
				W: texSize.W,
				H: texSize.H,
			}

			// If the source bitmap is already bigger than the Canvas widget
			// into which it will render, cap the source width and height.
			// This is especially useful for Doodad buttons because the drawing
			// is bigger than the button.
			if w.CroppedSize {
				// NOTE: this is a concern mainly for the Doodad Dropper so that
				// the doodads won't overflow the button size they appear in.
				if src.W > S.W {
					src.W = S.W
				}
				if src.H > S.H {
					src.H = S.H
				}
			}

			var size = int(chunk.Size)
			dst := render.Rect{
				X: p.X + w.Scroll.X + w.BoxThickness(1) + w.ZoomMultiply(coord.X*size),
				Y: p.Y + w.Scroll.Y + w.BoxThickness(1) + w.ZoomMultiply(coord.Y*size),

				// src.W and src.H will be AT MOST the full width and height of
				// a Canvas widget. Subtract the scroll offset to keep it bounded
				// visually on its right and bottom sides.
				W: src.W,
				H: src.H,
			}

			// TODO: all this shit is in TrimBox(), make it DRY

			// wtf? don't need all this code anymore??
			_ = ParentPosition
			/*

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
				if dst.X+src.W > p.X+S.W+w.BoxThickness(1) {
					// NOTE: delta is a negative number,
					// so it will subtract from the width.
					delta := (p.X + S.W - w.BoxThickness(1)) - (dst.W + dst.X)
					src.W += delta
					dst.W += delta
				}
				if dst.Y+src.H > p.Y+S.H+w.BoxThickness(1) {
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
				//
				// Note: the +w.BoxThickness works around a bug if the Actor Canvas has
				// a border on it (e.g. in the Actor/Link Tool mouse-over or debug setting)
				if dst.X == ParentPosition.X+w.BoxThickness(1) {
					// NOTE: delta is a positive number,
					// so it will add to the destination coordinates.
					delta := texSizeOrig.W - src.W
					dst.X = p.X + w.BoxThickness(1)
					src.X += delta
				}
				if dst.Y == ParentPosition.Y+w.BoxThickness(1) {
					delta := texSizeOrig.H - src.H
					dst.Y = p.Y + w.BoxThickness(1)
					src.Y += delta
				}

				// Trim the destination width so it doesn't overlap the Canvas border.
				if dst.W >= S.W-w.BoxThickness(1) {
					dst.W = S.W - w.BoxThickness(1)
				}

			*/

			// When zooming OUT, make sure the source rect is at least the
			// full size of the chunk texture; otherwise the ZoomMultiplies
			// above do correctly scale e.g. 128x128 to 64x64, but it only
			// samples the top-left 64x64 then and not the full texture so
			// it more crops it than scales it, but does fit it neatly with
			// its neighbors.
			if w.Zoom < 0 {
				src.W = texSizeOrig.W
				src.H = texSizeOrig.H
			}

			e.Copy(tex, src, dst)
		}
	}

	w.drawActors(e, p)
	w.presentStrokes(e)
	w.presentDoodadButtons(e)
	w.presentCursor(e)

	// Custom label in the canvas corner? (e.g. for Inventory item counts)
	if w.CornerLabel != "" {
		label := ui.NewLabel(ui.Label{
			Text: w.CornerLabel,
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
			Y: p.Y + S.H - label.Size().H - w.BoxThickness(1),
		})
	}

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

// Draw doodad buttons on mouseover in the level editor.
func (w *Canvas) presentDoodadButtons(e render.Engine) {
	if !w.ShowDoodadButtons || w.parent == nil {
		return
	}

	// Initialize the buttons the first time?
	if w.doodadButtonFrame == nil {
		var (
			img ui.Widget
			err error
		)

		img, err = sprites.LoadImage(e, balance.GearIcon)
		if err != nil {
			// Error loading sprite, make a fallback frame.
			frame := ui.NewFrame("Buttons")
			frame.Configure(ui.Config{
				Width:      balance.UICanvasDoodadButtonSize,
				Height:     balance.UICanvasDoodadButtonSize,
				Background: render.Green,
			})
			w.doodadButtonFrame = frame
		} else {
			w.doodadButtonFrame = img
		}

		w.doodadButtonFrame.Compute(e)
	}

	// log.Error("presentDoodadButtons: parentP=%s w at %s (abs %s) actor at %s draw at %s", parentP, w.Point(), P, actorPoint, drawAt)
	w.doodadButtonFrame.Present(e, w.doodadButtonRect().Point())
}

// Return the screen rectangle where the doodad buttons would draw.
// screenCords: pass true to get on-screen coords (ignores scroll offset)
func (w *Canvas) doodadButtonRect() render.Rect {
	if !w.ShowDoodadButtons || w.parent == nil {
		return render.Rect{}
	}

	var (
		parentP    = ui.AbsolutePosition(w.parent)
		scroll     = w.parent.Scroll
		actorPoint = w.actor.Position()
		actorSize  = w.Size()
		drawAt     = render.Point{
			X: parentP.X + scroll.X + actorPoint.X + actorSize.W - balance.UICanvasDoodadButtonSize - w.BoxThickness(1),
			Y: parentP.Y + scroll.Y + actorPoint.Y + w.BoxThickness(1),
		}
	)

	// If the doodad is Very Small so that its buttons take up a disproportionate
	// amount of its space, draw the buttons further to the right.
	if actorSize.W <= balance.UICanvasDoodadButtonSpaceNeeded {
		drawAt.X += balance.UICanvasDoodadButtonSize / 2
		drawAt.Y -= balance.UICanvasDoodadButtonSize / 2
	}

	return render.Rect{
		X: drawAt.X,
		Y: drawAt.Y,
		W: balance.UICanvasDoodadButtonSize,
		H: balance.UICanvasDoodadButtonSize,
	}
}
