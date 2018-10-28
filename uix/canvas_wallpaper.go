package uix

import (
	"git.kirsle.net/apps/doodle/level"
	"git.kirsle.net/apps/doodle/pkg/wallpaper"
	"git.kirsle.net/apps/doodle/render"
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
	var dx, dy int32
	if origin.X > 0 {
		for origin.X > 0 && origin.X > size.W {
			origin.X -= size.W
		}
		dx = origin.X
		origin.X = 0
	}
	if origin.Y > 0 {
		for origin.Y > 0 && origin.Y > size.H {
			origin.Y -= size.H
		}
		dy = origin.Y
		origin.Y = 0
	}

	// And capping the scroll delta in the other direction.
	if limit.X < S.W {
		limit.X = S.W
	}
	if limit.Y < S.H {
		// TODO: slight flicker on bottom edge when scrolling down
		limit.Y = S.H
	}

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
