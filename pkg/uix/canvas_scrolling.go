package uix

import (
	"errors"
	"fmt"

	"git.kirsle.net/apps/doodle/pkg/balance"
	"git.kirsle.net/apps/doodle/pkg/keybind"
	"git.kirsle.net/apps/doodle/pkg/level"
	"git.kirsle.net/apps/doodle/pkg/shmem"
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
	// Shift key to scroll very slowly.
	var (
		scrollBy    = render.Point{}
		scrollSpeed = balance.CanvasScrollSpeed
	)
	if keybind.Shift(ev) {
		scrollSpeed = 1
	}

	// Arrow key handlers.
	if keybind.Right(ev) {
		scrollBy.X -= scrollSpeed
	} else if keybind.Left(ev) {
		scrollBy.X += scrollSpeed
	}
	if keybind.Down(ev) {
		scrollBy.Y -= scrollSpeed
	} else if keybind.Up(ev) {
		scrollBy.Y += scrollSpeed
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
			vw       = w.ZoomDivide(Viewport.W)
			vh       = w.ZoomDivide(Viewport.H)
		)

		if vw > mw {
			delta := vw - mw
			w.Scroll.X += delta
			capped = true
		}
		if vh > mh {
			delta := vh - mh
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
		VP            = w.Viewport()
		engine        = shmem.CurrentRenderEngine
		Width, Height = engine.WindowSize()
		midpoint      = render.NewPoint(Width/2, Height/2)
		scrollboxHoz  = midpoint.X - balance.ScrollboxOffset.X
		scrollboxVert = midpoint.Y - balance.ScrollboxOffset.Y
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
		if APosition.X <= VP.X+scrollboxHoz {
			var delta = VP.X + scrollboxHoz - APosition.X

			// constrain in case they're FAR OFF SCREEN so we don't flip back around
			if delta < 0 {
				delta = -delta
			}
			scrollBy.X = delta
		}

		// Scroll right
		if APosition.X >= VP.W-ASize.W-scrollboxHoz {
			var delta = VP.W - ASize.W - APosition.X - scrollboxHoz
			scrollBy.X = delta
		}

		// Scroll up
		if APosition.Y <= VP.Y+scrollboxVert {
			var delta = VP.Y + scrollboxVert - APosition.Y

			if delta < 0 {
				delta = -delta
			}
			scrollBy.Y = delta
		}

		// Scroll down
		if APosition.Y >= VP.H-ASize.H-scrollboxVert {
			var delta = VP.H - ASize.H - APosition.Y - scrollboxVert
			if delta > 300 {
				delta = 300
			} else if delta < -300 {
				delta = -300
			}
			scrollBy.Y = delta
		}

		// Constrain the maximum scroll speed.
		if scrollBy.X > balance.FollowActorMaxScrollSpeed {
			scrollBy.X = balance.FollowActorMaxScrollSpeed
		} else if scrollBy.X < -balance.FollowActorMaxScrollSpeed {
			scrollBy.X = -balance.FollowActorMaxScrollSpeed
		}
		if scrollBy.Y > balance.FollowActorMaxScrollSpeed {
			scrollBy.Y = balance.FollowActorMaxScrollSpeed
		} else if scrollBy.Y < -balance.FollowActorMaxScrollSpeed {
			scrollBy.Y = -balance.FollowActorMaxScrollSpeed
		}

		if scrollBy != render.Origin {
			w.ScrollBy(scrollBy)
		}

		return nil
	}

	return fmt.Errorf("actor ID '%s' not found in level", w.FollowActor)
}
