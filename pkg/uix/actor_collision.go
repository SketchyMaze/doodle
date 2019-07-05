package uix

import (
	"errors"
	"sync"
	"time"

	"git.kirsle.net/apps/doodle/lib/render"
	"git.kirsle.net/apps/doodle/pkg/balance"
	"git.kirsle.net/apps/doodle/pkg/collision"
	"git.kirsle.net/apps/doodle/pkg/doodads"
	"git.kirsle.net/apps/doodle/pkg/log"
	"git.kirsle.net/apps/doodle/pkg/scripting"
	"github.com/robertkrimen/otto"
)

// loopActorCollision is the Loop function that checks if pairs of
// actors are colliding with each other, and handles their scripting
// responses to such collisions.
func (w *Canvas) loopActorCollision() error {
	if w.scripting == nil {
		return errors.New("Canvas.loopActorCollision: scripting engine not attached to Canvas")
	}

	var (
		// Current time of this tick so we can advance animations.
		now = time.Now()

		// As we iterate over all actors below to process their movement, track
		// their bounding rectangles so we can later see if any pair of actors
		// intersect each other. Also, in case of actor scripts protesting a
		// collision later, store each actor's original position before the move.
		boxes             = make([]render.Rect, len(w.actors))
		originalPositions = map[string]render.Point{}
	)

	// Loop over all the actors in parallel, processing their movement and
	// checking collision data against the level geometry.
	var wg sync.WaitGroup
	for i, a := range w.actors {
		wg.Add(1)
		go func(i int, a *Actor) {
			defer wg.Done()
			originalPositions[a.ID()] = a.Position()

			// Advance any animations for this actor.
			if a.activeAnimation != nil && a.activeAnimation.nextFrameAt.Before(now) {
				if done := a.TickAnimation(a.activeAnimation); done {
					// Animation has finished, run the callback script.
					if a.animationCallback.IsFunction() {
						a.animationCallback.Call(otto.NullValue())
					}

					// Clean up the animation state.
					a.StopAnimation()
				}
			}

			// Get the actor's velocity to see if it's moving this tick.
			v := a.Velocity()
			if a.hasGravity {
				v.Y += int32(balance.Gravity)
			}

			// If not moving, grab the bounding box right now.
			if v == render.Origin {
				boxes[i] = doodads.GetBoundingRect(a)
				return
			}

			// Create a delta point from their current location to where they
			// want to move to this tick.
			delta := a.Position()
			delta.Add(v)

			// Check collision with level geometry.
			info, ok := collision.CollidesWithGrid(a, w.chunks, delta)
			if ok {
				// Collision happened with world.
			}
			delta = info.MoveTo // Move us back where the collision check put us

			// Move the actor's World Position to the new location.
			a.MoveTo(delta)

			// Keep the actor from leaving the world borders of bounded maps.
			w.loopContainActorsInsideLevel(a)

			// Store this actor's bounding box after they've moved.
			boxes[i] = doodads.GetBoundingRect(a)
		}(i, a)
		wg.Wait()
	}

	var collidingActors = map[string]string{}
	for tuple := range collision.BetweenBoxes(boxes) {
		a, b := w.actors[tuple.A], w.actors[tuple.B]
		collidingActors[a.ID()] = b.ID()

		// Call the OnCollide handler.
		if w.scripting != nil {
			// Tell actor A about the collision with B.
			if err := w.scripting.To(a.ID()).Events.RunCollide(&CollideEvent{
				Actor:    b,
				Overlap:  tuple.Overlap,
				InHitbox: tuple.Overlap.Intersects(a.Hitbox()),
			}); err != nil {
				if err == scripting.ErrReturnFalse {
					if origPoint, ok := originalPositions[b.ID()]; ok {
						// Trace a vector back from the actor's current position
						// to where they originated from and find the earliest
						// point where they are not violating the hitbox.
						var (
							rect   = doodads.GetBoundingRect(b)
							hitbox = a.Hitbox()
						)
						for point := range render.IterLine2(
							b.Position(),
							origPoint,
						) {
							test := render.Rect{
								X: point.X,
								Y: point.Y,
								W: rect.W,
								H: rect.H,
							}
							info, err := collision.CompareBoxes(
								boxes[tuple.A],
								test,
							)
							if err != nil || !info.Overlap.Intersects(hitbox) {
								b.MoveTo(point)
								break
							}
						}
					} else {
						log.Error(
							"ERROR: Actors %s and %s overlap and the script returned false,"+
								"but I didn't store %s original position earlier??",
							a.Doodad.Title, b.Doodad.Title, b.Doodad.Title,
						)
					}
				} else {
					log.Error(err.Error())
				}
			}
		}
	}

	// Check for lacks of collisions since last frame.
	for sourceID, targetID := range w.collidingActors {
		if _, ok := collidingActors[sourceID]; !ok {
			w.scripting.To(sourceID).Events.RunLeave(targetID)
		}
	}

	// Store this frame's colliding actors for next frame.
	w.collidingActors = collidingActors
	return nil
}
