// Package sdl provides an SDL2 renderer for Doodle.
package sdl

import (
	"git.kirsle.net/apps/doodle/lib/render"
	"github.com/veandco/go-sdl2/sdl"
)

// Clear the canvas and set this color.
func (r *Renderer) Clear(color render.Color) {
	if color != r.lastColor {
		r.renderer.SetDrawColor(color.Red, color.Green, color.Blue, color.Alpha)
	}
	r.renderer.Clear()
}

// DrawPoint puts a color at a pixel.
func (r *Renderer) DrawPoint(color render.Color, point render.Point) {
	if color != r.lastColor {
		r.renderer.SetDrawColor(color.Red, color.Green, color.Blue, color.Alpha)
	}
	r.renderer.DrawPoint(point.X, point.Y)
}

// DrawLine draws a line between two points.
func (r *Renderer) DrawLine(color render.Color, a, b render.Point) {
	if color != r.lastColor {
		r.renderer.SetDrawColor(color.Red, color.Green, color.Blue, color.Alpha)
	}
	r.renderer.DrawLine(a.X, a.Y, b.X, b.Y)
}

// DrawRect draws a rectangle.
func (r *Renderer) DrawRect(color render.Color, rect render.Rect) {
	if color != r.lastColor {
		r.renderer.SetDrawColor(color.Red, color.Green, color.Blue, color.Alpha)
	}
	r.renderer.DrawRect(&sdl.Rect{
		X: rect.X,
		Y: rect.Y,
		W: rect.W,
		H: rect.H,
	})
}

// DrawBox draws a filled rectangle.
func (r *Renderer) DrawBox(color render.Color, rect render.Rect) {
	if color != r.lastColor {
		r.renderer.SetDrawColor(color.Red, color.Green, color.Blue, color.Alpha)
	}
	r.renderer.FillRect(&sdl.Rect{
		X: rect.X,
		Y: rect.Y,
		W: rect.W,
		H: rect.H,
	})
}
