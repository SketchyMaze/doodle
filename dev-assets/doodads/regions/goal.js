// Goal Region.
function main() {
    Self.Hide();

    Events.OnCollide(function (e) {
        if (!e.Settled) {
            return;
        }

        // Only care if it's the player.
        if (!e.Actor.IsPlayer()) {
            return;
        }

        if (e.InHitbox) {
            EndLevel();
        }
    });
}
