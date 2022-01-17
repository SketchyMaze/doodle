function main() {
	let pressed = false;

	// When a sticky button receives power, it pops back up.
	Message.Subscribe("power", (powered) => {
		if (powered && pressed) {
			Self.ShowLayer(0);
			pressed = false;
			Sound.Play("button-up.wav")
			Message.Publish("power", false);
			Message.Publish("sticky:down", false);
		}
	})

	Events.OnCollide((e) => {
		if (!e.Settled) {
			return;
		}

		if (pressed) {
			return;
		}

		// Verify they've touched the button.
		if (e.Overlap.Y + e.Overlap.H < 24) {
			return;
		}

		Sound.Play("button-down.wav")
		Self.ShowLayer(1);
		pressed = true;
		Message.Publish("power", true);
		Message.Publish("sticky:down", true);
	});
}
