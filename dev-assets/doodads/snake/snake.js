// Snake

/*
A.I. Behaviors:

- Always turns to face the nearest player character
- Jumps up when the player tries to jump over them,
  aiming to attack the player.
*/

let direction = "left",
    jumpSpeed = 12,
    watchRadius = 300, // player nearby distance for snake to jump
    jumpCooldownStart = time.Now(),
    size = Self.Size();

const states = {
    idle: 0,
    attacking: 1,
};
let state = states.idle;

function main() {
	Self.SetMobile(true);
	Self.SetGravity(true);
	Self.SetHitbox(20, 0, 28, 58);
	Self.AddAnimation("idle-left", 100, ["left-1", "left-2", "left-3", "left-2"]);
	Self.AddAnimation("idle-right", 100, ["right-1", "right-2", "right-3", "right-2"]);
    Self.AddAnimation("attack-left", 100, ["attack-left-1", "attack-left-2", "attack-left-3"])
    Self.AddAnimation("attack-right", 100, ["attack-right-1", "attack-right-2", "attack-right-3"])

	// Player Character controls?
	if (Self.IsPlayer()) {
		return player();
	}

	Events.OnCollide((e) => {
		// The snake is deadly to the touch.
		if (e.Settled && e.Actor.IsPlayer() && e.InHitbox) {
            // Friendly to fellow snakes.
            if (e.Actor.Doodad().Filename.indexOf("snake") > -1) {
                return;
            }

			FailLevel("Watch out for snakes!");
			return;
		}
	});

	setInterval(() => {
		// Find the player.
        let player = Actors.FindPlayer(),
            playerPoint = player.Position(),
            point = Self.Position(),
            delta = 0,
            nearby = false;

        // Face the player.
        if (playerPoint.X < point.X + (size.W / 2)) {
            direction = "left";
            delta = Math.abs(playerPoint.X - (point.X + (size.W/2)));
        }
        else if (playerPoint.X > point.X + (size.W / 2)) {
            direction = "right";
            delta = Math.abs(playerPoint.X - (point.X + (size.W/2)));
        }

        if (delta < watchRadius) {
            console.log("Player is nearby snake! %d", delta);
            nearby = true;
        }

        // If we are idle and the player is jumping nearby...
        if (state == states.idle && nearby && Self.Grounded()) {
            if (playerPoint.Y - point.Y+(size.H/2) < 20) {
                console.warn("Player is jumping near us!")

                // Enter attack state.
                if (time.Since(jumpCooldownStart) > 500 * time.Millisecond) {
                    state = states.attacking;
                    Self.SetVelocity(Vector(0, -jumpSpeed));
                    Self.StopAnimation();
                    Self.PlayAnimation("attack-"+direction, null);
                    return;
                }
            }
        }

        // If we are attacking and gravity has claimed us back.
        if (state === states.attacking && Self.Grounded()) {
            console.log("Landed again after jump!");
            state = states.idle;
            jumpCooldownStart = time.Now();
            Self.StopAnimation();
        }

		// Ensure that the animations are always rolling.
        if (state === states.idle && !Self.IsAnimating()) {
            Self.PlayAnimation("idle-"+direction, null);
        }
	}, 100);
}

// If under control of the player character.
function player() {
	let jumping = false;

	Events.OnKeypress((ev) => {
		Vx = 0;
		Vy = 0;

		if (ev.Right) {
			direction = "right";
		} else if (ev.Left) {
			direction = "left";
		}

        // Jump!
        if (ev.Up && !jumping) {
            Self.StopAnimation();
            Self.PlayAnimation("attack-"+direction, null);
            jumping = true;
            return;
        }

		if (jumping && Self.Grounded()) {
            Self.StopAnimation();
            jumping = false;
        }

        if (!jumping && !Self.IsAnimating()) {
            Self.PlayAnimation("idle-"+direction, null);
        }
	});
}
