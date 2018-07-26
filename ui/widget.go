package ui

import (
	"git.kirsle.net/apps/doodle/render"
)

// BorderStyle options for widget.SetBorderStyle()
type BorderStyle string

// Styles for a widget border.
const (
	BorderSolid  BorderStyle = "solid"
	BorderRaised             = "raised"
	BorderSunken             = "sunken"
)

// Widget is a user interface element.
type Widget interface {
	Point() render.Point
	MoveTo(render.Point)
	MoveBy(render.Point)
	Size() render.Rect // Return the Width and Height of the widget.
	Resize(render.Rect)

	Handle(string, func(render.Point))
	Event(string, render.Point) // called internally to trigger an event

	// Thickness of the padding + border + outline.
	BoxThickness(multiplier int32) int32
	DrawBox(render.Engine)

	// Widget configuration getters.
	Padding() int32               // Padding
	SetPadding(int32)             //
	Background() render.Color     // Background color
	SetBackground(render.Color)   //
	Foreground() render.Color     // Foreground color
	SetForeground(render.Color)   //
	BorderStyle() BorderStyle     // Border style: none, raised, sunken
	SetBorderStyle(BorderStyle)   //
	BorderColor() render.Color    // Border color (default is Background)
	SetBorderColor(render.Color)  //
	BorderSize() int32            // Border size (default 0)
	SetBorderSize(int32)          //
	OutlineColor() render.Color   // Outline color (default Invisible)
	SetOutlineColor(render.Color) //
	OutlineSize() int32           // Outline size (default 0)
	SetOutlineSize(int32)         //

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
	width        int32
	height       int32
	point        render.Point
	padding      int32
	background   render.Color
	foreground   render.Color
	borderStyle  BorderStyle
	borderColor  render.Color
	borderSize   int32
	outlineColor render.Color
	outlineSize  int32
	handlers     map[string][]func(render.Point)
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

// BoxThickness returns the full sum of the padding, border and outline.
// m = multiplier, i.e., 1 or 2
func (w *BaseWidget) BoxThickness(m int32) int32 {
	if m == 0 {
		m = 1
	}
	return (w.Padding() * m) + (w.BorderSize() * m) + (w.OutlineSize() * m)
}

// DrawBox draws the border and outline.
func (w *BaseWidget) DrawBox(e render.Engine) {
	var (
		P           = w.Point()
		S           = w.Size()
		outline     = w.OutlineSize()
		border      = w.BorderSize()
		borderColor = w.BorderColor()
		highlight   = borderColor.Add(20, 20, 20, 0)
		shadow      = borderColor.Add(-20, -20, -20, 0)
		color       render.Color
		box         = render.Rect{
			X: P.X,
			Y: P.Y,
			W: S.W,
			H: S.H,
		}
	)

	// Draw the outline layer as the full size of the widget.
	e.DrawBox(w.OutlineColor(), render.Rect{
		X: P.X - outline,
		Y: P.Y - outline,
		W: S.W + (outline * 2),
		H: S.H + (outline * 2),
	})

	// Highlight on the top left edge.
	if w.BorderStyle() == BorderRaised {
		color = highlight
	} else if w.BorderStyle() == BorderSunken {
		color = shadow
	} else {
		color = borderColor
	}
	e.DrawBox(color, box)
	box.W = S.W

	// Shadow on the bottom right edge.
	box.X += border
	box.Y += border
	box.W -= border
	box.H -= border
	if w.BorderStyle() == BorderRaised {
		color = shadow
	} else if w.BorderStyle() == BorderSunken {
		color = highlight
	} else {
		color = borderColor
	}
	e.DrawBox(color.Add(-20, -20, -20, 0), box)

	// Background color of the button.
	box.W -= border
	box.H -= border
	// if w.hovering {
	// 	e.DrawBox(render.Yellow, box)
	// } else {
	e.DrawBox(color, box)
}

// Padding returns the padding width.
func (w *BaseWidget) Padding() int32 {
	return w.padding
}

// SetPadding sets the padding width.
func (w *BaseWidget) SetPadding(v int32) {
	w.padding = v
}

// Background returns the background color.
func (w *BaseWidget) Background() render.Color {
	return w.background
}

// SetBackground sets the color.
func (w *BaseWidget) SetBackground(c render.Color) {
	w.background = c
}

// Foreground returns the foreground color.
func (w *BaseWidget) Foreground() render.Color {
	return w.foreground
}

// SetForeground sets the color.
func (w *BaseWidget) SetForeground(c render.Color) {
	w.foreground = c
}

// BorderStyle returns the border style.
func (w *BaseWidget) BorderStyle() BorderStyle {
	return w.borderStyle
}

// SetBorderStyle sets the border style.
func (w *BaseWidget) SetBorderStyle(v BorderStyle) {
	w.borderStyle = v
}

// BorderColor returns the border color, or defaults to the background color.
func (w *BaseWidget) BorderColor() render.Color {
	if w.borderColor == render.Invisible {
		return w.Background()
	}
	return w.borderColor
}

// SetBorderColor sets the border color.
func (w *BaseWidget) SetBorderColor(c render.Color) {
	w.borderColor = c
}

// BorderSize returns the border thickness.
func (w *BaseWidget) BorderSize() int32 {
	return w.borderSize
}

// SetBorderSize sets the border thickness.
func (w *BaseWidget) SetBorderSize(v int32) {
	w.borderSize = v
}

// OutlineColor returns the background color.
func (w *BaseWidget) OutlineColor() render.Color {
	return w.outlineColor
}

// SetOutlineColor sets the color.
func (w *BaseWidget) SetOutlineColor(c render.Color) {
	w.outlineColor = c
}

// OutlineSize returns the outline thickness.
func (w *BaseWidget) OutlineSize() int32 {
	return w.outlineSize
}

// SetOutlineSize sets the outline thickness.
func (w *BaseWidget) SetOutlineSize(v int32) {
	w.outlineSize = v
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
