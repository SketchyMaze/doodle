package collision

import (
	"errors"

	"git.kirsle.net/SketchyMaze/doodle/pkg/balance"
	"git.kirsle.net/SketchyMaze/doodle/pkg/level"
	"git.kirsle.net/go/render"
)

// Collide describes how a collision occurred.
type Collide struct {
	Top         bool
	TopPoint    render.Point
	TopPixel    *level.Swatch
	Left        bool
	LeftPoint   render.Point
	LeftPixel   *level.Swatch
	Right       bool
	RightPoint  render.Point
	RightPixel  *level.Swatch
	Bottom      bool
	BottomPoint render.Point
	BottomPixel *level.Swatch
	MoveTo      render.Point

	// Swatch attributes affecting the collision at this time.
	InFire     string // the name of the swatch, Fire = general ouchy color.
	InWater    bool
	IsSlippery bool
}

// Reset a Collide struct flipping all the bools off, but keeping MoveTo.
func (c *Collide) Reset() {
	c.Top = false
	c.Left = false
	c.Right = false
	c.Bottom = false
	c.InWater = false
}

// Side of the collision box (top, bottom, left, right)
type Side uint8

// Options for the Side type.
const (
	Top Side = iota
	Bottom
	Left
	Right
)

/*
CollidesWithGrid checks if a Doodad collides with level geometry.

The `target` is the point the actor wants to move to on this tick.

This function handles translation for doodads having an offset hitbox which doesn't begin
at the 0,0 coordinate on the X,Y axis.

For example:

  - The caller of this function cares about where on screen to display the actor's sprite at
    (the X,Y position of the top left corner of their sprite).
  - The target point is where the actor is moving on this tick, which is also relative to their
    current world coordinate (top corner of their sprite), NOT their hitbox coordinate.

The original collision detection code worked well when the actor's hitbox began at 0,0 as it
matched their world position. But when the hitbox is offset from the corner, collision detection
glitches abounded.

This function will compute the physical hitbox of the doodad regarding the level geometry
(simulating a simple doodad whose hitbox is a full 0,0,W,H) and translate the offset to and
from.
*/
func CollidesWithGrid(d Actor, grid *level.Chunker, target render.Point) (*Collide, bool) {
	var (
		actor     = NewActorOffset(d)
		offset    = actor.Offset()
		newTarget = render.Point{
			X: target.X + offset.X,
			Y: target.Y + offset.Y,
		}
	)

	collide, ok := BoxCollidesWithGrid(actor, grid, newTarget)

	// Undo the offset for the MoveTo target.
	collide.MoveTo.X -= offset.X
	collide.MoveTo.Y -= offset.Y

	return collide, ok
}

