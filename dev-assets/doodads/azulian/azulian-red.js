function main() {
	var playerSpeed = 4;
	var gravity = 4;
	var Vx = Vy = 0;

	var direction = "right";

	Self.SetHitbox(0, 0, 32, 32)
	Self.SetMobile(true);
	Self.SetGravity(true);
	Self.AddAnimation("walk-left", 100, ["red-wl1", "red-wl2", "red-wl3", "red-wl4"]);
	Self.AddAnimation("walk-right", 100, ["red-wr1", "red-wr2", "red-wr3", "red-wr4"]);

	// Sample our X position every few frames and detect if we've hit a solid wall.
	var sampleTick = 0;
	var sampleRate = 5;
	var lastSampledX = 0;

	setInterval(function() {
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
		var Vx = parseFloat(playerSpeed * (direction === "left" ? -1 : 1));
		Self.SetVelocity(Vector(Vx, 0.0));

		if (!Self.IsAnimating()) {
			Self.PlayAnimation("walk-"+direction, null);
		}
	}, 100);
}
