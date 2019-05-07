function main() {
	console.log("%s initialized!", Self.Doodad.Title);

	var pressed = false;

	Events.OnCollide(function(e) {
		if (pressed) {
			return;
		}

		// Verify they've touched the button.
		if (e.Overlap.Y + e.Overlap.H < 24) {
			return;
		}

		Self.ShowLayer(1);
		pressed = true;
	});
}
