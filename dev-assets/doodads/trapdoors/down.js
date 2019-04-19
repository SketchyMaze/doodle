function main() {
	console.log("%s initialized!", Self.Doodad.Title);

	var timer = 0;

	// Animation frames.
	var frame = 0;
	var frames = Self.LayerCount();
	var animationDirection = 1; // forward or backward
	var animationSpeed = 100;   // interval between frames when animating
	var animationDelay = 8;    // delay ticks at the end before reversing, in
	                            // multiples of animationSpeed
	var delayCountdown = 0;
	var animating      = false; // true if animation is actively happening

	console.warn("Trapdoor has %d frames", frames);

	// Animation interval function.
	setInterval(function() {
		if (!animating) {
			return;
		}

		// At the end of the animation (door is open), delay before resuming
		// the close animation.
		if (delayCountdown > 0) {
			delayCountdown--;
			return;
		}

		// Advance the frame forwards or backwards.
		frame += animationDirection;
		if (frame >= frames) {
			// Reached the last frame, start the pause and reverse direction.
			delayCountdown = animationDelay;
			animationDirection = -1;

			// also bounds check it
			frame = frames - 1;
		}

		if (frame < 0) {
			// reached the start again
			frame = 0;
			animationDirection = 1;
			animating = false;
		}

		Self.ShowLayer(frame);
	}, animationSpeed);

	Events.OnCollide( function() {
		animating = true; // start the animation
	})
}
