package uix

import (
	"errors"
	"fmt"
	"strings"

	"git.kirsle.net/apps/doodle/pkg/doodads"
	"git.kirsle.net/apps/doodle/pkg/level"
	"git.kirsle.net/apps/doodle/pkg/log"
	"git.kirsle.net/apps/doodle/pkg/scripting"
	"git.kirsle.net/go/render"
)

// InstallActors adds external Actors to the canvas to be superimposed on top
// of the drawing.
func (w *Canvas) InstallActors(actors level.ActorMap) error {
	var errs []string

	w.actors = make([]*Actor, 0)
	for id, actor := range actors {
		doodad, err := doodads.LoadFromEmbeddable(actor.Filename, w.level)
		if err != nil {
			errs = append(errs, fmt.Sprintf("%s: %s", actor.Filename, err.Error()))
			continue
		}

		// Create the "live" Actor to exist in the world, and set its world
		// position to the Point defined in the level data.
		liveActor := NewActor(id, actor, doodad)
		liveActor.MoveTo(actor.Point)

		w.actors = append(w.actors, liveActor)
	}

	if len(errs) > 0 {
		return errors.New(strings.Join(errs, "\n"))
	}
	return nil
}

// Actors returns the list of actors currently in the Canvas.
func (w *Canvas) Actors() []*Actor {
	return w.actors
}

// ClearActors removes all the actors from the Canvas.
func (w *Canvas) ClearActors() {
	w.actors = []*Actor{}
}

// SetScriptSupervisor assigns the Canvas scripting supervisor to enable
// interaction with actor scripts.
func (w *Canvas) SetScriptSupervisor(s *scripting.Supervisor) {
	w.scripting = s
}

// InstallScripts loads all the current actors' scripts into the scripting
// engine supervisor.
func (w *Canvas) InstallScripts() error {
	if w.scripting == nil {
		return errors.New("no script supervisor is configured for this canvas")
	}

	if len(w.actors) == 0 {
		return errors.New("no actors exist in this canvas to install scripts for")
	}

	for _, actor := range w.actors {
		vm := w.scripting.To(actor.ID())

		// Security: expose a selective API to the actor to the JS engine.
		vm.Self = w.MakeSelfAPI(actor)
		vm.Set("Self", vm.Self)

		if _, err := vm.Run(actor.Doodad().Script); err != nil {
			log.Error("Run script for actor %s failed: %s", actor.ID(), err)
		}

		// Call the main() function.
		log.Debug("Calling Main() for %s", actor.ID())
		if err := vm.Main(); err != nil {
			log.Error("main() for actor %s errored: %s", actor.ID(), err)
		}
	}

	return nil
}

// AddActor injects additional actors into the canvas, such as a Player doodad.
func (w *Canvas) AddActor(actor *Actor) error {
	w.actors = append(w.actors, actor)
	return nil
}

// drawActors is a subroutine of Present() that superimposes the actors on top
// of the level drawing.
func (w *Canvas) drawActors(e render.Engine, p render.Point) {
	var (
		Viewport = w.ViewportRelative()
		S        = w.Size()
	)

	// See if each Actor is in range of the Viewport.
	for i, a := range w.actors {
		if a == nil {
			log.Error("Canvas.drawActors: null actor at index %d (of %d actors)", i, len(w.actors))
			continue
		}

		// Skip hidden actors.
		if a.hidden {
			continue
		}

		var (
			can        = a.Canvas // Canvas widget that draws the actor
			actorPoint = a.Position()
			actorSize  = a.Size()
		)

		// Create a box of World Coordinates that this actor occupies. The
		// Actor X,Y from level data is already a World Coordinate;
		// accomodate for the size of the Actor.
		actorBox := render.Rect{
			X: actorPoint.X,
			Y: actorPoint.Y,
			W: actorSize.W,
			H: actorSize.H,
		}

		// Is any part of the actor visible?
		if !Viewport.Intersects(actorBox) {
			continue // not visible on screen
		}

		drawAt := render.Point{
			X: p.X + w.Scroll.X + actorPoint.X + w.BoxThickness(1),
			Y: p.Y + w.Scroll.Y + actorPoint.Y + w.BoxThickness(1),
		}
		resizeTo := actorSize

		// XXX TODO: when an Actor hits the left or top edge and shrinks,
		// scrolling to offset that shrink is currently hard to solve.
		scrollTo := render.Origin

		// Handle cropping and scaling if this Actor's canvas can't be
		// completely visible within the parent.
		if drawAt.X+resizeTo.W > p.X+S.W {
			// Hitting the right edge, shrunk the width now.
			delta := (drawAt.X + resizeTo.W) - (p.X + S.W)
			resizeTo.W -= delta
		} else if drawAt.X < p.X {
			// Hitting the left edge. Cap the X coord and shrink the width.
			delta := p.X - drawAt.X // positive number
			drawAt.X = p.X
			// scrollTo.X -= delta / 2 // TODO
			resizeTo.W -= delta
		}

		if drawAt.Y+resizeTo.H > p.Y+S.H {
			// Hitting the bottom edge, shrink the height.
			delta := (drawAt.Y + resizeTo.H) - (p.Y + S.H)
			resizeTo.H -= delta
		} else if drawAt.Y < p.Y {
			// Hitting the top edge. Cap the Y coord and shrink the height.
			delta := p.Y - drawAt.Y
			drawAt.Y = p.Y
			// scrollTo.Y -= delta // TODO
			resizeTo.H -= delta
		}

		if resizeTo != actorSize {
			can.Resize(resizeTo)
			can.ScrollTo(scrollTo)
		}
		can.Present(e, drawAt)

		// Clean up the canvas size and offset.
		can.Resize(actorSize) // restore original size in case cropped
		can.ScrollTo(render.Origin)
	}
}
