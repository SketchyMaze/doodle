function main() {
	var timer = 0;
	var pressed = false;

	// Has a linked Sticky Button been pressed permanently down?
	var stickyDown = false;
	Message.Subscribe("sticky:down", function(down) {
		stickyDown = down;
		Self.ShowLayer(stickyDown ? 1 : 0);
	});

	Events.OnCollide(function(e) {
		if (!e.Settled) {
			return;
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
		timer = setTimeout(function() {
			Sound.Play("button-up.wav")
			Self.ShowLayer(0);
			Message.Publish("power", false);
			timer = 0;
			pressed = false;
		}, 200);
	});
}
