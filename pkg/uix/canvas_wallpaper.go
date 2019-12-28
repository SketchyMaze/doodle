package uix

import (
	"git.kirsle.net/apps/doodle/pkg/level"
	"git.kirsle.net/apps/doodle/pkg/wallpaper"
	"git.kirsle.net/go/render"
)

// Wallpaper configures the wallpaper in a Canvas.
type Wallpaper struct {
	pageType  level.PageType
	maxWidth  int64
	maxHeight int64
	corner    render.Texturer
	top       render.Texturer
	left      render.Texturer
	repeat    render.Texturer
}

// Valid returns whether the Wallpaper is configured. Only Levels should
// have wallpapers and Doodads will have nil ones.
func (wp *Wallpaper) Valid() bool {
	return wp.repeat != nil
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
			}
		}
	}

	if !moveBy.IsZero() {
		a.MoveBy(moveBy)
	}
}

// PresentWallpaper draws the wallpaper.
func (w *Canvas) PresentWallpaper(e render.Engine, p render.Point) error {
	var (
		wp       = w.wallpaper
		S        = w.Size()
		size     = wp.corner.Size()
		Viewport = w.ViewportRelative()
		origin   = render.Point{
			X: p.X + w.Scroll.X + w.BoxThickness(1),
			Y: p.Y + w.Scroll.Y + w.BoxThickness(1),
		}
		limit = render.Point{
			// NOTE: we add + the texture size so we would actually draw one
			// full extra texture out-of-bounds for the repeating backgrounds.
			// This is cuz for scrolling we offset the draw spot on a loop.
			X: origin.X + S.W - w.BoxThickness(1) + size.W,
			Y: origin.Y + S.H - w.BoxThickness(1) + size.H,
		}
	)

	// For tiled textures, compute the offset amount. If we are scrolled away
	// from the Origin (0,0) we find out by how far (subtract full tile sizes)
	// and use the remainder as an offset for drawing the tiles.
	var dx, dy int
	if origin.X > p.X {
		for origin.X > p.X && origin.X > size.W {
			origin.X -= size.W
		}
		dx = origin.X
		origin.X = p.X
	}
	if origin.Y > p.Y {
		for origin.Y > p.Y && origin.Y > size.H {
			origin.Y -= size.H
		}
		dy = origin.Y
		origin.Y = p.Y
	}

	// And capping the scroll delta in the other direction.
	if limit.X < S.W {
		limit.X = S.W
	}
	if limit.Y < S.H {
		// TODO: slight flicker on bottom edge when scrolling down
		limit.Y = S.H
	}

	limit.X += size.W
	limit.Y += size.H

	// Tile the repeat texture.
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

			e.Copy(wp.repeat, src, dst)
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
			render.TrimBox(&src, &dst, p, S, w.BoxThickness(1))
			e.Copy(wp.left, src, dst)
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
			render.TrimBox(&src, &dst, p, S, w.BoxThickness(1))
			e.Copy(wp.top, src, dst)
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
			render.TrimBox(&src, &dst, p, S, w.BoxThickness(1))
			e.Copy(wp.corner, src, dst)
		}
	}
	return nil
}

// Load the wallpaper settings from a level.
func (wp *Wallpaper) Load(e render.Engine, pageType level.PageType, v *wallpaper.Wallpaper) error {
	wp.pageType = pageType
	if tex, err := v.CornerTexture(e); err == nil {
		wp.corner = tex
	} else {
		return err
	}

	if tex, err := v.TopTexture(e); err == nil {
		wp.top = tex
	} else {
		return err
	}

	if tex, err := v.LeftTexture(e); err == nil {
		wp.left = tex
	} else {
		return err
	}

	if tex, err := v.RepeatTexture(e); err == nil {
		wp.repeat = tex
	} else {
		return err
	}

	return nil
}
