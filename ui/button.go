package ui

import (
	"fmt"

	"git.kirsle.net/apps/doodle/render"
	"git.kirsle.net/apps/doodle/ui/theme"
)

// Button is a clickable button.
type Button struct {
	BaseWidget
	Label Label

	// Private options.
	hovering bool
	clicked  bool
}

// NewButton creates a new Button.
func NewButton(label Label) *Button {
	w := &Button{
		Label: label,
	}

	w.Configure(Config{
		Padding:      4,
		BorderSize:   2,
		BorderStyle:  BorderRaised,
		OutlineSize:  1,
		OutlineColor: theme.ButtonOutlineColor,
		Background:   theme.ButtonBackgroundColor,
	})

	w.Handle("MouseOver", func(p render.Point) {
		w.hovering = true
		w.SetBackground(theme.ButtonHoverColor)
	})
	w.Handle("MouseOut", func(p render.Point) {
		w.hovering = false
		w.SetBackground(theme.ButtonBackgroundColor)
	})

	w.Handle("MouseDown", func(p render.Point) {
		w.clicked = true
		w.SetBorderStyle(BorderSunken)
	})
	w.Handle("MouseUp", func(p render.Point) {
		w.clicked = false
		w.SetBorderStyle(BorderRaised)
	})

	w.IDFunc(func() string {
		return fmt.Sprintf("Button<%s>", w.Label.Text.Text)
	})

	return w
}

// SetText quickly changes the text of the label.
func (w *Button) SetText(text string) {
	w.Label.Text.Text = text
}

// Compute the size of the button.
func (w *Button) Compute(e render.Engine) {
	// Compute the size of the inner widget first.
	w.Label.Compute(e)
	size := w.Label.Size()
	w.Resize(render.Rect{
		W: size.W + w.BoxThickness(2),
		H: size.H + w.BoxThickness(2),
	})
}

// Present the button.
func (w *Button) Present(e render.Engine) {
	w.Compute(e)
	P := w.Point()

	// Draw the widget's border and everything.
	w.DrawBox(e)

	// Offset further if we are currently sunken.
	var clickOffset int32
	if w.clicked {
		clickOffset++
	}

	// Draw the text label inside.
	w.Label.MoveTo(render.Point{
		X: P.X + w.BoxThickness(1) + clickOffset,
		Y: P.Y + w.BoxThickness(1) + clickOffset,
	})
	w.Label.Present(e)
}
