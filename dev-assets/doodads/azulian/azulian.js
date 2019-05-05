function main() {
	log.Info("Azulian '%s' initialized!", Self.Doodad.Title);

	var playerSpeed = 12;
	var gravity = 4;
	var Vx = Vy = 0;

	var animating = false;
	var animStart = animEnd = 0;
	var animFrame = animStart;

	setInterval(function() {
		if (animating) {
			if (animFrame < animStart || animFrame > animEnd) {
				animFrame = animStart;
			}

			animFrame++;
			if (animFrame === animEnd) {
				animFrame = animStart;
			}
			Self.ShowLayer(animFrame);
		} else {
			Self.ShowLayer(animStart);
		}
	}, 100);

	Events.OnKeypress(function(ev) {
		Vx = 0;
		Vy = 0;

		if (ev.Right.Now) {
			animStart = 2;
			animEnd = animStart+4;
			animating = true;
			Vx = playerSpeed;
		} else if (ev.Left.Now) {
			animStart = 6;
			animEnd = animStart+4;
			animating = true;
			Vx = -playerSpeed;
		} else {
			animating = false;
		}

		if (!Self.Grounded()) {
			Vy += gravity;
		}

		// Self.SetVelocity(Point(Vx, Vy));
	})
}
