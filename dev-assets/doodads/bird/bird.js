// Bird

function main() {
	let speed = 4,
		Vx = Vy = 0,
		altitude = Self.Position().Y; // original height in the level

	let direction = "left",
		lastDirection = "left";
	let states = {
		flying: 0,
		diving: 1,
	};
	let state = states.flying;

	Self.SetMobile(true);
	Self.SetGravity(false);
	Self.SetHitbox(0, 0, 46, 32);
	Self.AddAnimation("fly-left", 100, ["left-1", "left-2"]);
	Self.AddAnimation("fly-right", 100, ["right-1", "right-2"]);

	// Player Character controls?
	if (Self.IsPlayer()) {
		return player();
	}

	Events.OnCollide((e) => {
		if (e.Actor.IsMobile() && e.InHitbox) {
			return false;
		}
	});

	// Sample our X position every few frames and detect if we've hit a solid wall.
	let sampleTick = 0,
		sampleRate = 2,
		lastSampledX = 0,
		lastSampledY = 0;

	setInterval(() => {
		if (sampleTick % sampleRate === 0) {
			let curX = Self.Position().X;
			let delta = Math.abs(curX - lastSampledX);
			if (delta < 5) {
				direction = direction === "right" ? "left" : "right";
			}
			lastSampledX = curX;
		}
		sampleTick++;

		// If we are not flying at our original altitude, correct for that.
		let curV = Self.Position();
		let Vy = 0.0;
		if (curV.Y != altitude) {
			Vy = curV.Y < altitude ? 1 : -1;
		}

		// TODO: Vector() requires floats, pain in the butt for JS,
		// the JS API should be friendlier and custom...
		let Vx = parseFloat(speed * (direction === "left" ? -1 : 1));
		Self.SetVelocity(Vector(Vx, Vy));

		// If we changed directions, stop animating now so we can
		// turn around quickly without moonwalking.
		if (direction !== lastDirection) {
			Self.StopAnimation();
		}

		if (!Self.IsAnimating()) {
			Self.PlayAnimation("fly-" + direction, null);
		}

		lastDirection = direction;
	}, 100);
}

// If under control of the player character.
function player() {
	Self.SetInventory(true);
	Events.OnKeypress((ev) => {
		Vx = 0;
		Vy = 0;

		if (ev.Up) {
			Vy = -playerSpeed;
		} else if (ev.Down) {
			Vy = playerSpeed;
		}

		if (ev.Right) {
			if (!Self.IsAnimating()) {
				Self.PlayAnimation("fly-right", null);
			}
			Vx = playerSpeed;
		} else if (ev.Left) {
			if (!Self.IsAnimating()) {
				Self.PlayAnimation("fly-left", null);
			}
			Vx = -playerSpeed;
		} else {
			Self.StopAnimation();
			animating = false;
		}

		Self.SetVelocity(Vector(Vx, Vy));
	})
}
