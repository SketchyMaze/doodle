package collision

import (
	"errors"
	"math"

	"git.kirsle.net/apps/doodle/lib/render"
)

// BoxCollision holds the result of a collision BetweenBoxes.
type BoxCollision struct {
	// A and B are the indexes of the boxes sent to BetweenBoxes.
	A int
	B int

	// Overlap is the rect of how the boxes overlap.
	Overlap render.Rect
}

// IndexTuple holds two integers used as array indexes.
type IndexTuple [2]int

// BetweenBoxes checks if there is a collision between any
// two bounding rectangles.
//
// This returns a generator that spits out indexes of the
// intersecting boxes.
func BetweenBoxes(boxes []render.Rect) chan BoxCollision {
	generator := make(chan BoxCollision)

	go func() {
		// Outer loop: test each box for intersection with the others.
		for i, box := range boxes {
			for j := i + 1; j < len(boxes); j++ {
				other := boxes[j]
				collision, err := CompareBoxes(box, other)
				if err == nil {
					collision.A = i
					collision.B = j
					generator <- collision
				}
			}
		}

		close(generator)
	}()

	return generator
}

// CompareBoxes checks if two boxes overlaps and returns information about
// the overlap. The boxes are bounding rectangles like those given to
// BetweenBoxes().
func CompareBoxes(box, other render.Rect) (BoxCollision, error) {
	if box.Intersects(other) {
		var (
			overlap     = OverlapRelative(box, other)
			topLeft     = overlap.TopLeft()
			bottomRight = overlap.BottomRight()
		)
		return BoxCollision{
			Overlap: render.Rect{
				X: topLeft.X,
				Y: topLeft.Y,
				W: bottomRight.X,
				H: bottomRight.Y,
			},
		}, nil
	}
	return BoxCollision{}, errors.New("boxes do not intersect")
}

/*
OverlapRelative returns the Overlap box using coordinates relative
to the source rect instead of absolute coordinates.
*/
func OverlapRelative(source, other render.Rect) CollisionBox {
	var (
		// Move the source rect to 0,0 and record the distance we need
		// to go to get there, so we can move the other rect the same.
		deltaX = 0 - source.X
		deltaY = 0 - source.Y
	)

	source.X = 0
	source.Y = 0
	other.X += deltaX
	other.Y += deltaY

	return Overlap(source, other)
}

/*
Overlap returns the overlap rectangle between two boxes.

The two rects given have an X,Y coordinate and their W,H are their
width and heights.

The returned CollisionBox uses absolute coordinates in the same space
as the passed-in rects.
*/
func Overlap(a, b render.Rect) CollisionBox {
	max := func(x, y int32) int32 {
		return int32(math.Max(float64(x), float64(y)))
	}
	min := func(x, y int32) int32 {
		return int32(math.Min(float64(x), float64(y)))
	}

	var (
		A = GetCollisionBox(a)
		B = GetCollisionBox(b)

		ATL = A.TopLeft()
		ABR = A.BottomRight()
		BTL = B.TopLeft()
		BBR = B.BottomRight()

		// Coordinates of the intersection box.
		X1, Y1 = max(ATL.X, BTL.X), max(ATL.Y, BTL.Y)
		X2, Y2 = min(ABR.X, BBR.X), min(ABR.Y, BBR.Y)
	)

	return NewBox(render.NewPoint(X1, Y1), render.NewPoint(X2, Y2))
}
