package ui

import (
	"git.kirsle.net/apps/doodle/render"
	"git.kirsle.net/apps/doodle/ui/theme"
)

// Button is a clickable button.
type Button struct {
	BaseWidget
	Label   Label
	Padding int32
	Border  int32
	Outline int32

	// Color options.
	Background     render.Color
	HighlightColor render.Color
	ShadowColor    render.Color
	OutlineColor   render.Color
}

// NewButton creates a new Button.
func NewButton(label Label) *Button {
	return &Button{
		Label:   label,
		Padding: 4, // TODO magic number
		Border:  2,
		Outline: 1,

		// Default theme colors.
		Background:     theme.ButtonBackgroundColor,
		HighlightColor: theme.ButtonHighlightColor,
		ShadowColor:    theme.ButtonShadowColor,
		OutlineColor:   theme.ButtonOutlineColor,
	}
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
		W: size.W + (w.Padding * 2) + (w.Border * 2) + (w.Outline * 2),
		H: size.H + (w.Padding * 2) + (w.Border * 2) + (w.Outline * 2),
	})
}

// Present the button.
func (w *Button) Present(e render.Engine) {
	w.Compute(e)
	P := w.Point()
	S := w.Size()

	box := render.Rect{
		X: P.X,
		Y: P.Y,
		W: S.W,
		H: S.H,
	}

	// Draw the outline layer as the full size of the widget.
	e.DrawBox(w.OutlineColor, render.Rect{
		X: P.X - w.Outline,
		Y: P.Y - w.Outline,
		W: S.W + (w.Outline * 2),
		H: S.H + (w.Outline * 2),
	})

	// Highlight on the top left edge.
	e.DrawBox(w.HighlightColor, box)
	box.W = S.W

	// Shadow on the bottom right edge.
	box.X += w.Border
	box.Y += w.Border
	box.W -= w.Border
	box.H -= w.Border
	e.DrawBox(w.ShadowColor, box)

	// Background color of the button.
	box.W -= w.Border
	box.H -= w.Border
	e.DrawBox(w.Background, box)

	// Draw the text label inside.
	w.Label.SetPoint(render.Point{
		X: P.X + w.Padding + w.Border + w.Outline,
		Y: P.Y + w.Padding + w.Border + w.Outline,
	})
	w.Label.Present(e)
}
