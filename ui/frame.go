package ui

import (
	"fmt"

	"git.kirsle.net/apps/doodle/render"
)

// Frame is a widget that contains other widgets.
type Frame struct {
	Name string
	BaseWidget
	packs   map[Anchor][]packedWidget
	widgets []Widget
}

// NewFrame creates a new Frame.
func NewFrame(name string) *Frame {
	w := &Frame{
		Name:    name,
		packs:   map[Anchor][]packedWidget{},
		widgets: []Widget{},
	}
	w.IDFunc(func() string {
		return fmt.Sprintf("Frame<%s; %d widgets>",
			name,
			len(w.widgets),
		)
	})
	return w
}

// Compute the size of the Frame.
func (w *Frame) Compute(e render.Engine) {
	var (
		frameSize = w.Size()

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

	// Iterate through all anchored directions and compute how much space to
	// reserve to contain all of their widgets.
	for anchor := AnchorMin; anchor <= AnchorMax; anchor++ {
		if _, ok := w.packs[anchor]; !ok {
			continue
		}

		var (
			x          int32
			y          int32
			yDirection int32 = 1
			xDirection int32 = 1
		)

		if anchor.IsSouth() {
			y = frameSize.H
			yDirection = -1 - w.BoxThickness(2) // parent + child BoxThickness(1) = 2
		} else if anchor == E {
			x = frameSize.W
			xDirection = -1 - w.BoxThickness(2)
		}

		for _, packedWidget := range w.packs[anchor] {
			child := packedWidget.widget
			pack := packedWidget.pack
			child.Compute(e)
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

			if anchor.IsSouth() {
				y -= size.H + (pack.PadY * 2)
			}
			if anchor.IsEast() {
				x -= size.W + (pack.PadX * 2)
			}

			child.MoveTo(render.Point{
				X: x + pack.PadX,
				Y: y + pack.PadY,
			})

			if anchor.IsNorth() {
				y += size.H + (pack.PadY * 2)
			}
			if anchor == W {
				x += size.W + (pack.PadX * 2)
			}

			visited = append(visited, packedWidget)
			if pack.Expand {
				expanded = append(expanded, packedWidget)
			}
		}
	}

	// If we have extra space in the Frame and any expanding widgets, let the
	// expanding widgets grow and share the remaining space.
	computedSize := render.NewRect(maxWidth, maxHeight)
	if len(expanded) > 0 && !frameSize.IsZero() && frameSize.Bigger(computedSize) {
		// Divy up the size available.
		growBy := render.Rect{
			W: ((frameSize.W - computedSize.W) / int32(len(expanded))) - w.BoxThickness(2),
			H: ((frameSize.H - computedSize.H) / int32(len(expanded))) - w.BoxThickness(2),
		}
		for _, pw := range expanded {
			pw.widget.ResizeBy(growBy)
			pw.widget.Compute(e)
		}
	}

	// If we're not using a fixed Frame size, use the dynamically computed one.
	if !w.FixedSize() {
		frameSize = render.NewRect(maxWidth, maxHeight)
	}

	// Rescan all the widgets in this anchor to re-center them
	// in their space.
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

		if pack.Anchor.IsNorth() || pack.Anchor.IsSouth() {
			if pack.FillX && resize.W < frameSize.W {
				resize.W = frameSize.W - w.BoxThickness(2)
				resized = true
			}
			if resize.W < frameSize.W-w.BoxThickness(4) {
				if pack.Anchor.IsCenter() {
					point.X = (frameSize.W / 2) - (resize.W / 2)
				} else if pack.Anchor.IsWest() {
					point.X = pack.PadX
				} else if pack.Anchor.IsEast() {
					point.X = frameSize.W - resize.W - pack.PadX
				}

				moved = true
			}
		} else if pack.Anchor.IsWest() || pack.Anchor.IsEast() {
			if pack.FillY && resize.H < frameSize.H {
				resize.H = frameSize.H - w.BoxThickness(2) // BoxThickness(2) for parent + child
				// point.Y -= (w.BoxThickness(4) + child.BoxThickness(2))
				moved = true
				resized = true
			}

			// Vertically align the widgets.
			if resize.H < frameSize.H {
				if pack.Anchor.IsMiddle() {
					point.Y = (frameSize.H / 2) - (resize.H / 2)
				} else if pack.Anchor.IsNorth() {
					point.Y = pack.PadY - w.BoxThickness(4)
				} else if pack.Anchor.IsSouth() {
					point.Y = frameSize.H - resize.H - pack.PadY
				}
				moved = true
			}
		} else {
			log.Error("unsupported pack.Anchor")
		}

		if resized && size != resize {
			child.Resize(resize)
			child.Compute(e)
		}
		if moved {
			child.MoveTo(point)
		}
	}

	if !w.FixedSize() {
		w.Resize(frameSize)
	}
}

// Present the Frame.
func (w *Frame) Present(e render.Engine) {
	var (
		P = w.Point()
		S = w.Size()
	)

	// Draw the widget's border and everything.
	w.DrawBox(e)

	// Draw the background color.
	e.DrawBox(w.Background(), render.Rect{
		X: P.X + w.BoxThickness(1),
		Y: P.Y + w.BoxThickness(1),
		W: S.W - w.BoxThickness(2),
		H: S.H - w.BoxThickness(2),
	})

	// Draw the widgets.
	for _, child := range w.widgets {
		p := child.Point()
		child.MoveTo(render.NewPoint(
			P.X+p.X+w.BoxThickness(1),
			P.Y+p.Y+w.BoxThickness(1),
		))
		child.Present(e)
	}
}

// Pack provides configuration fields for Frame.Pack().
type Pack struct {
	// Side of the parent to anchor the position to, like N, SE, W. Default
	// is Center.
	Anchor Anchor

	// If the widget is smaller than its allocated space, grow the widget
	// to fill its space in the Frame.
	Fill  bool
	FillX bool
	FillY bool

	Padding int32 // Equal padding on X and Y.
	PadX    int32
	PadY    int32
	Expand  bool // Widget should grow its allocated space to better fill the parent.
}

// Anchor is a cardinal direction.
type Anchor uint8

// Anchor values.
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
)

// Range of Anchor values.
const (
	AnchorMin = Center
	AnchorMax = NW
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

// Pack a widget along a side of the frame.
func (w *Frame) Pack(child Widget, config ...Pack) {
	var C Pack
	if len(config) > 0 {
		C = config[0]
	}

	// Initialize the pack list for this anchor?
	if _, ok := w.packs[C.Anchor]; !ok {
		w.packs[C.Anchor] = []packedWidget{}
	}

	// Padding: if the user only provided Padding add it to both
	// the X and Y value. If the user additionally provided the X
	// and Y value, it will add to the base padding as you'd expect.
	C.PadX += C.Padding
	C.PadY += C.Padding

	// Fill: true implies both directions.
	if C.Fill {
		C.FillX = true
		C.FillY = true
	}

	w.packs[C.Anchor] = append(w.packs[C.Anchor], packedWidget{
		widget: child,
		pack:   C,
	})
	w.widgets = append(w.widgets, child)
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
