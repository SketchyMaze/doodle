// Crusher

/*
A.I. Behaviors:

- Sleeps and hangs in the air in a high place.
- When a player gets nearby, it begins "peeking" in their direction.
- When the player is below, tries to drop and crush them or any
  other mobile doodad.
- The top edge is safe to walk on and ride back up like an elevator.
*/

let direction = "left",
    dropSpeed = 12,
    riseSpeed = 4,
    watchRadius = 300, // player nearby distance to start peeking
    fallRadius = 120,   // player distance before it drops
    helmetThickness = 48, // safe solid hitbox height
    fireThickness = 12,   // dangerous bottom thickness
    targetAltitude = Self.Position()
    lastAltitude = targetAltitude.Y
    size = Self.Size();

const states = {
    idle: 0,
    peeking: 1,
    drop: 2,
    falling: 3,
    hit: 4,
    rising: 5,
};
let state = states.idle;

function main() {
	Self.SetMobile(true);
    Self.SetGravity(false);
    Self.SetInvulnerable(true);
	Self.SetHitbox(5, 2, 90, 73);
    Self.AddAnimation("hit", 50,
        ["angry", "ouch", "angry", "angry", "angry", "angry",
         "sleep", "sleep", "sleep", "sleep", "sleep", "sleep",
         "sleep", "sleep", "sleep", "sleep", "sleep", "sleep",
         "sleep", "sleep", "sleep", "sleep", "sleep", "sleep"],
    )

	// Player Character controls?
	if (Self.IsPlayer()) {
		return player();
	}

    let hitbox = Self.Hitbox();

	Events.OnCollide((e) => {
        // The bottom is deadly if falling.
        if (state === states.falling || state === states.hit && e.Settled) {
            if (e.Actor.IsMobile() && e.InHitbox && !e.Actor.Invulnerable()) {
                if (e.Overlap.H > 72) {
                    if (e.Actor.IsPlayer()) {
                        FailLevel("Don't get crushed!");
                        return;
                    } else {
                        e.Actor.Destroy();
                    }
                }
            }
        }

        // Our top edge is always solid.
        if (e.Actor.IsPlayer() && e.InHitbox) {
            if (e.Overlap.Y < helmetThickness) {
                // Be sure to position them snug on top.
                // TODO: this might be a nice general solution in the
                // collision detector...
                e.Actor.MoveTo(Point(
                    e.Actor.Position().X,
                    Self.Position().Y - e.Actor.Hitbox().Y - e.Actor.Hitbox().H,
                ))
                e.Actor.SetGrounded(true);
            }
        }

        // The whole hitbox is ordinarily solid.
        if (state !== state.falling) {
            if (e.Actor.IsMobile() && e.InHitbox) {
                return false;
            }
        }
	});

	setInterval(() => {
		// Find the player.
        let player = Actors.FindPlayer(),
            playerPoint = player.Position(),
            point = Self.Position(),
            delta = 0,
            nearby = false,
            below = false;

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
            nearby = true;
        }
        if (delta < fallRadius) {
            // Check if the player is below us.
            if (playerPoint.Y > point.Y + size.H) {
                below = true;
            }
        }

        switch (state) {
            case states.idle:
                if (nearby) {
                    Self.ShowLayerNamed("peek-"+direction);
                } else {
                    Self.ShowLayerNamed("sleep");
                }

                if (below) {
                    state = states.drop;
                } else if (nearby) {
                    state = states.peeking;
                }

                break;
            case states.peeking:
                if (nearby) {
                    Self.ShowLayerNamed("peek-"+direction);
                } else {
                    state = states.idle;
                    break;
                }

                if (below) {
                    state = states.drop;
                }

                break;
            case states.drop:
                // Begin the fall.
                Self.ShowLayerNamed("angry");
                Self.SetVelocity(Vector(0.0, dropSpeed));
                lastAltitude = -point.Y;
                state = states.falling;
            case states.falling:
                Self.ShowLayerNamed("angry");
                Self.SetVelocity(Vector(0.0, dropSpeed));

                // Landed?
                if (point.Y === lastAltitude) {
                    Sound.Play("crumbly-break.wav")
                    state = states.hit;
                    Self.PlayAnimation("hit", () => {
                        state = states.rising;
                    });
                }
                break;
            case states.hit:
                // A transitory state while the hit animation
                // plays out.
                break;
            case states.rising:
                Self.ShowLayerNamed("sleep");
                Self.SetVelocity(Vector(0, -riseSpeed));

                point = Self.Position();
                if (point.Y <= targetAltitude.Y+4 || point.Y === lastAltitude.Y) {
                    Self.MoveTo(targetAltitude);
                    Self.SetVelocity(Vector(0, 0))
                    state = states.idle;
                }
        }

        lastAltitude = point.Y;
	}, 100);
}

// If under control of the player character.
function player() {
	Events.OnKeypress((ev) => {
		if (ev.Right) {
			direction = "right";
		} else if (ev.Left) {
			direction = "left";
		}

        // Jump!
        if (ev.Down) {
            Self.ShowLayerNamed("angry");
            return;
        } else if (ev.Right && ev.Left) {
            Self.ShowLayerNamed("ouch");
        } else if (ev.Right || ev.Left) {
            Self.ShowLayerNamed("peek-"+direction);
        } else {
            Self.ShowLayerNamed("sleep");
        }
	});
}
