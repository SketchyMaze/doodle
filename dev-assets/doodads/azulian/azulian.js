function main() {
	var playerSpeed = 12;
	var gravity = 4;
	var Vx = Vy = 0;

	var animating = false;
	var animStart = animEnd = 0;
	var animFrame = animStart;

	Self.SetGravity(true);
	Self.SetHitbox(7, 4, 17, 28);
	Self.AddAnimation("walk-left", 100, ["blu-wl1", "blu-wl2", "blu-wl3", "blu-wl4"]);
	Self.AddAnimation("walk-right", 100, ["blu-wr1", "blu-wr2", "blu-wr3", "blu-wr4"]);

	Events.OnKeypress(function(ev) {
		Vx = 0;
		Vy = 0;

		if (ev.Right.Now) {
			if (!Self.IsAnimating()) {
				Self.PlayAnimation("walk-right", null);
			}
			Vx = playerSpeed;
		} else if (ev.Left.Now) {
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
