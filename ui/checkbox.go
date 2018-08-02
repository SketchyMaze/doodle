package ui

import "git.kirsle.net/apps/doodle/render"

// Checkbox combines a CheckButton with a widget like a Label.
type Checkbox struct {
	Frame
	button *CheckButton
	child  Widget
}

// NewCheckbox creates a new Checkbox.
func NewCheckbox(name string, boolVar *bool, child Widget) *Checkbox {
	// Our custom checkbutton widget.
	mark := NewFrame(name + "_mark")

	w := &Checkbox{
		button: NewCheckButton(name+"_button", boolVar, mark),
		child:  child,
	}
	w.Frame.Setup()

	// Forward clicks on the child widget to the CheckButton.
	for _, e := range []string{"MouseOver", "MouseOut", "MouseUp", "MouseDown"} {
		func(e string) {
			w.child.Handle(e, func(p render.Point) {
				w.button.Event(e, p)
			})
		}(e)
	}

	w.Pack(w.button, Pack{
		Anchor: W,
	})
	w.Pack(w.child, Pack{
		Anchor: W,
	})

	return w
}

// Supervise the checkbutton inside the widget.
func (w *Checkbox) Supervise(s *Supervisor) {
	s.Add(w.button)
	s.Add(w.child)
}
