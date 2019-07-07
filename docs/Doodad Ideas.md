# Doodad Ideas and Implementation Notes

## Crumbly Floor

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
