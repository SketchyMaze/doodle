package collision

import (
	"git.kirsle.net/apps/doodle/lib/render"
)

// IndexTuple holds two integers used as array indexes.
type IndexTuple [2]int

// BetweenBoxes checks if there is a collision between any
// two bounding rectangles.
//
// This returns a generator that spits out indexes of the
// intersecting boxes.
func BetweenBoxes(boxes []render.Rect) chan IndexTuple {
	generator := make(chan IndexTuple)

	go func() {
		// Outer loop: test each box for intersection with the others.
		for i, box := range boxes {
			for j := i + 1; j < len(boxes); j++ {
				if box.Intersects(boxes[j]) {
					generator <- IndexTuple{i, j}
				}
			}
		}

		close(generator)
	}()

	return generator
}
