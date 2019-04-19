function main() {
	console.log("%s initialized!", Self.Doodad.Title);

	var timer = 0;

	Events.OnCollide( function() {
		if (timer > 0) {
			clearTimeout(timer);
		}

		Self.ShowLayer(1);
		timer = setTimeout(function() {
			Self.ShowLayer(0);
			timer = 0;
		}, 200);
	})
}
