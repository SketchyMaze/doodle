// Azulian (Red and Blue)
var playerSpeed = 12,
	animating = false,
	direction = "right";

function setupAnimations(color) {
	var left = color === 'blue' ? 'blu-wl' : 'red-wl',
		right = color === 'blue' ? 'blu-wr' : 'red-wr',
		leftFrames = [left + '1', left + '2', left + '3', left + '4'],
		rightFrames = [right + '1', right + '2', right + '3', right + '4'];

	Self.AddAnimation("walk-left", 100, leftFrames);
	Self.AddAnimation("walk-right", 100, rightFrames);
}

function main() {
	var color = Self.GetTag("color");
	playerSpeed = color === 'blue' ? 2 : 4;

	Self.SetMobile(true);
	Self.SetGravity(true);
	Self.SetInventory(true);
	Self.SetHitbox(0, 0, 24, 32);
	setupAnimations(color);

	if (Self.IsPlayer()) {
		return playerControls();
	}

	// A.I. pattern: walks back and forth, turning around
	// when it meets resistance.

	// Sample our X position every few frames and detect if we've hit a solid wall.
	var sampleTick = 0;
	var sampleRate = 5;
	var lastSampledX = 0;

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

		var Vx = parseFloat(playerSpeed * (direction === "left" ? -1 : 1));
		Self.SetVelocity(Vector(Vx, 0.0));

		if (!Self.IsAnimating()) {
			Self.PlayAnimation("walk-" + direction, null);
		}
	}, 100);
}

function playerControls() {
	// Note: player speed is controlled by the engine.
	Events.OnKeypress(function (ev) {
		if (ev.Right) {
			if (!Self.IsAnimating()) {
				Self.PlayAnimation("walk-right", null);
			}
		} else if (ev.Left) {
			if (!Self.IsAnimating()) {
				Self.PlayAnimation("walk-left", null);
			}
		} else {
			Self.StopAnimation();
			animating = false;
		}
	})
}
