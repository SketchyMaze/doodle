// Checkpoint Region
// Acts like an invisible checkpoint flag.
var isCurrentCheckpoint = false;

function main() {
    Self.Hide();
    setActive(false);

    // Checkpoints broadcast to all of their peers so they all
    // know which one is the most recently activated.
    Message.Subscribe("broadcast:checkpoint", function (currentID) {
        setActive(false);
    });

    Events.OnCollide(function (e) {
        if (isCurrentCheckpoint || !e.Settled) {
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
}