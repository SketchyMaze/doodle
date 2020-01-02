# Doodad Ideas and Implementation Notes

## Warp Doors

Warp Doors would connect two places in the same level and allow the player to
instantly travel between them.

Usage: drag two Warp Doors into the level and Link them together.

When the player activates a Warp Door (i.e. press Up key in front of it), it'd
play an open animation and the player would be warped to the door it's linked to.

A few ideas of variants on the Warp Door:

* Two-state Warp Door (orange and blue). They'd work like the two-state blocks;
  if the orange blocks are in dotted-outline mode, the Orange Door is too and
  can not be opened while the Blue Door can. When the two-state button is pushed,
  orange dotted-outline doors become solid and Blue Doors become dotted-outline.
* Locked door, requiring a key to unlock it (one-time) before entering. If the
  linked door is also locked, it becomes unlocked when the player travels thru
  it. If a Locked Door is linked to a normal one, players can travel into the
  normal door and out the Locked Door but not back again without a key.

## Clocks

Clock doodads would emit timer events globally to be responded to by other
doodads. Clocks would come in a few varieties (i.e. 5 seconds, 10 seconds,
30 seconds).

Each clock's sprite would consist of a clock symbol and the number of seconds
for the clock. Each sprite would be a distinct color.

Clock sprites are **ONLY** visible in the Editor Mode; in Play Mode, the sprite
hides itself immediately. Clock scripts will emit global pub/sub events, distinct
to each clock, on an interval. For example the 10-second clock would emit an
event named "clock:10s" every 10 seconds. Interested doodad scripts could
subscribe to that clock signal to run their own logic.

## Small Key Locked Doors

We already have color-coded Locked Doors (blue, red, yellow, green) where the
player needs to pick up the matching-color key, and then they can open ALL DOORS
of the matching color. (Colored keys are multiple-use items).

Add a new "Small Key Locked Door" which uses consumable keys: when the player
picks up a Small Key they can unlock one door with it, which consumes the key.
The player then needs another Small Key to unlock another door. The player can
carry multiple Small Keys at the same time.

## Movable Platform

Add a platform that the player can ride on that moves from one point to another.

This would come as two doodads:

* The platform itself.
* A dotted-line outline of the platform to indicate where the platform will move
  to when the level is played.

You will Link the platform to its destination outline to communicate which
destination is for which platform. In-game, the destination outline doodad will
be invisible and the platform(s) linked to it will move towards it, and then
move back to their original position, on a loop.

Implementation ideas:

* Add a concept of "rider/passenger" between doodads in a level.
* In the OnCollide() of the moving platform, if the player character is on top
  of the platform, call "SetPassenger(e.Actor.ID())" to mark the player as
  riding the platform.
* In OnLeave() call "RemovePassenger()" to un-mark the relationship.
* In the engine, if a moving actor has passengers, move the passengers along
  with the actor.

# Completed Doodads

## Start Flag (DONE)

To control the player spawn point in a level, a "Start Flag" could be dragged
into the level. On level startup, the first Spawn Flag is located and the player
is put there.

If no Start Flag is found in the level, the player spawns at coordinate 0,0 at
the top-left corner of the page.

If multiple Start Flags are found, consider it an error and notify the user. The
player would appear at one of the flags randomly in this case.

## Crumbly Floor (DONE)

A rectangular floor piece with lines indicating cracks. Most similar to:
the break-away floors in Tomb Raider.

Animation frames/states:

* Default: a rectangle with jagged lines through it indicating cracks.
* Rumble: draw little rumble mark lines and maybe shake the segments around.
* Break: the floor breaks apart and pieces fall/shrink into nothing over a few
  frames of animation.

Behavior:

* If touched, act as a solid object.
* If touched along its top edge, start the Rumble animation. If touched from
  the bottom, don't do anything, just act solid.
* After a moment of rumbling, stop acting solid and play the break animation.
  A player standing on top of the floor falls through it now.
* When the broken floor scrolls out of view it resets.

## Two-state Blocks (DONE)

This is an idea how to add a global two-state ON/OFF set of doodads for levels,
similar to the two-state blocks in _Mario Maker 2_ or the blue switches
on _Chip's Challenge._

Currently, we have switches that can toggle doors open and closed but the map
editor must link these together manually. The two-state doodads instead should
work globally, where the ON/OFF switch should toggle the state of ALL two-state
doodads without needing to be manually linked up.

## Pub/Sub Broadcast

To implement the global messaging without manually linking the doodads, add a
new method to the `Message` API for the Doodad scripts:

```javascript
Message.Broadcast("name", args...)
```

It will be like `Message.Publish()` but sends the message to ALL doodads whether
they're linked to the publisher or not.

On the recipient side, they will `Message.Subscribe("name", func)` as always.
