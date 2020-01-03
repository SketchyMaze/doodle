package uix

import (
	"errors"
	"fmt"

	"git.kirsle.net/apps/doodle/pkg/balance"
	"git.kirsle.net/apps/doodle/pkg/level"
	"git.kirsle.net/go/render"
	"git.kirsle.net/go/render/event"
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
func (w *Canvas) loopEditorScroll(ev *event.State) error {
	if !w.Scrollable {
		return errors.New("canvas not scrollable")
	}

	// Arrow keys to scroll the view.
	scrollBy := render.Point{}
	if ev.Right {
		scrollBy.X -= balance.CanvasScrollSpeed
	} else if ev.Left {
		scrollBy.X += balance.CanvasScrollSpeed
	}
	if ev.Down {
		scrollBy.Y -= balance.CanvasScrollSpeed
	} else if ev.Up {
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
	if w.wallpaper.pageType >= level.Bounded &&
		w.wallpaper.maxWidth+w.wallpaper.maxHeight > 0 {
		var (
			// TODO: downcast from int64!
			mw       = int(w.wallpaper.maxWidth)
			mh       = int(w.wallpaper.maxHeight)
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
func (w *Canvas) loopFollowActor(ev *event.State) error {
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
		if APosition.X <= VP.X+balance.ScrollboxHoz {
			var delta = VP.X + balance.ScrollboxHoz - APosition.X

			// constrain in case they're FAR OFF SCREEN so we don't flip back around
			if delta < 0 {
				delta = -delta
			}
			scrollBy.X = delta
		}

		// Scroll right
		if APosition.X >= VP.W-ASize.W-balance.ScrollboxHoz {
			var delta = VP.W - ASize.W - APosition.X - balance.ScrollboxHoz
			scrollBy.X = delta
		}

		// Scroll up
		if APosition.Y <= VP.Y+balance.ScrollboxVert {
			var delta = VP.Y + balance.ScrollboxVert - APosition.Y

			if delta < 0 {
				delta = -delta
			}
			scrollBy.Y = delta
		}

		// Scroll down
		if APosition.Y >= VP.H-ASize.H-balance.ScrollboxVert {
			var delta = VP.H - ASize.H - APosition.Y - balance.ScrollboxVert
			if delta > 300 {
				delta = 300
			} else if delta < -300 {
				delta = -300
			}
			scrollBy.Y = delta
		}

		if scrollBy != render.Origin {
			w.ScrollBy(scrollBy)
		}

		return nil
	}

	return fmt.Errorf("actor ID '%s' not found in level", w.FollowActor)
}
