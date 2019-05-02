package uix

import (
	"errors"
	"fmt"

	"git.kirsle.net/apps/doodle/lib/events"
	"git.kirsle.net/apps/doodle/lib/render"
	"git.kirsle.net/apps/doodle/pkg/balance"
	"git.kirsle.net/apps/doodle/pkg/level"
)

/*
Loop() subroutine to scroll the canvas using arrow keys (for edit mode).

If w.Scrollable is false this function won't do anything.

Cursor keys will scroll the drawing by balance.CanvasScrollSpeed per tick.
If the level pageType is constrained, the scrollable viewport will be
constrained to fit the bounds of the level.

The debug boolean `NoLimitScroll=true` will override the bounded level scroll
restriction and allow scrolling into out-of-bounds areas of the level.
*/
func (w *Canvas) loopEditorScroll(ev *events.State) error {
	if !w.Scrollable {
		return errors.New("canvas not scrollable")
	}

	// Arrow keys to scroll the view.
	scrollBy := render.Point{}
	if ev.Right.Now {
		scrollBy.X -= balance.CanvasScrollSpeed
	} else if ev.Left.Now {
		scrollBy.X += balance.CanvasScrollSpeed
	}
	if ev.Down.Now {
		scrollBy.Y -= balance.CanvasScrollSpeed
	} else if ev.Up.Now {
		scrollBy.Y += balance.CanvasScrollSpeed
	}
	if !scrollBy.IsZero() {
		w.ScrollBy(scrollBy)
	}

	return nil
}

/*
Loop() subroutine to constrain the scrolled view to within a bounded level.
*/
func (w *Canvas) loopConstrainScroll() error {
	if w.NoLimitScroll {
		return errors.New("NoLimitScroll enabled")
	}

	var capped bool

	// Constrain the top and left edges.
	if w.wallpaper.pageType > level.Unbounded {
		if w.Scroll.X > 0 {
			w.Scroll.X = 0
			capped = true
		}
		if w.Scroll.Y > 0 {
			w.Scroll.Y = 0
			capped = true
		}
	}

	// Constrain the bottom and right for limited world sizes.
	if w.wallpaper.maxWidth+w.wallpaper.maxHeight > 0 {
		var (
			// TODO: downcast from int64!
			mw       = int32(w.wallpaper.maxWidth)
			mh       = int32(w.wallpaper.maxHeight)
			Viewport = w.Viewport()
		)
		if Viewport.W > mw {
			delta := Viewport.W - mw
			w.Scroll.X += delta
			capped = true
		}
		if Viewport.H > mh {
			delta := Viewport.H - mh
			w.Scroll.Y += delta
			capped = true
		}
	}

	if capped {
		return errors.New("scroll limited by level constraint")
	}

	return nil
}

/*
Loop() subroutine for Play Mode to follow an actor in the camera's view.

Does nothing if w.FollowActor is an empty string. Set it to the ID of an Actor
to follow. If the actor exists, the Canvas will scroll to keep it on the
screen.
*/
func (w *Canvas) loopFollowActor(ev *events.State) error {
	// Are we following an actor?
	if w.FollowActor == "" {
		return nil
	}

	var (
		VP = w.Viewport()
	)

	// Find the actor.
	for _, actor := range w.actors {
		if actor.ID() != w.FollowActor {
			continue
		}

		var (
			APosition = actor.Position() // absolute world position
			ASize     = actor.Drawing.Size()
			scrollBy  render.Point
		)

		// Scroll left
		if APosition.X-VP.X <= int32(balance.ScrollboxHoz) {
			var delta = APosition.X - VP.X
			if delta > int32(balance.ScrollMaxVelocity) {
				delta = int32(balance.ScrollMaxVelocity)
			}

			if delta < 0 {
				// constrain in case they're FAR OFF SCREEN so we don't flip back around
				delta = -delta
			}
			scrollBy.X = delta
		}

		// Scroll right
		if APosition.X >= VP.W-ASize.W-int32(balance.ScrollboxHoz) {
			var delta = VP.W - ASize.W - int32(balance.ScrollboxHoz)
			if delta > int32(balance.ScrollMaxVelocity) {
				delta = int32(balance.ScrollMaxVelocity)
			}
			scrollBy.X = -delta
		}

		// Scroll up
		if APosition.Y-VP.Y <= int32(balance.ScrollboxVert) {
			var delta = APosition.Y - VP.Y
			if delta > int32(balance.ScrollMaxVelocity) {
				delta = int32(balance.ScrollMaxVelocity)
			}

			// TODO: add gravity to counteract jitters on scrolling vertically
			scrollBy.Y -= int32(balance.Gravity)

			if delta < 0 {
				delta = -delta
			}
			scrollBy.Y = delta
		}

		// Scroll down
		if APosition.Y >= VP.H-ASize.H-int32(balance.ScrollboxVert) {
			var delta = VP.H - ASize.H - int32(balance.ScrollboxVert)
			if delta > int32(balance.ScrollMaxVelocity) {
				delta = int32(balance.ScrollMaxVelocity)
			}
			scrollBy.Y = -delta

			// TODO: add gravity to counteract jitters on scrolling vertically
			scrollBy.Y += int32(balance.Gravity * 3)
		}

		if scrollBy != render.Origin {
			w.ScrollBy(scrollBy)
		}

		return nil
	}

	return fmt.Errorf("actor ID '%s' not found in level", w.FollowActor)
}
