package canvas

import (
	"fmt"
	"syscall/js"

	"git.kirsle.net/apps/doodle/lib/render"
)

// Methods here implement the drawing functions of the render.Engine

// RGBA turns a color into CSS RGBA string.
func RGBA(c render.Color) string {
	return fmt.Sprintf("rgba(%d,%d,%d,%f)",
		c.Red,
		c.Green,
		c.Blue,
		float64(c.Alpha)/255,
	)
}

// Clear the canvas to a certain color.
func (e *Engine) Clear(color render.Color) {
	e.canvas.ctx2d.Set("fillStyle", RGBA(color))
	e.canvas.ctx2d.Call("fillRect", 0, 0, e.width, e.height)
}

// SetTitle sets the window title.
func (e *Engine) SetTitle(title string) {
	js.Global().Get("document").Set("title", title)
}

// DrawPoint draws a pixel.
func (e *Engine) DrawPoint(color render.Color, point render.Point) {
	e.canvas.ctx2d.Set("fillStyle", RGBA(color))
	e.canvas.ctx2d.Call("fillRect",
		int(point.X),
		int(point.Y),
		1,
		1,
	)
}

// DrawLine draws a line between two points.
func (e *Engine) DrawLine(color render.Color, a, b render.Point) {
	e.canvas.ctx2d.Set("fillStyle", RGBA(color))
	for pt := range render.IterLine(a, b) {
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
	e.canvas.ctx2d.Set("strokeStyle", RGBA(color))
	e.canvas.ctx2d.Call("strokeRect",
		int(rect.X),
		int(rect.Y),
		int(rect.W),
		int(rect.H),
	)
}

// DrawBox draws a filled rectangle.
func (e *Engine) DrawBox(color render.Color, rect render.Rect) {
	e.canvas.ctx2d.Set("fillStyle", RGBA(color))
	e.canvas.ctx2d.Call("fillRect",
		int(rect.X),
		int(rect.Y),
		int(rect.W),
		int(rect.H),
	)
}
