// Generic "Anvil" Doodad Script
/*
A doodad that falls and is dangerous while it falls.

Can be attached to any doodad.
*/

var falling = false;

function main() {
    // Make the hitbox be the full canvas size of this doodad.
    // Adjust if you want a narrower hitbox.
    var size = Self.Size()
    Self.SetHitbox(0, 0, size.W, size.H)

    // Note: doodad is not "solid" but hurts if it falls on you.
    Self.SetMobile(true);
    Self.SetGravity(true);

    // Monitor our Y position to tell if we've been falling.
    var lastPoint = Self.Position();
    setInterval(function () {
        var nowAt = Self.Position();
        if (nowAt.Y > lastPoint.Y) {
            falling = true;
        } else {
            falling = false;
        }
        lastPoint = nowAt;
    }, 100);

    Events.OnCollide(function (e) {
        if (!e.Settled) {
            return;
        }

        // Were we falling?
        if (falling) {
            if (e.InHitbox) {
                if (e.Actor.IsPlayer()) {
                    // Fatal to the player.
                    Sound.Play("crumbly-break.wav");
                    FailLevel("Watch out for " + Self.Title + "!");
                    return;
                }
                else if (e.Actor.IsMobile()) {
                    // Destroy mobile doodads.
                    Sound.Play("crumbly-break.wav");
                    e.Actor.Destroy();
                }
            }
        }
    });

    // When we receive power, we reset to our original position.
    var origPoint = Self.Position();
    Message.Subscribe("power", function (powered) {
        Self.MoveTo(origPoint);
        Self.SetVelocity(Vector(0, 0));
    });
}
