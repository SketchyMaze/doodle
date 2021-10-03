// Stall Player.
// Tags: ms (int)
// Grabs the player one time. Resets if it receives power.
function main() {
    Self.Hide();

    var active = true,
        timeout = 250,
        ms = Self.GetTag("ms");

    if (ms.length > 0) {
        timeout = parseInt(ms);
    }

    Events.OnCollide(function (e) {
        if (!active || !e.Settled) {
            return;
        }

        // Only care if it's the player.
        if (!e.Actor.IsPlayer()) {
            return;
        }

        if (e.InHitbox) {
            // Grab hold of the player.
            e.Actor.Freeze();
            setTimeout(function () {
                e.Actor.Unfreeze();
            }, timeout);

            active = false;
        }
    });

    // Reset the trap if powered by a button.
    Message.Subscribe("power", function (powered) {
        active = true;
    });
}
