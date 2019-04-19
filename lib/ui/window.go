package ui

import (
	"fmt"

	"git.kirsle.net/apps/doodle/lib/render"
)

// Window is a frame with a title bar.
type Window struct {
	BaseWidget
	Title  string
	Active bool

	// Private widgets.
	body     *Frame
	titleBar *Label
	content  *Frame
}

// NewWindow creates a new window.
func NewWindow(title string) *Window {
	w := &Window{
		Title: title,
		body:  NewFrame("body:" + title),
	}
	w.IDFunc(func() string {
		return fmt.Sprintf("Window<%s>",
			w.Title,
		)
	})

	w.body.Configure(Config{
		Background:  render.Grey,
		BorderSize:  2,
		BorderStyle: BorderRaised,
	})

	// Title bar widget.
	titleBar := NewLabel(Label{
		TextVariable: &w.Title,
		Font: render.Text{
			Color:   render.White,
			Size:    10,
			Stroke:  render.DarkBlue,
			Padding: 2,
		},
	})
	titleBar.Configure(Config{
		Background: render.Blue,
	})
	w.body.Pack(titleBar, Pack{
		Side: Top,
		Fill: FillX,
	})
	w.titleBar = titleBar

	// Window content frame.
	content := NewFrame("content:" + title)
	content.Configure(Config{
		Background: render.Grey,
	})
	w.body.Pack(content, Pack{
		Side: Top,
		Fill: FillBoth,
	})
	w.content = content

	return w
}

// Children returns the window's child widgets.
func (w *Window) Children() []Widget {
	return []Widget{
		w.body,
	}
}

// TitleBar returns the title bar widget.
func (w *Window) TitleBar() *Label {
	return w.titleBar
}

// Frame returns the content frame of the window.
func (w *Window) Frame() *Frame {
	return w.content
}

// Configure the widget. Color and style changes are passed down to the inner
// content frame of the window.
func (w *Window) Configure(C Config) {
	w.BaseWidget.Configure(C)
	w.body.Configure(C)

	// Don't pass dimensions down any further than the body.
	C.Width = 0
	C.Height = 0
	w.content.Configure(C)
}

// ConfigureTitle configures the title bar widget.
func (w *Window) ConfigureTitle(C Config) {
	w.titleBar.Configure(C)
}

// Compute the window.
func (w *Window) Compute(e render.Engine) {
	var size = w.Size()

	w.body.Compute(e)

	// Assign a manual Height to the title bar using its naturally computed
	// height, but leave the Width empty so the frame packer can stretch it
	// horizontally.
	w.titleBar.Configure(Config{
		Height: w.titleBar.Size().H,
	})

	// Shrink down the content frame to leave room for the title bar.
	w.content.Resize(render.Rect{
		W: size.W - w.BoxThickness(2) - w.titleBar.BoxThickness(2),
		H: size.H - w.titleBar.Size().H - w.BoxThickness(4) -
			((w.titleBar.Font.Padding + w.titleBar.Font.PadY) * 2),
	})
}

// Present the window.
func (w *Window) Present(e render.Engine, P render.Point) {
	w.body.Present(e, P)
}

// Pack a widget into the window's frame.
func (w *Window) Pack(child Widget, config ...Pack) {
	w.content.Pack(child, config...)
}
