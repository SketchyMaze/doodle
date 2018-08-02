package ui

import (
	"errors"
	"fmt"

	"git.kirsle.net/apps/doodle/render"
	"git.kirsle.net/apps/doodle/ui/theme"
)

// Button is a clickable button.
type Button struct {
	BaseWidget
	child Widget

	// Private options.
	hovering bool
	clicked  bool
}

// NewButton creates a new Button.
func NewButton(name string, child Widget) *Button {
	w := &Button{
		child: child,
	}
	w.IDFunc(func() string {
		return fmt.Sprintf("Button<%s>", name)
	})

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

	return w
}

// Compute the size of the button.
func (w *Button) Compute(e render.Engine) {
	// Compute the size of the inner widget first.
	w.child.Compute(e)

	// Auto-resize only if we haven't been given a fixed size.
	if !w.FixedSize() {
		size := w.child.Size()
		w.Resize(render.Rect{
			W: size.W + w.BoxThickness(2),
			H: size.H + w.BoxThickness(2),
		})
	}
}

// SetText conveniently sets the button text, for Label children only.
func (w *Button) SetText(text string) error {
	if label, ok := w.child.(*Label); ok {
		label.Text.Text = text
	}
	return errors.New("child is not a Label widget")
}

// Present the button.
func (w *Button) Present(e render.Engine) {
	w.Compute(e)
	var (
		P         = w.Point()
		S         = w.Size()
		ChildSize = w.child.Size()
	)

	// Draw the widget's border and everything.
	w.DrawBox(e)

	// Offset further if we are currently sunken.
	var clickOffset int32
	if w.clicked {
		clickOffset++
	}

	// Where to place the child widget.
	moveTo := render.Point{
		X: P.X + w.BoxThickness(1) + clickOffset,
		Y: P.Y + w.BoxThickness(1) + clickOffset,
	}

	// If we're bigger than we need to be, center the child widget.
	if S.Bigger(ChildSize) {
		moveTo.X = P.X + (S.W / 2) - (ChildSize.W / 2)
	}
	_ = S
	_ = ChildSize

	// Draw the text label inside.
	w.child.MoveTo(moveTo)
	w.child.Present(e)
}
