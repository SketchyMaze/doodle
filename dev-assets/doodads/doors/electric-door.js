function main() {
	console.log("%s initialized!", Self.Doodad.Title);

	var timer = 0;

	// Animation frames.
	var frame = 0;
	var frames = Self.LayerCount();
	var animationDirection = 1; // forward or backward
	var animationSpeed = 100;   // interval between frames when animating
	var animating      = false; // true if animation is actively happening

	console.warn("Electric Door has %d frames", frames);

	// Animation interval function.
	setInterval(function() {
		if (!animating) {
			return;
		}

		// Advance the frame forwards or backwards.
		frame += animationDirection;
		if (frame >= frames) {
			// Reached the last frame, start the pause and reverse direction.
			animating = false;
			frame = frames - 1;
		}

		Self.ShowLayer(frame);
	}, animationSpeed);

	Events.OnCollide( function() {
		animating = true; // start the animation
	})
}
