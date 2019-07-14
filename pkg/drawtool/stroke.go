package drawtool

import "git.kirsle.net/apps/doodle/lib/render"

/*
Stroke holds temporary pixel data with a shape and color.

It is used for myriad purposes:

- As a staging area for drawing new pixels to the drawing without committing
  them until completed.
- As a unit of work for the Undo/Redo History when editing a drawing.
- As imaginary visual lines superimposed on top of a drawing, for example to
  visualize the link between two doodads or to draw collision hitboxes and other
  debug lines to the drawing.
*/
type Stroke struct {
	ID        int // Unique ID per each stroke
	Shape     Shape
	Color     render.Color
	Thickness int         // 0 = 1px; thickness creates a box N pixels away from each point
	ExtraData interface{} // arbitrary storage for extra data to attach

	// Start and end points for Lines, Rectangles, etc.
	PointA render.Point
	PointB render.Point

	// Array of points for Freehand shapes.
	Points    []render.Point
	uniqPoint map[render.Point]interface{} // deduplicate points added

	// Storage space to recall the previous values of points that were replaced,
	// especially for the Undo/Redo History tool. When the uix.Canvas commits a
	// Stroke to the level data, any pixel that has replaced an existing color
	// will cache its color here, so we can easily page forwards and backwards
	// in history and not lose data.
	//
	// The data is implementation defined and controlled by the caller. This
	// package does not modify OriginalPoints or do anything with it.
	OriginalPoints map[render.Point]interface{}
}

var nextStrokeID int

// NewStroke initializes a new Stroke with a shape and a color.
func NewStroke(shape Shape, color render.Color) *Stroke {
	nextStrokeID++
	return &Stroke{
		ID:    nextStrokeID,
		Shape: shape,
		Color: color,

		// Initialize data structures.
		Points:    []render.Point{},
		uniqPoint: map[render.Point]interface{}{},

		OriginalPoints: map[render.Point]interface{}{},
	}
}

// Copy returns a duplicate of the Stroke reference.
func (s *Stroke) Copy() *Stroke {
	nextStrokeID++
	return &Stroke{
		ID:        nextStrokeID,
		Shape:     s.Shape,
		Color:     s.Color,
		Thickness: s.Thickness,
		ExtraData: s.ExtraData,

		Points:    []render.Point{},
		uniqPoint: map[render.Point]interface{}{},
	}
}

// IterPoints returns an iterator of points represented by the stroke.
//
// For a Line, returns all of the points between PointA and PointB. For freehand,
// returns every point added to the stroke.
func (s *Stroke) IterPoints() chan render.Point {
	ch := make(chan render.Point)
	go func() {
		switch s.Shape {
		case Eraser:
			fallthrough
		case Freehand:
			for _, point := range s.Points {
				ch <- point
			}
		case Line:
			for point := range render.IterLine(s.PointA, s.PointB) {
				ch <- point
			}
		case Rectangle:
			for point := range render.IterRect(s.PointA, s.PointB) {
				ch <- point
			}
		case Ellipse:
			for point := range render.IterEllipse2(s.PointA, s.PointB) {
				ch <- point
			}
		}
		close(ch)
	}()
	return ch
}

// IterThickPoints iterates over the points and yield Rects of each one.
func (s *Stroke) IterThickPoints() chan render.Rect {
	ch := make(chan render.Rect)
	go func() {
		for pt := range s.IterPoints() {
			ch <- render.Rect{
				X: pt.X - int32(s.Thickness),
				Y: pt.Y - int32(s.Thickness),
				W: int32(s.Thickness) * 2,
				H: int32(s.Thickness) * 2,
			}
		}
		close(ch)
	}()
	return ch
}

// AddPoint adds a point to the stroke, for freehand shapes.
func (s *Stroke) AddPoint(p render.Point) {
	if _, ok := s.uniqPoint[p]; ok {
		return
	}
	s.uniqPoint[p] = nil
	s.Points = append(s.Points, p)
}
