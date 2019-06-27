package canvas

import (
	"syscall/js"

	"git.kirsle.net/apps/doodle/lib/render"
)

// Methods here implement the drawing functions of the render.Engine

// Clear the canvas to a certain color.
func (e *Engine) Clear(color render.Color) {
	e.canvas.ctx2d.Set("fillStyle", color.ToHex())
	e.canvas.ctx2d.Call("fillRect", 0, 0, e.width, e.height)
}

// SetTitle sets the window title.
func (e *Engine) SetTitle(title string) {
	js.Global().Get("document").Set("title", title)
}

// DrawPoint draws a pixel.
func (e *Engine) DrawPoint(color render.Color, point render.Point) {
	e.canvas.ctx2d.Set("fillStyle", color.ToHex())
	e.canvas.ctx2d.Call("fillRect",
		int(point.X),
		int(point.Y),
		1,
		1,
	)
}

// DrawLine draws a line between two points.
func (e *Engine) DrawLine(color render.Color, a, b render.Point) {
	e.canvas.ctx2d.Set("fillStyle", color.ToHex())
	for pt := range render.IterLine2(a, b) {
		e.canvas.ctx2d.Call("fillRect",
			int(pt.X),
			int(pt.Y),
			1,
			1,
		)
	}
}

// DrawRect draws a rectangle.
func (e *Engine) DrawRect(color render.Color, rect render.Rect) {
	e.canvas.ctx2d.Set("strokeStyle", color.ToHex())
	e.canvas.ctx2d.Call("strokeRect",
		int(rect.X),
		int(rect.Y),
		int(rect.W),
		int(rect.H),
	)
}

// DrawBox draws a filled rectangle.
func (e *Engine) DrawBox(color render.Color, rect render.Rect) {
	e.canvas.ctx2d.Set("fillStyle", color.ToHex())
	e.canvas.ctx2d.Call("fillRect",
		int(rect.X),
		int(rect.Y),
		int(rect.W),
		int(rect.H),
	)
}
