package drawtool

// Shape of a stroke line.
type Shape int

// Shape values.
const (
	Freehand Shape = iota
	Line
	Rectangle
	Eraser // not really a shape but communicates the intention
)
