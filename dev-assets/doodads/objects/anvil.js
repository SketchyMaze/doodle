// Anvil
var falling = false;

function main() {
    // Note: doodad is not "solid" but hurts if it falls on you.
    Self.SetHitbox(0, 0, 48, 25);
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
                    FailLevel("Watch out for anvils!");
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
