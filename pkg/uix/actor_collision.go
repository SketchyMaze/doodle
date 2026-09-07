package uix

import (
	"errors"
	"time"

	"git.kirsle.net/SketchyMaze/doodle/pkg/balance"
	"git.kirsle.net/SketchyMaze/doodle/pkg/collision"
	"git.kirsle.net/SketchyMaze/doodle/pkg/log"
	"git.kirsle.net/SketchyMaze/doodle/pkg/physics"
	"git.kirsle.net/SketchyMaze/doodle/pkg/scripting"
	"git.kirsle.net/go/render"
	"github.com/dop251/goja"
)

// loopActorCollision is the Loop function that checks if pairs of
// actors are colliding with each other, and handles their scripting
// responses to such collisions.
func (w *Canvas) loopActorCollision() error {
	if w.scripting == nil {
		return errors.New("Canvas.loopActorCollision: scripting engine not attached to Canvas")
	}

	// As we iterate over all actors below to process their movement, track
	// their bounding rectangles so we can later see if any pair of actors
	// intersect each other. Also, in case of actor scripts protesting a
	// collision later, store each actor's original position before the move.
	//
	// These are scratch buffers owned by the Canvas and reused (not
	// reallocated) every tick -- see their declaration for why.
	if cap(w.collisionBoxesBuf) < len(w.actors) {
		w.collisionBoxesBuf = make([]render.Rect, len(w.actors))
	}
	var (
		// Current time of this tick so we can advance animations.
		now = time.Now()

		boxes             = w.collisionBoxesBuf[:len(w.actors)]
		originalPositions = w.collisionOrigPosBuf
		originalHitboxes  = w.collisionOrigHitboxBuf
	)
	clear(boxes)
	if originalPositions == nil {
		originalPositions = map[string]render.Point{}
		w.collisionOrigPosBuf = originalPositions
	} else {
		clear(originalPositions)
	}
	if originalHitboxes == nil {
		originalHitboxes = map[string]render.Rect{}
		w.collisionOrigHitboxBuf = originalHitboxes
	} else {
		clear(originalHitboxes)
	}

	// Loop over all the actors in parallel, processing their movement and
	// checking collision data against the level geometry.
	// NOTE: parallelism wasn't good for race conditions like the Thief
	//       trying to take your inventory.
	// var wg sync.WaitGroup
	for i, a := range w.actors {
		if a.IsFrozen() {
			continue
		}

		// wg.Add(1)
		//go
		func(i int, a *Actor) {
			// defer wg.Done()
			originalPositions[a.ID()] = a.Position()
			originalHitboxes[a.ID()] = collision.GetBoundingRectHitbox(a, a.Hitbox())

			// Advance any animations for this actor.
			// TODO: wallclock time here, should be set by FPS for consistency.
			if a.activeAnimation != nil && a.activeAnimation.nextFrameAt.Before(now) {
				if done := a.TickAnimation(a.activeAnimation); done {
					// Animation has finished, get the callback function.
					callback := a.animationCallback

					// Clean up the animation state, in case the callback wants
					// to immediately play another animation.
					a.StopAnimation()

					// Call the callback function.
					if function, ok := goja.AssertFunction(callback); ok {
						function(goja.Undefined())
					}

				}
			}

			// Get the actor's velocity to see if it's moving this tick.
			v := a.Velocity()

			// Apply gravity to the actor's velocity.
			if a.hasGravity && !a.Grounded() { //v.Y >= 0 {
				if !a.Grounded() {
					var (
						gravity      = balance.GravityMaximum
						acceleration = balance.GravityAcceleration
					)
					if a.IsWet() {
						gravity = balance.SwimGravity
					}

					// If the actor is jumping/moving upwards, apply softer gravity.
					if v.Y < 0 {
						acceleration = balance.GravityJumpAcceleration
					}

					v.Y = physics.Lerp(
						v.Y,     // current speed
						gravity, // target max gravity falling downwards
						acceleration,
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
			info, _ := collision.CollidesWithGrid(a, w.chunks, chkPoint)

			// Inform the caller about the collision state every tick
			if w.OnLevelCollision != nil {
				w.OnLevelCollision(a, info)
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
			boxes[i] = collision.GetBoundingRect(a)
		}(i, a)
		// wg.Wait()
	}

	// log.Warn("== BEGIN BetweenBoxes")

	// Check pairs of all our Actor boxes for overlap and running their OnCollide
	// scripts for mobile actors.
	var collidingActors = ActorCollisionMap{}
	for tuple := range collision.BetweenBoxes(boxes) {

		// Give the A, B tuple of boxes names: their order doesn't matter.
		// Example: stable could be the Button and mover is the Player walking onto it.
		// Or: stable could be the Player and mover is a Key that they walked onto.
		stable, mover := w.actors[tuple.A], w.actors[tuple.B]

		// If neither actor is mobile, don't run collision handlers.
		if !(stable.IsMobile() || mover.IsMobile()) {
			continue
		}

		collidingActors.Set(stable, mover)

		// log.Error("between boxes: %+v  A=<%s>  B=<%s>", tuple, stable.ID(), mover.ID())

		// Call the OnCollide handler for A informing them of B's intersection.
		if w.scripting != nil {
			var (
				// rect is the mover's hitbox rect (world coordinates) at their
				// position pending this actor-vs-actor collision pass.
				rect = collision.GetBoundingRectHitbox(mover, mover.Hitbox())

				// lastGoodBox tracks the mover's hitbox rect (world coordinates,
				// sized to their declared Hitbox) as we trace their movement below.
				// It starts at their pre-move hitbox and is only converted back to
				// a sprite-corner point (what MoveTo expects) once, right before
				// we actually move them.
				lastGoodBox = originalHitboxes[mover.ID()]
			)

			// Below, when we determine the moving actor is "onTop" (or hitting
			// the underside) of the doodad's solid hitbox, we lockY their
			// movement flush against that edge so they don't fall down further
			// (or rise up further). It's snapped exactly to the edge, not left
			// wherever the collision trace happened to detect the protest, so
			// the final Settled overlap check below reliably registers the
			// touch (e.g. Crumbly Floor's shake trigger).

			// Firstly we want to make sure B isn't able to clip through A's
			// solid hitbox if A protests the movement. Trace a vector from
			// B's original position to their current one and ping A's
			// OnCollide handler for each step, with Settled=false. A should
			// only return false if it protests the movement, but not trigger
			// any actions (such as emit messages to linked doodads) until
			// Settled=true.
			if origHitbox, ok := originalHitboxes[mover.ID()]; ok {

				var (
					// Special case for when a mobile actor lands ON TOP OF a solid
					// actor. We want to stop their Y movement downwards, but allow
					// horizontal movement on the X axis.
					// Touching the solid actor from the side is already fine.
					onTop    bool
					onBottom bool // they hit the bottom instead
					onLeft   bool // mover is to the stable's left, touching its left edge
					onRight  bool // mover is to the stable's right, touching its right edge

					// If we lock their movement coordinate.
					lockX *int
					lockY *int
				)

				// If their original hitbox is offset from their sprite corner,
				// gather the offset now.
				var (
					origPosition  = originalPositions[mover.ID()]
					hitboxPadding = render.Point{
						X: origHitbox.X - origPosition.X,
						Y: origHitbox.Y - origPosition.Y,
					}
				)

				// Trace a vector back from the mover's current position
				// to where they originated from. If A protests B's position at
				// ANY time, we ?mark didProtest=true? and continue backscanning
				// B's movement. The next time A does NOT protest, that is to be
				// B's new position.
				for point := range render.IterLine(
					origHitbox.Point(),
					rect.Point(),
				) {
					point := point

					// Once an axis has been locked (the mover is resting flush
					// against a solid edge), don't keep probing points further
					// along that axis. Otherwise we keep re-invoking the stable
					// doodad's OnCollide with points deeper than where the mover
					// will actually end up, which can corrupt doodads that keep
					// state across calls: e.g. Trapdoor's "opened" flag could
					// flip true from a spurious deep-overlap probe fired after
					// the correct landing point was already found, causing it
					// to silently stop being solid on the very next tick.
					if lockY != nil {
						point.Y = *lockY
					}
					if lockX != nil {
						point.X = *lockX
					}

					test := render.Rect{
						X: point.X,
						Y: point.Y,
						W: rect.W,
						H: rect.H,
					}

					if info, err := collision.CompareBoxes(boxes[tuple.A], test); err == nil {
						// A and B have their drawings overlapping on the page. Get each
						// of their declared hitboxes (if smaller) to see if their hitboxes
						// intersect as well.
						var (
							stableHitbox = collision.GetBoundingRectHitbox(stable, stable.Hitbox())
							moverHitbox  = collision.GetBoundingRectHitbox(mover, mover.Hitbox())
						)

						// B is overlapping A's box, call its OnCollide handler
						// with Settled=false and see if it protests the overlap.
						err := w.scripting.To(stable.ID()).Events.RunCollide(&CollideEvent{
							Actor:    mover,
							Overlap:  info.Overlap,
							InHitbox: stableHitbox.Intersects(moverHitbox),
							Settled:  false,
						})

						// log.Warn("ActorCollision: CompareBoxes info was %+v", info)

						// Did A protest?
						if err == scripting.ErrReturnFalse {
							// Are they on top?
							var (
								stableTop    = stableHitbox.Y
								stableBottom = stableHitbox.Y + stableHitbox.H
								moverTop     = test.Y
								moverBottom  = test.Y + test.H // bottom of falling actor
							)

							// Is the colliding actor on top? (e.g. mover=player character)
							if render.AbsInt(moverBottom-stableTop) < balance.OnTopThreshold {
								onTop = true
							}

							// Or are they hitting from below?
							if render.AbsInt(stableBottom-moverTop) < balance.OnTopThreshold {
								onBottom = true
							}

							// Same idea, but for the horizontal axis: is the mover
							// touching our left or right edge?
							var (
								stableLeft  = stableHitbox.X
								stableRight = stableHitbox.X + stableHitbox.W
								moverLeft   = test.X
								moverRight  = test.X + test.W
							)
							if render.AbsInt(moverRight-stableLeft) < balance.OnTopThreshold {
								onLeft = true
							}
							if render.AbsInt(stableRight-moverLeft) < balance.OnTopThreshold {
								onRight = true
							}

							// if onTop || onBottom {
							// 	log.Error("onTop=%+v onBottom=%+v", onTop, onBottom)
							// }

							// What direction were we moving?
							if test.Y != lastGoodBox.Y {

								// If we are hitting the top or bottom, lock our Y coordinate here.
								if onTop || onBottom {

									// Lock the Y coordinate flush against the stable hitbox's
									// edge (rather than "one step back" from the protested
									// point) so the mover's hitbox exactly touches it. Leaving
									// so much as a 1px gap fails the final Settled overlap
									// check below and e.g. Crumbly Floor's shake trigger would
									// intermittently not fire.
									if lockY == nil {
										lockY = new(int)
										switch {
										case onTop:
											*lockY = stableTop - test.H
										case onBottom:
											*lockY = stableBottom
										default:
											*lockY = lastGoodBox.Y
										}
									}

									// If on top, set the mover to Grounded here.
									if onTop {
										mover.SetGrounded(true)
									}
								}

							}
							if test.X != lastGoodBox.X {
								if lockX == nil && !(onTop || onBottom) {
									// Lock the X coordinate flush against the stable
									// hitbox's edge for the same reason as the Y lock
									// above: doors like Trapdoor Left/Right need the
									// mover's hitbox to exactly touch theirs for the
									// final Settled overlap check to reliably register
									// the contact.
									lockX = new(int)
									switch {
									case onLeft:
										*lockX = stableLeft - test.W
									case onRight:
										*lockX = stableRight
									default:
										*lockX = lastGoodBox.X
									}
								}
							}

							// Move them back to the last good box (world coordinates of
							// their hitbox rect).
							lastGoodBox = render.Rect{
								X: test.X,
								Y: test.Y,
								W: test.W,
								H: test.H,
							}
						} else {
							if err != nil {
								log.Error("RunCollide on %s (%s) errored: %s", stable.ID(), stable.Actor.Filename, err)
							}
							// Move them back to the last good box.
							lastGoodBox = test
						}
					} else {
						// No collision between boxes, increment the lastGoodBox
						lastGoodBox = test
					}
				}

				// Did we lock their X or Y coordinate from moving further?
				if lockY != nil {
					lastGoodBox.Y = *lockY
				}
				if lockX != nil {
					lastGoodBox.X = *lockX
				}

				if !mover.noclip {
					// lastGoodBox has been tracked in the mover's hitbox coordinate
					// frame (world coordinates of their declared Hitbox); convert it
					// back to their sprite's top-left corner, which is what MoveTo
					// positions them by.
					moveTo := render.Point{
						X: lastGoodBox.X - hitboxPadding.X,
						Y: lastGoodBox.Y - hitboxPadding.Y,
					}
					// log.Error("Move B to: %s", moveTo)

					// The stationary doodad should move the moving one only.
					mover.MoveTo(moveTo)
				}
			} else {
				log.Error(
					"ERROR: Actors %s and %s overlap and the script returned false,"+
						"but I didn't store %s original position earlier??",
					stable.Doodad().Title, mover.Doodad().Title, mover.Doodad().Title,
				)
			}

			// Movement has been settled. Check if B's point is still invading
			// A's box and call its OnCollide handler one last time in
			// Settled=true mode so it can run its actions.
			if info, err := collision.CompareBoxes(boxes[tuple.A], lastGoodBox); err == nil {
				if err := w.scripting.To(stable.ID()).Events.RunCollide(&CollideEvent{
					Actor:    mover,
					Overlap:  info.Overlap,
					InHitbox: info.Overlap.Intersects(stable.Hitbox()),
					Settled:  true,
				}); err != nil && err != scripting.ErrReturnFalse {
					log.Error("VM(%s).RunCollide: %s", stable.ID(), err.Error())
				}

				// If the (player) is pressing the Use key, call the colliding
				// actor's OnUse event.
				if mover.flagUsing {
					if err := w.scripting.To(stable.ID()).Events.RunUse(&UseEvent{
						Actor: mover,
					}); err != nil {
						log.Error("VM(%s).RunUse: %s", stable.ID(), err.Error())
					}
				}
			}
		}
	}

	// log.Warn("-- END BetweenBoxes")

	// Check for lacks of collisions since last frame.
	// Note: w.collidingActors is "last frame's" map of colliding actor boxes.
	w.collidingActors.Iter(func(stable, mover *Actor) {

		// Are these not colliding this frame?
		// TODO: does this work with three-way actor collisions?
		if !collidingActors.Exists(stable, mover) {
			w.scripting.To(stable.ID()).Events.RunLeave(&CollideEvent{
				Actor:   mover,
				Settled: true,
			})
		}
	})

	// Store this frame's colliding actors for next frame.
	w.collidingActors = collidingActors
	return nil
}

// ActorCollisionMap keeps a cache of collision box overlaps between
// an Actor and one or more other Actors.
type ActorCollisionMap map[*Actor]map[*Actor]interface{}

// Set a collision to the other actor.
func (m ActorCollisionMap) Set(stable, mover *Actor) {
	if m[stable] == nil {
		m[stable] = map[*Actor]interface{}{}
	}

	m[stable][mover] = nil
}

// Exists checks if the actor is colliding with the other.
func (m ActorCollisionMap) Exists(stable, mover *Actor) bool {
	if m[stable] == nil {
		return false
	}

	_, ok := m[stable][mover]
	return ok
}

// Iter the collision data.
func (m ActorCollisionMap) Iter(fn func(stable, mover *Actor)) {
	for stable, moverMap := range m {
		for mover := range moverMap {
			fn(stable, mover)
		}
	}
}
