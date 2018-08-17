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

	grid         Grid
	pixelHistory []*Pixel
	lastPixel    *Pixel

	// We inherit the ui.Widget which manages the width and height.
	Scroll render.Point // Scroll offset for which parts of canvas are visible.
}

// NewCanvas initializes a Canvas widget.
func NewCanvas(editable bool) *Canvas {
	w := &Canvas{
		Editable: editable,
		Palette:  NewPalette(),
		grid:     Grid{},
	}
	w.setup()
	return w
}

// Load initializes the Canvas using an existing Palette and Grid.
func (w *Canvas) Load(p *Palette, g *Grid) {
	w.Palette = p
	w.grid = *g
}

// LoadFilename initializes the Canvas using a file on disk.
func (w *Canvas) LoadFilename(filename string) error {
	w.grid = Grid{}

	m, err := LoadJSON(filename)
	if err != nil {
		return err
	}

	for _, pixel := range m.Pixels {
		w.grid[pixel] = nil
	}
	w.Palette = m.Palette

	if len(w.Palette.Swatches) > 0 {
		w.SetSwatch(w.Palette.Swatches[0])
	}

	return nil
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
	log.Info("my territory")
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
		pixel := &Pixel{
			X:       ev.CursorX.Now - P.X + w.Scroll.X,
			Y:       ev.CursorY.Now - P.Y + w.Scroll.Y,
			Palette: w.Palette,
			Swatch:  w.Palette.ActiveSwatch,
		}

		// Append unique new pixels.
		if len(w.pixelHistory) == 0 || w.pixelHistory[len(w.pixelHistory)-1] != pixel {
			if lastPixel != nil {
				// Draw the pixels in between.
				if lastPixel != pixel {
					for point := range render.IterLine(lastPixel.X, lastPixel.Y, pixel.X, pixel.Y) {
						dot := &Pixel{
							X:       point.X,
							Y:       point.Y,
							Palette: lastPixel.Palette,
							Swatch:  lastPixel.Swatch,
						}
						w.grid[dot] = nil
					}
				}
			}

			w.lastPixel = pixel
			w.pixelHistory = append(w.pixelHistory, pixel)

			// Save in the pixel canvas map.
			w.grid[pixel] = nil
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

// Grid returns the underlying grid object.
func (w *Canvas) Grid() *Grid {
	return &w.grid
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

	for pixel := range w.grid {
		point := render.NewPoint(pixel.X, pixel.Y)
		if point.Inside(Viewport) {
			// This pixel is visible in the canvas, but offset it by the
			// scroll height.
			point.Add(render.Point{
				X: -Viewport.X,
				Y: -Viewport.Y,
			})
			color := pixel.Swatch.Color
			e.DrawPoint(color, render.Point{
				X: p.X + w.BoxThickness(1) + point.X,
				Y: p.Y + w.BoxThickness(1) + point.Y,
			})
		}
	}
}