/*
BoxCollidesWithGrid handles the core logic for level collision checks.
*/
func BoxCollidesWithGrid(d Actor, grid *level.Chunker, target render.Point) (*Collide, bool) {
	var (
		P      = d.Position()
		S      = d.Size()
		hitbox = d.Hitbox()

		result = &Collide{
			MoveTo: P,
		}
		ceiling   bool // Has hit a ceiling?
		capHeight int  // Stop vertical movement thru a ceiling
		capLeft   int  // Stop movement thru a wall
		capRight  int
		capFloor  int  // Stop movement thru the floor
		hitLeft   bool // Has hit an obstacle on the left
		hitRight  bool // or right
		hitFloor  bool
	)

	// Adjust the actor's bounding rect by its stated Hitbox from its script.
	// e.g.: Boy's Canvas size is 56x56 but he is a narrower character with a
	// hitbox width smaller than its Canvas size.
	S = GetBoundingRectHitbox(d, hitbox)
	actorHeight := P.Y + S.H

	// Wrap the grid in a chunkCache for the duration of this call. See the
	// chunkCache doc comment: this scan can query well over a hundred points
	// against `grid`, almost all of which land in the same chunk, so this
	// memoizes the chunk lookup across every ScanBoundingBox call below.
	cache := newChunkCache(grid)

	// Test if we are ALREADY colliding with level geometry and try and wiggle
	// free. ScanBoundingBox scans level pixels along the four edges of the
	// actor's hitbox in world space.
	if ok := result.ScanBoundingBox(GetBoundingRectHitbox(d, hitbox), cache); ok {
		// We've already collided! Try to wiggle free.
		if result.Bottom {
			if !d.Grounded() {
				d.SetGrounded(true)
			}
		} else {
			d.SetGrounded(false)
		}
		if result.Top {
			ceiling = true
			P.Y++
		}
		if result.Left && !result.LeftPixel.SemiSolid {
			P.X++
		}
		if result.Right && !result.RightPixel.SemiSolid {
			P.X--
		}
	}

	// If grounded, cap our Y position.
	if d.Grounded() {
		if !result.Bottom {
			// We've fallen off a ledge.
			d.SetGrounded(false)
		} else if target.Y < P.Y {
			// We're moving upward.
			d.SetGrounded(false)
		} else {
			// Cap our downward motion to our current position,
			// fighting the force of gravity.
			target.Y = P.Y
		}
	}

	// Cap our horizontal movement if we're touching walls.
	if (result.Left && target.X < P.X) || (result.Right && target.X > P.X) {
		// Handle walking up slopes, if the step is short enough.
		var slopeHeight int
		if result.Left {
			slopeHeight = result.LeftPoint.Y
		} else if result.Right {
			slopeHeight = result.RightPoint.Y
		}

		if offset, ok := CanStepUp(actorHeight, slopeHeight, target.X > P.X); ok {
			target.Add(offset)
		} else {
			// Not a slope.. may be a solid wall. If the wall is a SemiSolid though,
			// do not cap our direction just yet.
			if !(result.Left && result.LeftPixel.SemiSolid) && !(result.Right && result.RightPixel.SemiSolid) {
				target.X = P.X
			}
		}
	}

	// Cap our vertical movement if we're touching ceilings.
	if ceiling {
		// The existing box intersects a ceiling, this will almost never
		// happen because gravity will always pull you away at the last frame.
		// But if we do somehow get here, may as well cap it where it's at.
		capHeight = P.Y
	}

	// Trace a line from where we are to where we wanna go.
	result.Reset()
	result.MoveTo = P
	render.WalkLine(P, target, func(point render.Point) {
		// Before we compute their next move, if we're already capping their
		// height make sure the new point stays capped too. This prevents them
		// clipping thru a ceiling if they were also holding right/left too.
		if capHeight != 0 && point.Y < capHeight {
			point.Y = capHeight
		}
		if capLeft != 0 && point.X < capLeft {
			// TODO: this along with a "+ 1" hack prevents clipping thru the
			// left wall sometimes, but breaks walking up leftward slopes.
			point.X = capLeft
		}
		if capRight != 0 && point.X > capRight {
			// This if check fixes the climbing-walls-on-the-right bug.
			point.X = capRight
		}

		if has := result.ScanBoundingBox(render.Rect{
			X: point.X,
			Y: point.Y,
			W: S.W,
			H: S.H,
		}, cache); has {
			if result.Bottom {
				if !hitFloor {
					hitFloor = true
					capFloor = result.BottomPoint.Y - S.H
				}
				d.SetGrounded(true)
			}

			if result.Top && !ceiling {
				// This is a newly discovered ceiling.
				ceiling = true
				capHeight = result.TopPoint.Y + 1
				// TODO: the "+ 1" helps prevent clip thru ceiling, probably.
				// Similar to the "+ 1" on the left side, below.
			}

			// TODO: this block of code is interesting. For SemiSolid slopes, the character
			// walks up the slopes FAST (full speed) which is nice; would like to do this
			// for regular solid slopes too. But if this block of code is dummied out for
			// solid walls, the player is able to clip thru thin walls (couple px thick); the
			// capLeft/capRight behavior is good at stopping the player here.

			// See if they have hit a solid wall on their left or right edge. If the wall
			// is short enough to step up, allow them to pass through.
			if result.Left && !hitLeft && !result.LeftPixel.SemiSolid {
				if _, ok := CanStepUp(actorHeight, result.LeftPoint.Y, false); !ok {
					hitLeft = true
					capLeft = result.LeftPoint.X
				}
			}
			if result.Right && !hitRight && !result.RightPixel.SemiSolid {
				if _, ok := CanStepUp(actorHeight, result.RightPoint.Y, false); !ok {
					hitRight = true
					capRight = result.RightPoint.X - S.W
				}
			}
		}

		// So far so good, keep following the MoveTo to
		// the last good point before a collision.
		result.MoveTo = point
	})

	// If they hit the roof, cap them to the roof.
	if ceiling && result.MoveTo.Y < capHeight {
		result.Top = true
		result.MoveTo.Y = capHeight
	}
	if hitFloor && result.MoveTo.Y > capFloor {
		result.Bottom = true
		result.MoveTo.Y = capFloor
	}
	if hitLeft && !result.LeftPixel.SemiSolid {
		result.Left = true
		result.MoveTo.X = capLeft
	}
	if hitRight && !result.RightPixel.SemiSolid {
		result.Right = true
		result.MoveTo.X = capRight
	}

	return result, result.IsColliding()
}

