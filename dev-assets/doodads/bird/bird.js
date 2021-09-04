// Bird

function main() {
	var speed = 4;
	var Vx = Vy = 0;
	var altitude = Self.Position().Y; // original height in the level

	var direction = "left",
		lastDirection = "left";
	var states = {
		flying: 0,
		diving: 1,
	};
	var state = states.flying;

	Self.SetMobile(true);
	Self.SetGravity(false);
	Self.SetHitbox(0, 0, 46, 32);
	Self.AddAnimation("fly-left", 100, ["left-1", "left-2"]);
	Self.AddAnimation("fly-right", 100, ["right-1", "right-2"]);

	// Player Character controls?
	if (Self.IsPlayer()) {
		return player();
	}

	Events.OnCollide(function (e) {
		if (e.Actor.IsMobile() && e.InHitbox) {
			return false;
		}
	});

	// Sample our X position every few frames and detect if we've hit a solid wall.
	var sampleTick = 0;
	var sampleRate = 2;
	var lastSampledX = 0;
	var lastSampledY = 0;

	setInterval(function () {
		if (sampleTick % sampleRate === 0) {
			var curX = Self.Position().X;
			var delta = Math.abs(curX - lastSampledX);
			if (delta < 5) {
				direction = direction === "right" ? "left" : "right";
			}
			lastSampledX = curX;
		}
		sampleTick++;

		// TODO: Vector() requires floats, pain in the butt for JS,
		// the JS API should be friendlier and custom...
		var Vx = parseFloat(speed * (direction === "left" ? -1 : 1));
		Self.SetVelocity(Vector(Vx, 0.0));

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
	Events.OnKeypress(function (ev) {
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