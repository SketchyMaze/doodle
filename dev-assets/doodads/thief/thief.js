// Thief

function main() {
    Self.SetMobile(true);
    Self.SetGravity(true);
    Self.SetInventory(true);
    Self.SetHitbox(0, 0, 32, 58);
    Self.AddAnimation("walk-left", 200, ["stand-left"]); //, "walk-left-1", "walk-left-2", "walk-left-3", "walk-left-2", "walk-left-1"]);
    Self.AddAnimation("walk-right", 200, ["stand-right"]); //, "walk-right-1", "walk-right-2", "walk-right-3", "walk-right-2", "walk-right-1"]);

    // All thieves can steal items.
    stealable();

    // Controlled by the player character?
    if (Self.IsPlayer()) {
        return playable();
    }
    return ai();
}

// Common "steal" power between playable and A.I. thieves.
function stealable() {
    // Steals your items.
    Events.OnCollide(function (e) {
        var victim = e.Actor;
        if (!e.Settled) {
            return;
        }

        // Thieves don't steal from Thieves (unless controlled by the player).
        if (!Self.IsPlayer() && victim.Drawing.Doodad.Filename === "thief.doodad") {
            return;
        }

        // Steal inventory
        var stolen = 0;
        if (victim.HasInventory()) {
            var myInventory = Self.Inventory(),
                theirInventory = victim.Inventory();

            for (var key in theirInventory) {
                if (!theirInventory.hasOwnProperty(key)) {
                    continue;
                }

                var value = theirInventory[key];
                if (value > 0 || myInventory[key] === undefined) {
                    victim.RemoveItem(key, value);
                    Self.AddItem(key, value);
                    stolen += (value === 0 ? 1 : value);
                }
            }

            // If the player lost their items, notify them.
            if (victim.IsPlayer() && stolen > 0) {
                Flash("Watch out for thieves! %d item%s stolen!", parseInt(stolen), stolen === 1 ? ' was' : 's were');
            }

            // If the Thief IS the player, notify your earnings.
            if (Self.IsPlayer() && stolen > 0) {
                Flash("Awesome! Stole %d item%s from the %s!", parseInt(stolen), stolen === 1 ? '' : 's', e.Actor.Drawing.Doodad.Title);
            }
        }
    });
}

// Enemy Doodad AI: walks back and forth, changing direction
// when it encounters and obstacle.
function ai() {
    // Walks back and forth.
    var Vx = Vy = 0.0,
        playerSpeed = 4,
        direction = "right",
        lastSampledX = 0,
        sampleTick = 0,
        sampleRate = 2;

    setInterval(function () {
        if (sampleTick % sampleRate === 0) {
            var curX = Self.Position().X,
                delta = Math.abs(curX - lastSampledX);
            if (delta < 5) {
                direction = direction === "right" ? "left" : "right";
            }
            lastSampledX = curX;
        }
        sampleTick++;

        Vx = parseFloat(playerSpeed * (direction === "left" ? -1 : 1));
        Self.SetVelocity(Vector(Vx, Vy));

        Self.StopAnimation();
        Self.PlayAnimation("walk-" + direction, null);
    }, 100);
}

// If under control of the player character.
function playable() {
    Events.OnKeypress(function (ev) {
        Vx = 0;
        Vy = 0;

        if (ev.Right) {
            if (!Self.IsAnimating()) {
                Self.PlayAnimation("walk-right", null);
            }
            Vx = playerSpeed;
        } else if (ev.Left) {
            if (!Self.IsAnimating()) {
                Self.PlayAnimation("walk-left", null);
            }
            Vx = -playerSpeed;
        } else {
            Self.StopAnimation();
            animating = false;
        }

        // Self.SetVelocity(Point(Vx, Vy));
    })
}