function main() {
	console.log("%s initialized!", Self.Title);

	var pressed = false;

	// When a sticky button receives power, it pops back up.
	Message.Subscribe("power", function(powered) {
		if (powered && pressed) {
			Self.ShowLayer(0);
			pressed = false;
			Sound.Play("button-up.wav")
			Message.Publish("power", false);
		}
	})

	Events.OnCollide(function(e) {
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
	});
}
