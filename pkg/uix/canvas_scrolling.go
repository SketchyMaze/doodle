package uix

import (
	"errors"
	"fmt"

	"git.kirsle.net/SketchyMaze/doodle/pkg/balance"
	"git.kirsle.net/SketchyMaze/doodle/pkg/drawtool"
	"git.kirsle.net/SketchyMaze/doodle/pkg/keybind"
	"git.kirsle.net/SketchyMaze/doodle/pkg/level"
	"git.kirsle.net/SketchyMaze/doodle/pkg/shmem"
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

	// Multitouch events to pan the level, like middle click on desktop.
	if ev.Touching && ev.TouchNumFingers > 1 {
		// Intention: user drags with 2 fingers to scroll the canvas.
		// SDL2 will register one finger also as a Button1 mouse click.
		// We need to record the "mouse cursor" as start point but then
		// fake that no click occurs so we don't nick the drawing.
		if !w.scrollDragging {
			w.scrollDragging = true
			w.scrollStartAt = shmem.Cursor
			w.scrollWasAt = w.Scroll
			w.scrollLastDelta = render.Point{}
		} else {
			delta := shmem.Cursor.Compare(w.scrollStartAt)
			w.Scroll = w.scrollWasAt
			w.Scroll.Subtract(delta)

			// So, SDL2 spams us with events for every subtle movement of 2+ fingers
			// on the screen, but we don't know when that STOPS. As a heuristic, it
			// seems we can tell by if the delta stops updating.
			if !w.scrollLastDelta.IsZero() {
				if w.scrollLastDelta == delta {
					ev.Touching = false
				}
			}

			w.scrollLastDelta = delta
		}

		// Lift the mouse button.
		ev.Button1 = false

		return nil
	}

	// Middle click of the mouse to pan the level.
	// NOTE: PanTool intercepts both Left and MiddleClick.
	if w.Tool != drawtool.PanTool {
		if keybind.MiddleClick(ev) {
			if !w.scrollDragging {
				w.scrollDragging = true
				w.scrollStartAt = shmem.Cursor
				w.scrollWasAt = w.Scroll
			} else {
				delta := shmem.Cursor.Compare(w.scrollStartAt)
				w.Scroll = w.scrollWasAt
				w.Scroll.Subtract(delta)
			}
		} else {
			if w.scrollDragging {
				w.scrollDragging = false
			}
		}
	}

	return nil
}

/*
Loop() subroutine to constrain the scrolled view to within a bounded level.
*/
func (w *Canvas) loopConstrainScroll() error {
	if w.NoLimitScroll || w.scrollOutOfBounds {
		return errors.New("NoLimitScroll enabled")
	}

	// Levels only.
	if w.level == nil {
		return nil
	}

	var (
		capped    bool
		maxWidth  = w.level.MaxWidth
		maxHeight = w.level.MaxHeight
	)

	// Constrain the bottom and right for limited world sizes.
	if w.wallpaper.pageType >= level.Bounded &&
		maxWidth+maxHeight > 0 {
		var (
			// TODO: downcast from int64!
			mw       = int(maxWidth)
			mh       = int(maxHeight)
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

		// If we are VERY FAR away, allow greater leaps.
		if scrollBy.X > balance.FollowActorMaxScrollSpeed*4 {
			scrollBy.X = balance.FollowActorMaxScrollSpeed * 4
		} else if scrollBy.X < -balance.FollowActorMaxScrollSpeed*4 {
			scrollBy.X = -balance.FollowActorMaxScrollSpeed * 4
		}
		if scrollBy.Y > balance.FollowActorMaxScrollSpeed*4 {
			scrollBy.Y = balance.FollowActorMaxScrollSpeed * 4
		} else if scrollBy.Y < -balance.FollowActorMaxScrollSpeed*4 {
			scrollBy.Y = -balance.FollowActorMaxScrollSpeed * 4
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
