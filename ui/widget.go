package ui

import "git.kirsle.net/apps/doodle/render"

// Widget is a user interface element.
type Widget interface {
	Point() render.Point
	MoveTo(render.Point)
	MoveBy(render.Point)
	Size() render.Rect // Return the Width and Height of the widget.
	Resize(render.Rect)

	Handle(string, func(render.Point))
	Event(string, render.Point) // called internally to trigger an event

	// Run any render computations; by the end the widget must know its
	// Width and Height. For example the Label widget will render itself onto
	// an SDL Surface and then it will know its bounding box, but not before.
	Compute(render.Engine)

	// Render the final widget onto the drawing engine.
	Present(render.Engine)
}

// BaseWidget holds common functionality for all widgets, such as managing
// their widths and heights.
type BaseWidget struct {
	width    int32
	height   int32
	point    render.Point
	handlers map[string][]func(render.Point)
}

// Point returns the X,Y position of the widget on the window.
func (w *BaseWidget) Point() render.Point {
	return w.point
}

// MoveTo updates the X,Y position to the new point.
func (w *BaseWidget) MoveTo(v render.Point) {
	w.point = v
}

// MoveBy adds the X,Y values to the widget's current position.
func (w *BaseWidget) MoveBy(v render.Point) {
	w.point.X += v.X
	w.point.Y += v.Y
}

// Size returns the box with W and H attributes containing the size of the
// widget. The X,Y attributes of the box are ignored and zero.
func (w *BaseWidget) Size() render.Rect {
	return render.Rect{
		W: w.width,
		H: w.height,
	}
}

// Resize sets the size of the widget to the .W and .H attributes of a rect.
func (w *BaseWidget) Resize(v render.Rect) {
	w.width = v.W
	w.height = v.H
}

// Event is called internally by Doodle to trigger an event.
func (w *BaseWidget) Event(name string, p render.Point) {
	if handlers, ok := w.handlers[name]; ok {
		for _, fn := range handlers {
			fn(p)
		}
	}
}

// Handle an event in the widget.
func (w *BaseWidget) Handle(name string, fn func(render.Point)) {
	if w.handlers == nil {
		w.handlers = map[string][]func(render.Point){}
	}

	if _, ok := w.handlers[name]; !ok {
		w.handlers[name] = []func(render.Point){}
	}

	w.handlers[name] = append(w.handlers[name], fn)
}

// OnMouseOut should be overridden on widgets who want this event.
func (w *BaseWidget) OnMouseOut(render.Point) {}
