package uix

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"git.kirsle.net/SketchyMaze/doodle/pkg/level"
	"git.kirsle.net/SketchyMaze/doodle/pkg/log"
	"git.kirsle.net/SketchyMaze/doodle/pkg/plus/dpp"
	"git.kirsle.net/SketchyMaze/doodle/pkg/scripting"
	"git.kirsle.net/SketchyMaze/doodle/pkg/scripting/exceptions"
	"git.kirsle.net/go/render"
)

// InstallActors adds external Actors to the canvas to be superimposed on top
// of the drawing.
func (w *Canvas) InstallActors(actors level.ActorMap) error {
	var errs []string

	// Order the actors deterministically, by their ID string. Actors get
	// a time-ordered UUID ID by default so the most recently added actor
	// should render on top of the others.
	var actorIDs []string
	for id := range actors {
		actorIDs = append(actorIDs, id)
	}
	sort.Strings(actorIDs)

	// In case we are replacing the actors, free up all their textures first!
	for _, actor := range w.actors {
		actor.Canvas.Destroy()
	}

	// Signed Levels: the free version normally won't load embedded assets from
	// a level and the call to LoadFromEmbeddable below returns the error. If the
	// level is signed it is allowed to use its embedded assets.
	isSigned := w.IsSignedLevelPack != nil || dpp.Driver.IsLevelSigned(w.level)

	w.actors = make([]*Actor, 0)
	for _, id := range actorIDs {
		var actor = actors[id]

		// Try loading the doodad from the level's own attached files.
		doodad, err := dpp.Driver.LoadFromEmbeddable(actor.Filename, w.level, isSigned)
		if err != nil {
			// If we have a signed levelpack, try loading from the levelpack.
			if w.IsSignedLevelPack != nil {
				if found, err := dpp.Driver.LoadFromEmbeddable(actor.Filename, w.IsSignedLevelPack, true); err == nil {
					doodad = found
				}
			}

			// If not found, append the error and continue.
			if doodad == nil {
				errs = append(errs, fmt.Sprintf("%s: %s", actor.Filename, err.Error()))
				continue
			}
		}

		// Create the "live" Actor to exist in the world, and set its world
		// position to the Point defined in the level data.
		liveActor := NewActor(id, actor, doodad)
		liveActor.Canvas.parent = w
		liveActor.LevelCanvas = w
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

		if vm.Self != nil {
			// Already initialized!
			continue
		}

		// Security: expose a selective API to the actor to the JS engine.
		vm.Self = w.MakeSelfAPI(actor)
		w.MakeScriptAPI(vm)
		vm.Set("Self", vm.Self)

		// If there is no script attached, do not try and load or call the main() function.
		if actor.Doodad().Script == "" {
			continue
		}

		if _, err := vm.Run(actor.Doodad().Script); err != nil {
			log.Error("Run script for actor %s failed: %s", actor.ID(), err)
		}

		// Call the main() function.
		if err := vm.Main(); err != nil {
			exceptions.FormatAndCatch(
				nil,
				"Error in main() for actor %s:\n\n%s\n\nActor ID: %s\nFilename: %s\nPosition: %s",
				actor.Actor.Filename,
				err,
				actor.ID(),
				actor.Actor.Filename,
				actor.Position(),
			)
		}
	}

	// Broadcast the "ready" signal to any actors that want to publish
	// messages ASAP on level start.
	for _, actor := range w.actors {
		w.scripting.To(actor.ID()).Inbound <- scripting.Message{
			Name: "broadcast:ready",
			Args: nil,
		}
	}

	return nil
}

// AddActor injects additional actors into the canvas, such as a Player doodad.
func (w *Canvas) AddActor(actor *Actor) error {
	actor.LevelCanvas = w
	w.actors = append(w.actors, actor)
	return nil
}

// RemoveActor removes the actor from the canvas.
func (w *Canvas) RemoveActor(actor *Actor) {
	var actors = []*Actor{}
	for _, exist := range w.actors {
		if actor == exist {
			w.scripting.RemoveVM(actor.ID())
			continue
		}
		actors = append(actors, exist)
	}
	w.actors = actors
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
			resizeTo   = actorSize
		)

		// Adjust actor position and size by the zoom level.
		actorPoint.X = w.ZoomMultiply(actorPoint.X)
		actorPoint.Y = w.ZoomMultiply(actorPoint.Y)
		resizeTo.W = w.ZoomMultiply(resizeTo.W)
		resizeTo.H = w.ZoomMultiply(resizeTo.H)

		// Tell the actor's canvas to copy our zoom level so it resizes its image too.
		can.Zoom = w.Zoom

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
			scrollTo.X -= delta
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
			scrollTo.Y -= delta
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
