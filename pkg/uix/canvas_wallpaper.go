package uix

import (
	"git.kirsle.net/SketchyMaze/doodle/pkg/level"
	"git.kirsle.net/SketchyMaze/doodle/pkg/wallpaper"
	"git.kirsle.net/go/render"
)

// Wallpaper configures the wallpaper in a Canvas.
type Wallpaper struct {
	pageType  level.PageType
	maxWidth  int64
	maxHeight int64

	// Pointer to the Wallpaper datum.
	WP *wallpaper.Wallpaper
}

// Valid returns whether the Wallpaper is configured. Only Levels should
// have wallpapers and Doodads will have nil ones.
func (wp *Wallpaper) Valid() bool {
	return wp.WP != nil && wp.WP.Repeat() != nil
}

// Canvas Loop() task that keeps mobile actors constrained inside the borders
// of the world for bounded map types.
func (w *Canvas) loopContainActorsInsideLevel(a *Actor) {
	// Infinite maps do not need to constrain the actors.
	if w.wallpaper.pageType == level.Unbounded {
		return
	}

	var (
		orig   = a.Position() // Actor's World Position
		moveBy render.Point
		size   = a.Size()
	)

	// Bound it on the top left edges.
	if orig.X < 0 {
		moveBy.X = -orig.X
	}
	if orig.Y < 0 {
		moveBy.Y = -orig.Y
	}

	// Bound it on the right bottom edges. XXX: downcast from int64!
	if w.wallpaper.pageType >= level.Bounded {
		if w.wallpaper.maxWidth > 0 {
			if int64(orig.X+size.W) > w.wallpaper.maxWidth {
				var delta = w.wallpaper.maxWidth - int64(orig.X+size.W)
				moveBy.X = int(delta)
			}
		}
		if w.wallpaper.maxHeight > 0 {
			if int64(orig.Y+size.H) > w.wallpaper.maxHeight {
				var delta = w.wallpaper.maxHeight - int64(orig.Y+size.H)
				moveBy.Y = int(delta)

				// Allow them to jump from the bottom by marking them as grounded.
				a.SetGrounded(true)
			}
		}
	}

	if !moveBy.IsZero() {
		a.MoveBy(moveBy)
	}
}

