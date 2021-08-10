package uix

import (
	"errors"
	"time"

	"git.kirsle.net/apps/doodle/pkg/balance"
	"git.kirsle.net/apps/doodle/pkg/collision"
	"git.kirsle.net/apps/doodle/pkg/log"
	"git.kirsle.net/apps/doodle/pkg/physics"
	"git.kirsle.net/apps/doodle/pkg/scripting"
	"git.kirsle.net/go/render"
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
	// NOTE: parallelism wasn't good for race conditions like the Thief
	//       trying to take your inventory.
	// var wg sync.WaitGroup
	for i, a := range w.actors {
		// wg.Add(1)
		//go
		func(i int, a *Actor) {
			// defer wg.Done()
			originalPositions[a.ID()] = a.Position()

			// Advance any animations for this actor.
			if a.activeAnimation != nil && a.activeAnimation.nextFrameAt.Before(now) {
				if done := a.TickAnimation(a.activeAnimation); done {
					// Animation has finished, get the callback function.
					callback := a.animationCallback

					// Clean up the animation state, in case the callback wants
					// to immediately play another animation.
					a.StopAnimation()

					// Call the callback function.
					if callback.IsFunction() {
						callback.Call(otto.NullValue())
					}

				}
			}

			// Get the actor's velocity to see if it's moving this tick.
			v := a.Velocity()

			// Apply gravity to the actor's velocity.
			if a.hasGravity && !a.Grounded() { //v.Y >= 0 {
				if !a.Grounded() {
					v.Y = physics.Lerp(
						v.Y,             // current speed
						balance.Gravity, // target max gravity falling downwards
						balance.GravityAcceleration,
					)
				} else {
					v.Y = 0
				}
				a.SetVelocity(v)
				// v.Y += balance.Gravity
			}

			// If not moving, grab the bounding box right now.
			if v.IsZero() {
				boxes[i] = collision.GetBoundingRect(a)
				return
			}

			// Create a delta point from their current location to where they
			// want to move to this tick.
			delta := physics.VectorFromPoint(a.Position())
			delta.Add(v)

			// Check collision with level geometry.
			chkPoint := delta.ToPoint()
			info, ok := collision.CollidesWithGrid(a, w.chunks, chkPoint)
			if ok {
				// Collision happened with world.
				if w.OnLevelCollision != nil {
					w.OnLevelCollision(a, info)
				}
			}

			// Move us back where the collision check put us
			if !a.noclip {
				delta = physics.VectorFromPoint(info.MoveTo)
			}

			// Move the actor's World Position to the new location.
			a.MoveTo(delta.ToPoint())

			// Keep the actor from leaving the world borders of bounded maps.
			w.loopContainActorsInsideLevel(a)

			// Store this actor's bounding box after they've moved.
			boxes[i] = collision.SizePlusHitbox(collision.GetBoundingRect(a), a.Hitbox())
		}(i, a)
		// wg.Wait()
	}

	var collidingActors = map[*Actor]*Actor{}
	for tuple := range collision.BetweenBoxes(boxes) {
		a, b := w.actors[tuple.A], w.actors[tuple.B]

		// If neither actor is mobile, don't run collision handlers.
		if !(a.IsMobile() || b.IsMobile()) {
			continue
		}

		collidingActors[a] = b

		// log.Error("between boxes: %+v  <%s> <%s>", tuple, a.ID(), b.ID())

		// Call the OnCollide handler for A informing them of B's intersection.
		if w.scripting != nil {
			var (
				rect        = collision.SizePlusHitbox(collision.GetBoundingRect(b), b.Hitbox())
				lastGoodBox = render.Rect{
					X: originalPositions[b.ID()].X,
					Y: originalPositions[b.ID()].Y,
					W: boxes[tuple.B].W,
					H: boxes[tuple.B].H,
				}
			)

			// HACK: below, when we determine the moving actor is "onTop" of
			// the doodad's solid hitbox, we lockY their movement so they don't
			// fall down further; but sometimes there's an off-by-one error if
			// the actor fell a distance before landing, and so the final
			// Settled collision check doesn't fire (i.e. if they fell onto a
			// Crumbly Floor which should begin shaking when walked on).
			//
			// When we decide they're onTop, record the Y position, and then
			// use it for collision-check purposes but DON'T physically move
			// the character by it (moving the character may clip them thru
			// other solid hitboxes like the upside-down trapdoor)
			var onTopY int

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

				// Special case for when a mobile actor lands ON TOP OF a solid
				// actor. We want to stop their Y movement downwards, but allow
				// horizontal movement on the X axis.
				// Touching the solid actor from the side is already fine.
				var onTop = false

				var (
					lockX int
					lockY int
				)

				for point := range render.IterLine(
					origPoint,
					b.Position(),
				) {
					point := point
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
							// Are they on top?
							aHitbox := collision.SizePlusHitbox(collision.GetBoundingRect(a), a.Hitbox())
							if render.AbsInt(test.Y+test.H-aHitbox.Y) == 0 {
								// log.Error("ActorCollision: onTop=true at Y=%s", test.Y)
								onTop = true
								onTopY = test.Y
							}

							// What direction were we moving?
							if test.Y != lastGoodBox.Y {
								if lockY == 0 {
									lockY = lastGoodBox.Y
								}
								if onTop {
									// log.Error("ActorCollision: setGrounded(true)", test.Y)
									b.SetGrounded(true)
								}
							}
							if test.X != lastGoodBox.X {
								if !onTop {
									lockX = lastGoodBox.X
								}
							}

							// Move them back to the last good box.
							lastGoodBox = test
							if lockX != 0 {
								lastGoodBox.X = lockX
							}
						} else {
							// Move them back to the last good box.
							lastGoodBox = test
						}
					} else {
						// No collision between boxes, increment the lastGoodBox
						lastGoodBox = test
					}
				}

				// Did we lock their X or Y coordinate from moving further?
				if lockY != 0 {
					lastGoodBox.Y = lockY
				}
				if lockX != 0 {
					lastGoodBox.X = lockX
				}

				if !b.noclip {
					b.MoveTo(lastGoodBox.Point())
				}
			} else {
				log.Error(
					"ERROR: Actors %s and %s overlap and the script returned false,"+
						"but I didn't store %s original position earlier??",
					a.Doodad().Title, b.Doodad().Title, b.Doodad().Title,
				)
			}

			if onTopY != 0 && lastGoodBox.Y-onTopY <= 1 {
				lastGoodBox.Y = onTopY
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
					log.Error("VM(%s).RunCollide: %s", a.ID(), err.Error())
				}

				// If the (player) is pressing the Use key, call the colliding
				// actor's OnUse event.
				if b.flagUsing {
					if err := w.scripting.To(a.ID()).Events.RunUse(&UseEvent{
						Actor: b,
					}); err != nil {
						log.Error("VM(%s).RunUse: %s", a.ID(), err.Error())
					}
				}
			}
		}
	}

	// Check for lacks of collisions since last frame.
	for sourceActor, targetActor := range w.collidingActors {
		if _, ok := collidingActors[sourceActor]; !ok {
			w.scripting.To(sourceActor.ID()).Events.RunLeave(&CollideEvent{
				Actor:   targetActor,
				Settled: true,
			})
		}
	}

	// Store this frame's colliding actors for next frame.
	w.collidingActors = collidingActors
	return nil
}
