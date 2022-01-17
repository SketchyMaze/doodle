// Checkpoint Flag.
var isCurrentCheckpoint = false;

function main() {
    Self.SetHitbox(22 + 16, 16, 75 - 16, 86);
    setActive(false);

    // Checkpoints broadcast to all of their peers so they all
    // know which one is the most recently activated.
    Message.Subscribe("broadcast:checkpoint", (currentID) => {
        setActive(false);
    });

    Events.OnCollide((e) => {
        if (!e.Settled) {
            return;
        }

        // Only care about the player character.
        if (!e.Actor.IsPlayer()) {
            return;
        }

        // Set the player checkpoint.
        SetCheckpoint(Self.Position());
        setActive(true);
        Message.Broadcast("broadcast:checkpoint", Self.ID())
    });
}

function setActive(v) {
    if (v && !isCurrentCheckpoint) {
        Flash("Checkpoint!");
    }

    isCurrentCheckpoint = v;
    Self.ShowLayerNamed(v ? "checkpoint-active" : "checkpoint-inactive");
}
