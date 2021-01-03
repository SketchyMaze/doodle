// Red bird mob.
function main() {
	var speed = 4;
	var Vx = Vy = 0;
	var altitude = Self.Position().Y; // original height in the level

	var direction = "left";
	var states = {
		flying: 0,
		diving: 1,
	};
	var state = states.flying;

	Self.SetMobile(true);
	Self.SetGravity(false);
	Self.SetHitbox(0, 10, 46, 32);
	Self.AddAnimation("fly-left", 100, ["left-1", "left-2"]);
	Self.AddAnimation("fly-right", 100, ["right-1", "right-2"]);

	Events.OnCollide(function(e) {
		if (e.Actor.IsMobile() && e.InHitbox) {
			return false;
		}
	});

	// Sample our X position every few frames and detect if we've hit a solid wall.
	var sampleTick = 0;
	var sampleRate = 2;
	var lastSampledX = 0;
	var lastSampledY = 0;

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
		var Vx = parseFloat(speed * (direction === "left" ? -1 : 1));
		Self.SetVelocity(Vector(Vx, 0.0));

		if (!Self.IsAnimating()) {
			Self.PlayAnimation("fly-"+direction, null);
		}
	}, 100);
}
