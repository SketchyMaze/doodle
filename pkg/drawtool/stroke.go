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
	ExtraData interface{} // arbitrary storage for extra data to attach

	// Start and end points for Lines, Rectangles, etc.
	PointA render.Point
	PointB render.Point

	// Array of points for Freehand shapes.
	Points    []render.Point
	uniqPoint map[render.Point]interface{} // deduplicate points added
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
	}
}

// Copy returns a duplicate of the Stroke reference.
func (s *Stroke) Copy() *Stroke {
	nextStrokeID++
	return &Stroke{
		ID:    nextStrokeID,
		Shape: s.Shape,
		Color: s.Color,

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
		case Freehand:
			for _, point := range s.Points {
				ch <- point
			}
		case Line:
			for point := range render.IterLine2(s.PointA, s.PointB) {
				ch <- point
			}
		case Rectangle:
			for point := range render.IterRect(s.PointA, s.PointB) {
				ch <- point
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
