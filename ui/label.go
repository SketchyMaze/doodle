package ui

import (
	"fmt"

	"git.kirsle.net/apps/doodle/render"
)

// Label is a simple text label widget.
type Label struct {
	BaseWidget
	width  int32
	height int32
	Text   render.Text
}

// NewLabel creates a new label.
func NewLabel(t render.Text) *Label {
	w := &Label{
		Text: t,
	}
	w.Configure(Config{
		Padding: 4,
	})
	w.IDFunc(func() string {
		return fmt.Sprintf("Label<%s>", w.Text.Text)
	})
	return w
}

// Compute the size of the label widget.
func (w *Label) Compute(e render.Engine) {
	rect, _ := e.ComputeTextRect(w.Text)
	w.Resize(render.Rect{
		W: rect.W + w.Padding(),
		H: rect.H + w.Padding(),
	})
	w.MoveTo(render.Point{
		X: rect.X + w.BoxThickness(1),
		Y: rect.Y + w.BoxThickness(1),
	})
}

// Present the label widget.
func (w *Label) Present(e render.Engine) {
	var (
		P      = w.Point()
		border = w.BoxThickness(1)
	)
	w.DrawBox(e)
	e.DrawText(w.Text, render.Point{
		X: P.X + border,
		Y: P.Y + border,
	})
}
