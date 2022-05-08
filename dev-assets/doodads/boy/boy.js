const playerSpeed = 12;

let Vx = Vy = 0,
	walking = false,
	direction = "right",
	lastDirection = direction;

function main() {
	Self.SetMobile(true);
	Self.SetInventory(true);
	Self.SetGravity(true);
	Self.SetHitbox(0, 0, 32, 52);
	Self.AddAnimation("walk-left", 200, ["stand-left", "walk-left-1", "walk-left-2", "walk-left-3", "walk-left-2", "walk-left-1"]);
	Self.AddAnimation("walk-right", 200, ["stand-right", "walk-right-1", "walk-right-2", "walk-right-3", "walk-right-2", "walk-right-1"]);
	Self.AddAnimation("idle-left", 200, ["idle-left-1", "idle-left-2", "idle-left-3", "idle-left-2"]);
	Self.AddAnimation("idle-right", 200, ["idle-right-1", "idle-right-2", "idle-right-3", "idle-right-2"]);

	// If the player suddenly changes direction, reset the animation state to quickly switch over.
	let lastVelocity = Vector(0, 0);

	Events.OnKeypress((ev) => {
		Vx = 0;
		Vy = 0;

		let curVelocity = Self.GetVelocity();
		if ((lastVelocity.X < 0 && curVelocity.X > 0) ||
			(lastVelocity.X > 0 && curVelocity.X < 0)) {
			Self.StopAnimation();
		}
		lastVelocity = curVelocity;
		lastDirection = direction;

		let wasWalking = walking;
		if (ev.Right) {
			direction = "right";
			Vx = playerSpeed;
			walking = true;
		} else if (ev.Left) {
			direction = "left";
			Vx = -playerSpeed;
			walking = true;
		} else {
			// Has stopped walking!
			walking = false;
			stoppedWalking = true;
		}

		// Should we stop animating? (changed state)
		if (direction !== lastDirection || wasWalking !== walking) {
			Self.StopAnimation();
		}

		// And play what animation?
		if (!Self.IsAnimating()) {
			if (walking) {
				Self.PlayAnimation("walk-"+direction, null);
			} else {
				Self.PlayAnimation("idle-"+direction, null);
			}
		}
	})
}
