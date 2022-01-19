// Checkpoint Flag.
var isCurrentCheckpoint = false,
    playerEntered = false
    broadcastCooldown = time.Now();

function main() {
    Self.SetHitbox(22 + 16, 16, 75 - 16, 86);
    setActive(false);

    // If the checkpoint is linked to any doodad, the player character will
    // become that doodad when they cross this checkpoint.
    let skin = null;
    for (let actor of Self.GetLinks()) {
        skin = actor.Filename;
        actor.Destroy();
    }

    // Checkpoints broadcast to all of their peers so they all
    // know which one is the most recently activated.
    Message.Subscribe("broadcast:checkpoint", (currentID) => {
        setActive(false);
        return "a ok";
    });

    Events.OnCollide((e) => {
        if (!e.Settled) {
            return;
        }

        // Only care about the player character.
        if (!e.Actor.IsPlayer()) {
            return;
        }

        SetCheckpoint(Self.Position());
        setActive(true);

        // Don't spam the PubSub queue or we get races and deadlocks.
        if (time.Now().After(broadcastCooldown)) {
            Message.Broadcast("broadcast:checkpoint", Self.ID());
            broadcastCooldown = time.Now().Add(5 * time.Second)
        }

        // Are we setting a new player skin?
        if (skin && e.Actor.Doodad().Filename !== skin) {
            Actors.SetPlayerCharacter(skin);
        }
    });
}

function setActive(v) {
    if (v && !isCurrentCheckpoint) {
        Flash("Checkpoint!");
    }

    isCurrentCheckpoint = v;
    Self.ShowLayerNamed(v ? "checkpoint-active" : "checkpoint-inactive");
}
