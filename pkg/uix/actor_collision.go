package uix

import (
	"errors"
	"sync"
	"time"

	"git.kirsle.net/go/render"
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
				if w.OnLevelCollision != nil {
					w.OnLevelCollision(a, info)
				}
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

		// Call the OnCollide handler for A informing them of B's intersection.
		if w.scripting != nil {
			var (
				rect        = doodads.GetBoundingRect(b)
				lastGoodBox = boxes[tuple.B] // worst case scenario we get blocked right away
			)

			// Firstly we want to make sure B isn't able to clip through A's
			// solid hitbox if A protests the movement. Trace a vector from
			// B's original position to their current one and ping A's
			// OnCollide handler for each step, with Settled=false. A should
			// only return false if it protests the movement, but not trigger
			// any actions (such as emit messages to linked doodads) until
			// Settled=true.
			if origPoint, ok := originalPositions[b.ID()]; ok {
				// Trace a vector back from the actor's current position
				// to where they originated from. If A protests B's position at
				// ANY time, we mark didProtest=true and continue backscanning
				// B's movement. The next time A does NOT protest, that is to be
				// B's new position.

				var firstPoint = true
				for point := range render.IterLine(
					origPoint,
					b.Position(),
				) {
					test := render.Rect{
						X: point.X,
						Y: point.Y,
						W: rect.W,
						H: rect.H,
					}

					if info, err := collision.CompareBoxes(boxes[tuple.A], test); err == nil {
						// B is overlapping A's box, call its OnCollide handler
						// with Settled=false and see if it protests the overlap.
						err := w.scripting.To(a.ID()).Events.RunCollide(&CollideEvent{
							Actor:    b,
							Overlap:  info.Overlap,
							InHitbox: info.Overlap.Intersects(a.Hitbox()),
							Settled:  false,
						})

						// Did A protest?
						if err == scripting.ErrReturnFalse {
							break
						} else {
							lastGoodBox = test
						}
					}

					firstPoint = false
				}

				// Were we stopped before we even began?
				if firstPoint {
					// TODO: undo the effect of gravity this tick. Use case:
					// the player lands on top of a solid door, and their
					// movement is blocked the first step by the door. Originally
					// he'd continue falling, so I had to move him up to stop it,
					// turns out moving up by the -gravity is exactly the distance
					// to go. Don't know why.
					b.MoveBy(render.NewPoint(0, int32(-balance.Gravity)))
				} else {
					b.MoveTo(lastGoodBox.Point())
				}
			} else {
				log.Error(
					"ERROR: Actors %s and %s overlap and the script returned false,"+
						"but I didn't store %s original position earlier??",
					a.Doodad.Title, b.Doodad.Title, b.Doodad.Title,
				)
			}

			// Movement has been settled. Check if B's point is still invading
			// A's box and call its OnCollide handler one last time in
			// Settled=true mode so it can run its actions.
			if info, err := collision.CompareBoxes(boxes[tuple.A], lastGoodBox); err == nil {
				if err := w.scripting.To(a.ID()).Events.RunCollide(&CollideEvent{
					Actor:    b,
					Overlap:  info.Overlap,
					InHitbox: info.Overlap.Intersects(a.Hitbox()),
					Settled:  true,
				}); err != nil && err != scripting.ErrReturnFalse {
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
