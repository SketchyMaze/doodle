function main() {
	let timer = 0;
	let pressed = false;

	// Has a linked Sticky Button been pressed permanently down?
	let stickyDown = false;
	Message.Subscribe("sticky:down", (down) => {
		stickyDown = down;
		Self.ShowLayer(stickyDown ? 1 : 0);
	});

	// Track who all is colliding with us.
	let colliders = {};

	Events.OnCollide((e) => {
		if (!e.Settled) {
			return;
		}

		if (colliders[e.Actor.ID()] == undefined) {
			colliders[e.Actor.ID()] = true;
		}

		// If a linked Sticky Button is pressed, button stays down too and
		// doesn't interact.
		if (stickyDown) {
			return;
		}

		// Verify they've touched the button.
		if (e.Overlap.Y + e.Overlap.H < 24) {
			return;
		}

		if (!pressed && !stickyDown) {
			Sound.Play("button-down.wav")
			Message.Publish("power", true);
			pressed = true;
		}


		if (timer > 0) {
			clearTimeout(timer);
		}

		Self.ShowLayer(1);
	});

	Events.OnLeave((e) => {
		delete colliders[e.Actor.ID()];

		if (Object.keys(colliders).length === 0 && !stickyDown) {
			Sound.Play("button-up.wav")
			Self.ShowLayer(0);
			Message.Publish("power", false);
			timer = 0;
			pressed = false;
		}
	});
}