// PresentWallpaper draws the wallpaper.
// Point p is the one given to Canvas.Present(), i.e., the position of the
// top-left corner of the Canvas widget relative to the application window.
func (w *Canvas) PresentWallpaper(e render.Engine, p render.Point) error {
	var (
		wp       = w.wallpaper
		S        = w.Size()
		size     = wp.WP.QuarterRect()
		sizeOrig = wp.WP.QuarterRect()

		// Get the relative viewport of world coordinates looked at by the canvas.
		// The X,Y values are the negative Scroll value
		// The W,H values are the Canvas size same as var S above.
		Viewport = w.ViewportRelative()

		// origin and limit seem to be the boundaries of where on screen
		// we are rendering inside.
		origin = render.Point{
			X: p.X + w.Scroll.X, // + w.BoxThickness(1),
			Y: p.Y + w.Scroll.Y, // + w.BoxThickness(1),
		}
		limit render.Point // TBD later
	)

	// Grow or shrink the render limit if we're zoomed.
	if w.Zoom != 0 {
		// I was surprised to discover that just zooming the texture
		// quadrant size handled most of the problem! For reference, the
		// Blueprint wallpaper has a size of 120x120 for the tiling pattern.
		size.H = w.ZoomMultiply(size.H)
		size.W = w.ZoomMultiply(size.W)
	}

	// SCRATCH
	// at bootup, scroll position 0,0:
	//		origin=44,20    p=44,20     p=relative to application window
	// scroll right and down to -60,-60:
	//		origin=-16,-40   p=44,20    and looks good in that direction
	// scroll left and up to 60,60:
	//		origin=104,80    p=44,20
	//		becomes origin=44,20  p=44,20   d=-16,-40
	// the latter case is handled below. walking thru:
	//    if o(104) > p(44):
	//        while o(104) > p(44):
	//            o -= size(120) of texture block
	//            o is now -16,-40
	//        while o(-16) > p(44): it's not; break
	//    dx = o(-16)
	//    origin.X = p.X
	//    (becomes origin=44,20  p=44,20  d=-16,-40)
	//
	// The visual bug is: if you scroll left or up on an Unbounded level from
	// the origin (0, 0), the tiling of the wallpaper jumps to the right and
	// down by an offset of 44x20 pixels.
	//
	// what is meant to happen:
	// -

	// For tiled textures, compute the offset amount. If we are scrolled away
	// from the Origin (0,0) we find out by how far (subtract full tile sizes)
	// and use the remainder as an offset for drawing the tiles.
	// p      = position on screen of the Canvas widget
	// origin = p.X + Scroll.X, p.Y + scroll.Y
	// note: negative Scroll values means to the right and down
	var dx, dy int
	if origin.X > p.X {
		// View is scrolled leftward (into negative world coordinates)
		dx = origin.X
		for dx > p.X {
			dx -= size.W
		}
		origin.X = 0 // note: origin 0,0 will be the corner of the app window
	}
	if origin.Y > p.Y {
		// View is scrolled upward (into negative world coordinates)
		dy = origin.Y
		for dy > p.Y {
			dy -= size.H
		}
		origin.Y = 0
	}

	limit = render.Point{
		// NOTE: we add + the texture size so we would actually draw one
		// full extra texture out-of-bounds for the repeating backgrounds.
		// This is cuz for scrolling we offset the draw spot on a loop.
		X: origin.X + S.W + size.W,
		Y: origin.Y + S.H + size.H,
	}

	// And capping the scroll delta in the other direction. Always draw
	// pixels until the Canvas size is covered.
	if limit.X < S.W {
		limit.X = S.W
	}
	if limit.Y < S.H {
		// TODO: slight flicker on bottom edge when scrolling down
		limit.Y = S.H
	}

	// TODO: was still getting some slight flicker on the right and bottom
	// when scrolling.. add a bit extra margin.
	limit.X += size.W
	limit.Y += size.H

	// Tile the repeat texture. Start from 1 full wallpaper tile out of bounds
	for x := origin.X - size.W; x < limit.X; x += size.W {
		for y := origin.Y - size.H; y < limit.Y; y += size.H {
			src := render.Rect{
				W: size.W,
				H: size.H,
			}
			dst := render.Rect{
				X: x + dx,
				Y: y + dy,
				W: src.W,
				H: src.H,
			}

			// Trim the edges of the destination box, like in canvas.go#Present
			render.TrimBox(&src, &dst, p, S, w.BoxThickness(1))

			// When zooming OUT, make sure the source rect is at least the
			// full size of the chunk texture; otherwise the ZoomMultiplies
			// above do correctly scale e.g. 128x128 to 64x64, but it only
			// samples the top-left 64x64 then and not the full texture so
			// it more crops it than scales it, but does fit it neatly with
			// its neighbors.
			if w.Zoom < 0 {
				src.W = sizeOrig.W
				src.H = sizeOrig.H
			}

			if tex, err := wp.WP.RepeatTexture(e); err == nil {
				e.Copy(tex, src, dst)
			}
		}
	}

	// The left edge corner tiled along the left edge.
	if wp.pageType > level.Unbounded {
		for y := origin.Y; y < limit.Y; y += size.H {
			src := render.Rect{
				W: size.W,
				H: size.H,
			}
			dst := render.Rect{
				X: origin.X,
				Y: y + dy,
				W: src.W,
				H: src.H,
			}

			// Zoom-out min size constraint.
			if w.Zoom < 0 {
				src.W = sizeOrig.W
				src.H = sizeOrig.H
			}

			render.TrimBox(&src, &dst, p, S, w.BoxThickness(1))
			if tex, err := wp.WP.LeftTexture(e); err == nil {
				e.Copy(tex, src, dst)
			}
		}

		// The top edge tiled along the top edge.
		for x := origin.X; x < limit.X; x += size.W {
			src := render.Rect{
				W: size.W,
				H: size.H,
			}
			dst := render.Rect{
				X: x,
				Y: origin.Y,
				W: src.W,
				H: src.H,
			}

			// Zoom-out min size constraint.
			if w.Zoom < 0 {
				src.W = sizeOrig.W
				src.H = sizeOrig.H
			}

			render.TrimBox(&src, &dst, p, S, w.BoxThickness(1))
			if tex, err := wp.WP.TopTexture(e); err == nil {
				e.Copy(tex, src, dst)
			}
		}

		// The top left corner for all page types except Unbounded.
		if Viewport.Intersects(size) {
			src := render.Rect{
				W: size.W,
				H: size.H,
			}
			dst := render.Rect{
				X: origin.X,
				Y: origin.Y,
				W: src.W,
				H: src.H,
			}

			// Zoom out min size constraint.
			if w.Zoom < 0 {
				src.W = sizeOrig.W
				src.H = sizeOrig.H
			}

			render.TrimBox(&src, &dst, p, S, w.BoxThickness(1))
			if tex, err := wp.WP.CornerTexture(e); err == nil {
				e.Copy(tex, src, dst)
			}
		}
	}
	return nil
}

// Load the wallpaper settings from a level.
func (wp *Wallpaper) Load(pageType level.PageType, v *wallpaper.Wallpaper) error {
	wp.pageType = pageType
	wp.WP = v
	return nil
}