/*
CanStepUp checks whether the actor is moving left or right onto a gentle slope which
they can step on top of instead of being blocked by the solid wall.

* actorHeight is the actor's Y position + their hitbox height.
* slopeHeight is the Y position of the left or right edge of the level they collide with.
* moveRight is true if moving right, false if moving left.

If the actor can step up the slope, the return value is the Point of how to offset their
X,Y position to move up the slope and the boolean is whether they can step up.
*/
func CanStepUp(actorHeight, slopeHeight int, moveRight bool) (render.Point, bool) {
	var (
		height = actorHeight - slopeHeight
		target render.Point
	)

	if height <= balance.SlopeMaxHeight {
		target.Y -= height
		if moveRight {
			target.X++
		} else {
			target.X--
		}

		return target, true
	}

	return target, false
}

// IsColliding returns whether any sort of collision has occurred.
func (c *Collide) IsColliding() bool {
	return c.Top || c.Bottom || (c.Left && !c.LeftPixel.SemiSolid) || (c.Right && !c.RightPixel.SemiSolid) ||
		c.InFire != "" || c.InWater
}

// Grid is the minimal interface ScanBoundingBox and ScanGridLine need to
// query level pixels. *level.Chunker satisfies it directly; chunkCache
// (below) wraps a Chunker and satisfies it too, adding memoization for the
// hot collision-scanning path in BoxCollidesWithGrid.
type Grid interface {
	Get(render.Point) (*level.Swatch, error)
}

// chunkCache memoizes every chunk looked up during a scan against the level
// grid.
//
// BoxCollidesWithGrid can query well over a hundred points per doodad per
// tick: four edges of the hitbox, rescanned once per pixel of movement along
// the path. Chunker.Get resolves the chunk for a point through three
// separate mutex sections (the chunk map, the LRU "chopping block", and the
// per-tick access log) before it even looks up the pixel. Since a doodad's
// hitbox and its per-tick movement are almost always tiny next to a chunk
// (see balance.ChunkSize, 128px), a single scan only ever touches a handful
// of chunks -- at most about 4, for a hitbox straddling a chunk corner while
// moving diagonally. Remembering all of them (not just the last one) means a
// hitbox that straddles a chunk boundary and keeps crossing back and forth
// between two chunks -- e.g. its top edge in one chunk, bottom edge in the
// next -- still only pays Chunker.GetChunk's mutex/bookkeeping cost once per
// chunk for the whole scan, rather than once per crossing.
//
// A chunkCache is only ever touched by the one goroutine running a given
// doodad's collision check (this package never scans two doodads
// concurrently), so it needs no locking of its own.
type chunkCache struct {
	grid   *level.Chunker
	chunks map[render.Point]cachedChunk

	// last* is a fast path for the common case of consecutive points
	// landing in the same chunk as the previous one, which avoids even the
	// map lookup below.
	lastCoord render.Point
	lastEntry cachedChunk
	haveLast  bool
}

type cachedChunk struct {
	chunk *level.Chunk
	found bool
}

// errNoChunk is returned in place of Chunker.Get's fmt.Errorf when no chunk
// covers a point. Most of a level is undrawn open space with no chunk ever
// created for it, so this is the common case for a good fraction of the
// points a collision scan tests -- not worth an allocation and a format
// call every time it happens.
var errNoChunk = errors.New("no chunk for point")

