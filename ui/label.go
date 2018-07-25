package ui

import "git.kirsle.net/apps/doodle/render"

// Label is a simple text label widget.
type Label struct {
	BaseWidget
	width  int32
	height int32
	Text   render.Text
}

// NewLabel creates a new label.
func NewLabel(t render.Text) *Label {
	return &Label{
		Text: t,
	}
}

// Compute the size of the label widget.
func (w *Label) Compute(e render.Engine) {
	rect, err := e.ComputeTextRect(w.Text)
	w.Resize(rect)
	_ = rect
	_ = err
}

// Present the label widget.
func (w *Label) Present(e render.Engine) {
	e.DrawText(w.Text, w.Point())
}
