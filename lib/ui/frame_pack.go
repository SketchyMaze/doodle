package ui

import (
	"fmt"

	"git.kirsle.net/apps/doodle/lib/render"
)

// Pack provides configuration fields for Frame.Pack().
type Pack struct {
	// Side of the parent to anchor the position to, like N, SE, W. Default
	// is Center.
	Anchor Anchor
	Side   Side

	// If the widget is smaller than its allocated space, grow the widget
	// to fill its space in the Frame.
	Fill  string // "x", "y", "both" or "" for none
	fillX bool
	fillY bool

	Padding int32 // Equal padding on X and Y.
	PadX    int32
	PadY    int32
	Expand  bool // Widget should grow its allocated space to better fill the parent.
}

// Pack a widget along a side of the frame.
func (w *Frame) Pack(child Widget, config ...Pack) {
	var C Pack
	if len(config) > 0 {
		C = config[0]
	}

	// Initialize the pack list for this anchor?
	if _, ok := w.packs[C.Side]; !ok {
		w.packs[C.Side] = []packedWidget{}
	}

	// Padding: if the user only provided Padding add it to both
	// the X and Y value. If the user additionally provided the X
	// and Y value, it will add to the base padding as you'd expect.
	C.PadX += C.Padding
	C.PadY += C.Padding

	// Cache the full X and Y booleans.
	C.fillX = C.Fill == FillX || C.Fill == FillBoth
	C.fillY = C.Fill == FillY || C.Fill == FillBoth

	// Adopt the child widget so it can access the Frame.
	child.Adopt(w)

	w.packs[C.Side] = append(w.packs[C.Side], packedWidget{
		widget: child,
		pack:   C,
	})
	w.widgets = append(w.widgets, child)
}

