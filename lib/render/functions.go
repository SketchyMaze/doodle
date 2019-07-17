package render

import (
	"fmt"
	"regexp"
	"strconv"
)

var regexpResolution = regexp.MustCompile(`^(\d+)x(\d+)$`)

// ParseResolution turns a resolution string like "1024x768" and returns the
// width and height values.
func ParseResolution(resi string) (int, int, error) {
	m := regexpResolution.FindStringSubmatch(resi)
	if m == nil {
		return 0, 0, fmt.Errorf("invalid resolution format, should be %s",
			regexpResolution.String(),
		)
	}

	width, err := strconv.Atoi(m[1])
	if err != nil {
		return 0, 0, err
	}

	height, err := strconv.Atoi(m[2])
	if err != nil {
		return 0, 0, err
	}

	return width, height, nil
}

// TrimBox helps with Engine.Copy() to trim a destination box so that it
// won't overflow with the parent container.
func TrimBox(src, dst *Rect, p Point, S Rect, thickness int32) {
	// Constrain source width to not bigger than Canvas width.
	if src.W > S.W {
		src.W = S.W
	}
	if src.H > S.H {
		src.H = S.H
	}

	// If the destination width will cause it to overflow the widget
	// box, trim off the right edge of the destination rect.
	//
	// Keep in mind we're dealing with chunks here, and a chunk is
	// a small part of the image. Example:
	// - Canvas is 800x600 (S.W=800  S.H=600)
	// - Chunk wants to render at 790,0 width 100,100 or whatever
	//   dst={790, 0, 100, 100}
	// - Chunk box would exceed 800px width (X=790 + W=100 == 890)
	// - Find the delta how much it exceeds as negative (800 - 890 == -90)
	// - Lower the Source and Dest rects by that delta size so they
	//   stay proportional and don't scale or anything dumb.
	if dst.X+src.W > p.X+S.W {
		// NOTE: delta is a negative number,
		// so it will subtract from the width.
		delta := (p.X + S.W - thickness) - (dst.W + dst.X)
		src.W += delta
		dst.W += delta
	}
	if dst.Y+src.H > p.Y+S.H {
		// NOTE: delta is a negative number
		delta := (p.Y + S.H - thickness) - (dst.H + dst.Y)
		src.H += delta
		dst.H += delta
	}

	// The same for the top left edge, so the drawings don't overlap
	// menu bars or left side toolbars.
	// - Canvas was placed 80px from the left of the screen.
	//   Canvas.MoveTo(80, 0)
	// - A texture wants to draw at 60, 0 which would cause it to
	//   overlap 20 pixels into the left toolbar. It needs to be cropped.
	// - The delta is: p.X=80 - dst.X=60 == 20
	// - Set destination X to p.X to constrain it there: 20
	// - Subtract the delta from destination W so we don't scale it.
	// - Add 20 to X of the source: the left edge of source is not visible
	if dst.X < p.X {
		// NOTE: delta is a positive number,
		// so it will add to the destination coordinates.
		delta := p.X - dst.X
		dst.X = p.X + thickness
		dst.W -= delta
		src.X += delta
	}
	if dst.Y < p.Y {
		delta := p.Y - dst.Y
		dst.Y = p.Y + thickness
		dst.H -= delta
		src.Y += delta
	}

	// Trim the destination width so it doesn't overlap the Canvas border.
	if dst.W >= S.W-thickness {
		dst.W = S.W - thickness
	}
}

// AbsInt32 returns the absolute value of an int32.
func AbsInt32(v int32) int32 {
	if v < 0 {
		return -v
	}
	return v
}
