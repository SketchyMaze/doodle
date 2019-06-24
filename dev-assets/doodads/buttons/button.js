function main() {
	console.log("%s initialized!", Self.Doodad.Title);

	var timer = 0;

	Events.OnCollide(function(e) {
		// Verify they've touched the button.
		if (e.Overlap.Y + e.Overlap.H < 24) {
			return;
		}

		Message.Publish("power", true);

		if (timer > 0) {
			clearTimeout(timer);
		}

		Self.ShowLayer(1);
		timer = setTimeout(function() {
			Self.ShowLayer(0);
			Message.Publish("power", false);
			timer = 0;
		}, 200);
	});

	// Events.OnLeave(function(e) {
	// 	console.log("%s has stopped touching %s", e, Self.Doodad.Title)
	// 	Self.Canvas.SetBackground(RGBA(0, 0, 1, 0));
	// })
}
