// Generic "Solid" Doodad Script
/*
The entire square shape of your doodad acts similar to "solid"
pixels - blocking collision from all sides.

Can be attached to any doodad.
*/

function main() {
    // Make the hitbox be the full canvas size of this doodad.
    // Adjust if you want a narrower hitbox.
    var size = Self.Size()
    Self.SetHitbox(0, 0, size.W, size.H)

    // Solid to all collisions.
    Events.OnCollide(function (e) {
        return false;
    })
}