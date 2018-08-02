package ui

import (
	"fmt"

	"git.kirsle.net/apps/doodle/render"
	"git.kirsle.net/apps/doodle/ui/theme"
)

// CheckButton is a button that is bound to a boolean variable and stays clicked
// once pressed, until clicked again to release.
type CheckButton struct {
	Button
	BoolVar *bool
}

// NewCheckButton creates a new CheckButton.
func NewCheckButton(name string, boolVar *bool, child Widget) *CheckButton {
	w := &CheckButton{
		BoolVar: boolVar,
	}
	w.Button.child = child
	w.IDFunc(func() string {
		return fmt.Sprintf("CheckButton<%s %+v>", name, w.BoolVar)
	})

	var borderStyle BorderStyle = BorderRaised
	if w.BoolVar != nil {
		if *w.BoolVar == true {
			borderStyle = BorderSunken
		}
	}

	w.Configure(Config{
		Padding:      4,
		BorderSize:   2,
		BorderStyle:  borderStyle,
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
	})

	w.Handle("MouseDown", func(p render.Point) {
		if w.BoolVar != nil {
			if *w.BoolVar {
				*w.BoolVar = false
				w.SetBorderStyle(BorderRaised)
			} else {
				*w.BoolVar = true
				w.SetBorderStyle(BorderSunken)
			}
		}
	})

	return w
}
