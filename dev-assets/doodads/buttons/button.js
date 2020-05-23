function main() {
	console.log("%s initialized!", Self.Title);

	var timer = 0;
	var pressed = false;

	Events.OnCollide(function(e) {
		if (!e.Settled) {
			return;
		}

		// Verify they've touched the button.
		if (e.Overlap.Y + e.Overlap.H < 24) {
			return;
		}

		if (!pressed) {
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