// computePacked processes all the Pack layout widgets in the Frame.
func (w *Frame) computePacked(e render.Engine) {
	var (
		frameSize = w.BoxSize()

		// maxWidth and maxHeight are always the computed minimum dimensions
		// that the Frame must be to contain all of its children. If the Frame
		// was configured with an explicit Size, the Frame will be that Size,
		// but we still calculate how much space the widgets _actually_ take
		// so we can expand them to fill remaining space in fixed size widgets.
		maxWidth  int32
		maxHeight int32
		visited   = []packedWidget{}
		expanded  = []packedWidget{}
	)

	// Iterate through all packed sides and compute how much space to
	// reserve to contain all of their widgets.
	for side := SideMin; side <= SideMax; side++ {
		if _, ok := w.packs[side]; !ok {
			continue
		}

		var (
			x          int32
			y          int32
			yDirection int32 = 1
			xDirection int32 = 1
		)

		if side == Bottom { // TODO: these need tuning
			y = frameSize.H - w.BoxThickness(4)
			yDirection = -1 * w.BoxThickness(4) // parent + child BoxThickness(1) = 2
		} else if side == Right {
			x = frameSize.W - w.BoxThickness(4)
			xDirection = -1 - w.BoxThickness(4) // - w.BoxThickness(2)
		}

		for _, packedWidget := range w.packs[side] {

			child := packedWidget.widget
			pack := packedWidget.pack
			child.Compute(e)

			if child.Hidden() {
				continue
			}

			x += pack.PadX
			y += pack.PadY

			var (
				// point = child.Point()
				size  = child.Size()
				yStep = y * yDirection
				xStep = x * xDirection
			)

			if xStep+size.W+(pack.PadX*2) > maxWidth {
				maxWidth = xStep + size.W + (pack.PadX * 2)
			}
			if yStep+size.H+(pack.PadY*2) > maxHeight {
				maxHeight = yStep + size.H + (pack.PadY * 2)
			}

			if side == Bottom {
				y -= size.H - pack.PadY
			} else if side == Right {
				x -= size.W - pack.PadX
			}

			child.MoveTo(render.NewPoint(x, y))

			if side == Top {
				y += size.H + pack.PadY
			} else if side == Left {
				x += size.W + pack.PadX
			}

			visited = append(visited, packedWidget)
			if pack.Expand { // TODO: don't fuck with children of fixed size
				expanded = append(expanded, packedWidget)
			}
		}
	}

	// If we have extra space in the Frame and any expanding widgets, let the
	// expanding widgets grow and share the remaining space.
	computedSize := render.NewRect(maxWidth, maxHeight)
	if w.fixedWidth > 0 {
		computedSize.W = w.fixedWidth
	}
	if w.fixedHeight > 0 {
		computedSize.H = w.fixedHeight
	}

	if len(expanded) > 0 && !frameSize.IsZero() && frameSize.Bigger(computedSize) {
		// Divy up the size available.
		growBy := render.Rect{
			W: ((frameSize.W - computedSize.W) / int32(len(expanded))), // - w.BoxThickness(2),
			H: ((frameSize.H - computedSize.H) / int32(len(expanded))), // - w.BoxThickness(2),
		}
		for _, pw := range expanded {
			fmt.Printf("expand %s by %s (comp size %s)\n", pw.widget.ID(), growBy, computedSize)
			pw.widget.ResizeAuto(growBy)
			pw.widget.Compute(e)
		}
	}

	// If we're not using a fixed Frame size, use the dynamically computed one.
	if !w.FixedSize() {
		frameSize = render.NewRect(maxWidth, maxHeight)
	} else {
		// If either of the sizes were left zero, use the dynamically computed one.
		if frameSize.W == 0 {
			frameSize.W = maxWidth
		}
		if frameSize.H == 0 {
			frameSize.H = maxHeight
		}
	}

	// Rescan all the widgets in this side to re-center them
	// in their space.
	innerFrameSize := render.NewRect(
		frameSize.W-w.BoxThickness(2),
		frameSize.H-w.BoxThickness(2),
	)
	for _, pw := range visited {
		var (
			child   = pw.widget
			pack    = pw.pack
			point   = child.Point()
			size    = child.Size()
			resize  = size
			resized bool
			moved   bool
		)

		if pack.Side == Top || pack.Side == Bottom {
			if pack.fillX && resize.W < innerFrameSize.W {
				resize.W = innerFrameSize.W - w.BoxThickness(2)
				resized = true
			}
			if resize.W < innerFrameSize.W-w.BoxThickness(4) {
				if pack.Anchor.IsCenter() {
					point.X = (innerFrameSize.W / 2) - (resize.W / 2)
				} else if pack.Anchor.IsWest() {
					point.X = pack.PadX
				} else if pack.Anchor.IsEast() {
					point.X = innerFrameSize.W - resize.W - pack.PadX
				}

				moved = true
			}
		} else if pack.Side == Left || pack.Side == Right {
			if pack.fillY && resize.H < innerFrameSize.H {
				resize.H = innerFrameSize.H - w.BoxThickness(2) // BoxThickness(2) for parent + child
				// point.Y -= (w.BoxThickness(4) + child.BoxThickness(2))
				moved = true
				resized = true
			}

			// Vertically align the widgets.
			if resize.H < innerFrameSize.H {
				if pack.Anchor.IsMiddle() {
					point.Y = (innerFrameSize.H / 2) - (resize.H / 2) - w.BoxThickness(1)
				} else if pack.Anchor.IsNorth() {
					point.Y = pack.PadY - w.BoxThickness(4)
				} else if pack.Anchor.IsSouth() {
					point.Y = innerFrameSize.H - resize.H - pack.PadY
				}
				moved = true
			}
		} else {
			panic("unsupported pack.Side")
		}

		if resized && size != resize {
			fmt.Printf("fill: resize %s to %s\n", child.ID(), resize)
			child.ResizeAuto(resize)
			child.Compute(e)
		}
		if moved {
			child.MoveTo(point)
		}
	}

	// if !w.FixedSize() {
	w.ResizeAuto(render.NewRect(
		frameSize.W-w.BoxThickness(2),
		frameSize.H-w.BoxThickness(2),
	))
	// }
}

// Anchor is a cardinal direction.
type Anchor uint8

// Side of a parent widget to pack children against.
type Side uint8

// Anchor and Side constants.
const (
	Center Anchor = iota
	N
	NE
	E
	SE
	S
	SW
	W
	NW

	Top Side = iota
	Left
	Right
	Bottom
)

// Range of Anchor and Side values.
const (
	AnchorMin = Center
	AnchorMax = NW

	SideMin = Top
	SideMax = Bottom
)

// IsNorth returns if the anchor is N, NE or NW.
func (a Anchor) IsNorth() bool {
	return a == N || a == NE || a == NW
}

// IsSouth returns if the anchor is S, SE or SW.
func (a Anchor) IsSouth() bool {
	return a == S || a == SE || a == SW
}

// IsEast returns if the anchor is E, NE or SE.
func (a Anchor) IsEast() bool {
	return a == E || a == NE || a == SE
}

// IsWest returns if the anchor is W, NW or SW.
func (a Anchor) IsWest() bool {
	return a == W || a == NW || a == SW
}

// IsCenter returns if the anchor is Center, N or S, to determine
// whether to align text as centered for North/South anchors.
func (a Anchor) IsCenter() bool {
	return a == Center || a == N || a == S
}

// IsMiddle returns if the anchor is Center, E or W, to determine
// whether to align text as middled for East/West anchors.
func (a Anchor) IsMiddle() bool {
	return a == Center || a == W || a == E
}

type packLayout struct {
	widgets []packedWidget
}

type packedWidget struct {
	widget Widget
	pack   Pack
	fill   uint8
}

// packedWidget.fill values
const (
	fillNone uint8 = iota
	fillX
	fillY
	fillBoth
)
