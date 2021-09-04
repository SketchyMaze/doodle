// Generic "Fire" Doodad Script
/*
The entire square shape of your doodad acts similar to "Fire"
pixels - killing the player character upon contact.

Can be attached to any doodad.
*/

function main() {
    // Make the hitbox be the full canvas size of this doodad.
    // Adjust if you want a narrower hitbox.
    if (Self.Hitbox().IsZero()) {
        var size = Self.Size()
        Self.SetHitbox(0, 0, size.W, size.H)
    }

    Events.OnCollide(function (e) {
        if (!e.Settled || !e.InHitbox) {
            return;
        }

        // Turn mobile actors black, like real fire does.
        if (e.Actor.IsMobile()) {
            e.Actor.Canvas.MaskColor = RGBA(1, 1, 1, 255)
        }

        // End the level if it's the player.
        if (e.Actor.IsPlayer()) {
            FailLevel("Watch out for " + Self.Title + "!");
        }
    })

    Events.OnLeave(function (e) {
        if (e.Actor.IsMobile()) {
            e.Actor.MaskColor = RGBA(0, 0, 0, 0)
        }
    })
}