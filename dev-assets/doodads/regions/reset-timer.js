// Reset Level Timer.
function main() {
    Self.Hide();

    // Reset the level timer only once.
    let hasReset = false;

    Events.OnCollide((e) => {
        if (!e.Settled) {
            return;
        }

        // Only care if it's the player.
        if (!e.Actor.IsPlayer()) {
            return;
        }

        if (e.InHitbox && !hasReset) {
            Level.ResetTimer();
            hasReset = true;
        }
    });

    // Receive a power signal resets the doodad.
    Message.Subscribe("power", (powered) => {
        if (powered) {
            hasReset = true;
        }
    });
}
