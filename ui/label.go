package ui

import (
	"fmt"

	"git.kirsle.net/apps/doodle/render"
)

// Label is a simple text label widget.
type Label struct {
	BaseWidget

	// Configurable fields for the constructor.
	Text         string
	TextVariable *string
	Font         render.Text

	width  int32
	height int32
}

// NewLabel creates a new label.
func NewLabel(c Label) *Label {
	w := &Label{
		Text:         c.Text,
		TextVariable: c.TextVariable,
		Font:         c.Font,
	}
	w.IDFunc(func() string {
		return fmt.Sprintf("Label<%s>", w.text().Text)
	})
	return w
}

// text returns the label's displayed text, coming from the TextVariable if
// available or else the Text attribute instead.
func (w *Label) text() render.Text {
	if w.TextVariable != nil {
		w.Font.Text = *w.TextVariable
		return w.Font
	}
	w.Font.Text = w.Text
	return w.Font
}

// Compute the size of the label widget.
func (w *Label) Compute(e render.Engine) {
	rect, _ := e.ComputeTextRect(w.text())

	if !w.FixedSize() {
		w.resizeAuto(render.Rect{
			W: rect.W + (w.Font.Padding * 2),
			H: rect.H + (w.Font.Padding * 2),
		})
	}

	w.MoveTo(render.Point{
		X: rect.X + w.BoxThickness(1),
		Y: rect.Y + w.BoxThickness(1),
	})
}

// Present the label widget.
func (w *Label) Present(e render.Engine, P render.Point) {
	border := w.BoxThickness(1)

	w.DrawBox(e, P)
	e.DrawText(w.text(), render.Point{
		X: P.X + border + w.Font.Padding,
		Y: P.Y + border + w.Font.Padding,
	})
}
