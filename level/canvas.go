package level

import (
	"git.kirsle.net/apps/doodle/balance"
	"git.kirsle.net/apps/doodle/events"
	"git.kirsle.net/apps/doodle/render"
	"git.kirsle.net/apps/doodle/ui"
)

// Canvas is a custom ui.Widget that manages a single drawing.
type Canvas struct {
	ui.Frame
	Palette *Palette

	// Set to true to allow clicking to edit this canvas.
	Editable bool

	chunks       *Chunker
	pixelHistory []*Pixel
	lastPixel    *Pixel

	// We inherit the ui.Widget which manages the width and height.
	Scroll render.Point // Scroll offset for which parts of canvas are visible.
}

// NewCanvas initializes a Canvas widget.
func NewCanvas(size int, editable bool) *Canvas {
	w := &Canvas{
		Editable: editable,
		Palette:  NewPalette(),
		chunks:   NewChunker(size),
	}
	w.setup()
	return w
}

// Load initializes the Canvas using an existing Palette and Grid.
func (w *Canvas) Load(p *Palette, g *Chunker) {
	w.Palette = p
	w.chunks = g

	if len(w.Palette.Swatches) > 0 {
		w.SetSwatch(w.Palette.Swatches[0])
	}
}

// LoadLevel initializes a Canvas from a Level object.
func (w *Canvas) LoadLevel(level *Level) {
	w.Load(level.Palette, level.Chunker)
}

// SetSwatch changes the currently selected swatch for editing.
func (w *Canvas) SetSwatch(s *Swatch) {
	w.Palette.ActiveSwatch = s
}

// setup common configs between both initializers of the canvas.
func (w *Canvas) setup() {
	w.SetBackground(render.White)
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
	var (
		P = w.Point()
		_ = P
	)

	// Arrow keys to scroll the view.
	scrollBy := render.Point{}
	if ev.Right.Now {
		scrollBy.X += balance.CanvasScrollSpeed
	} else if ev.Left.Now {
		scrollBy.X -= balance.CanvasScrollSpeed
	}
	if ev.Down.Now {
		scrollBy.Y += balance.CanvasScrollSpeed
	} else if ev.Up.Now {
		scrollBy.Y -= balance.CanvasScrollSpeed
	}
	if !scrollBy.IsZero() {
		w.ScrollBy(scrollBy)
	}

	// Only care if the cursor is over our space.
	cursor := render.NewPoint(ev.CursorX.Now, ev.CursorY.Now)
	if !cursor.Inside(w.Rect()) {
		return nil
	}

	// If no swatch is active, do nothing with mouse clicks.
	if w.Palette.ActiveSwatch == nil {
		return nil
	}

	// Clicking? Log all the pixels while doing so.
	if ev.Button1.Now {
		// log.Warn("Button1: %+v", ev.Button1)
		lastPixel := w.lastPixel
		cursor := render.Point{
			X: ev.CursorX.Now - P.X + w.Scroll.X,
			Y: ev.CursorY.Now - P.Y + w.Scroll.Y,
		}
		pixel := &Pixel{
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
func (w *Canvas) Viewport() render.Rect {
	var S = w.Size()
	return render.Rect{
		X: w.Scroll.X,
		Y: w.Scroll.Y,
		W: S.W - w.BoxThickness(2),
		H: S.H - w.BoxThickness(2),
	}
}

// Chunker returns the underlying Chunker object.
func (w *Canvas) Chunker() *Chunker {
	return w.chunks
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
	w.MoveTo(p)
	w.DrawBox(e, p)
	e.DrawBox(w.Background(), render.Rect{
		X: p.X + w.BoxThickness(1),
		Y: p.Y + w.BoxThickness(1),
		W: S.W - w.BoxThickness(2),
		H: S.H - w.BoxThickness(2),
	})

	for px := range w.chunks.IterViewport(Viewport) {
		// This pixel is visible in the canvas, but offset it by the
		// scroll height.
		px.X -= Viewport.X
		px.Y -= Viewport.Y
		color := px.Swatch.Color
		e.DrawPoint(color, render.Point{
			X: p.X + w.BoxThickness(1) + px.X,
			Y: p.Y + w.BoxThickness(1) + px.Y,
		})
	}
}