func newChunkCache(grid *level.Chunker) *chunkCache {
	// A doodad's hitbox scan realistically touches at most ~4 chunks (see
	// the chunkCache doc comment), so pre-size the map for that.
	return &chunkCache{
		grid:   grid,
		chunks: make(map[render.Point]cachedChunk, 4),
	}
}

// Get behaves like Chunker.Get but reuses a chunk this cache has already
// looked up, if the point falls within one.
func (cc *chunkCache) Get(p render.Point) (*level.Swatch, error) {
	coord := cc.grid.ChunkCoordinate(p)

	var entry cachedChunk
	if cc.haveLast && coord == cc.lastCoord {
		entry = cc.lastEntry
	} else if cached, ok := cc.chunks[coord]; ok {
		entry = cached
	} else {
		chunk, found := cc.grid.GetChunk(coord)
		entry = cachedChunk{chunk: chunk, found: found}
		cc.chunks[coord] = entry
	}

	cc.lastCoord = coord
	cc.lastEntry = entry
	cc.haveLast = true

	if !entry.found {
		return nil, errNoChunk
	}
	return entry.chunk.Get(p)
}

// ScanBoundingBox scans all of the pixels in a bounding box on the grid and
// returns if any of them intersect with level geometry.
func (c *Collide) ScanBoundingBox(box render.Rect, grid Grid) bool {
	col := GetCollisionBox(box)

	// Each side only ever sets its own dedicated fields (c.Top/TopPoint/...,
	// etc.) and never clears them, so calling these directly on `c` in a
	// fixed order is equivalent to the old parallel-scan-then-merge and
	// preserves the same "sticky" semantics BoxCollidesWithGrid relies on
	// across repeated ScanBoundingBox calls along a movement path, and the
	// same last-writer-wins order for the shared InFire/InWater/IsSlippery
	// fields when more than one side finds them this call.
	c.ScanGridLine(col.Top[0], col.Top[1], grid, Top)
	c.ScanGridLine(col.Bottom[0], col.Bottom[1], grid, Bottom)
	c.ScanGridLine(col.Left[0], col.Left[1], grid, Left)
	c.ScanGridLine(col.Right[0], col.Right[1], grid, Right)

	return c.IsColliding()
}

// ScanGridLine scans all of the pixels between p1 and p2 on the grid and tests
// for any pixels to be set, implying a collision between level geometry and the
// bounding boxes of the doodad.
func (c *Collide) ScanGridLine(p1, p2 render.Point, grid Grid, side Side) {
	// If scanning the top or bottom line, offset the X coordinate by 1 pixel.
	// This is because the 4 corners of the bounding box share their corner
	// pixel with each side, so the Left and Right edges will check the
	// left- and right-most point.
	if side == Top || side == Bottom {
		p1.X++
		p2.X--
	}

	render.WalkLine(p1, p2, func(point render.Point) {
		if swatch, err := grid.Get(point); err == nil {
			// We're intersecting a pixel! If it's a solid one we'll return it
			// in our result. If non-solid, we'll collect attributes from it
			// and return them in the final result for gameplay behavior.
			if swatch.Fire {
				c.InFire = swatch.Name
			}
			if swatch.Water {
				c.InWater = true
			}

			// Slippery floor?
			if side == Bottom && swatch.Slippery {
				c.IsSlippery = true
			}

			// Non-solid swatches don't collide so don't pay them attention.
			if !swatch.Solid && !swatch.SemiSolid {
				return
			}

			// A semisolid only has collision on the bottom (and a little on the
			// sides, for slope walking only)
			if swatch.SemiSolid && side == Top {
				return
			}

			switch side {
			case Top:
				c.Top = true
				c.TopPoint = point
				c.TopPixel = swatch
			case Bottom:
				c.Bottom = true
				c.BottomPoint = point
				c.BottomPixel = swatch
			case Left:
				c.Left = true
				c.LeftPoint = point
				c.LeftPixel = swatch
			case Right:
				c.Right = true
				c.RightPoint = point
				c.RightPixel = swatch
			}
		}
	})
}
