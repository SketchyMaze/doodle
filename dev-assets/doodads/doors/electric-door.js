function main() {
	console.log("%s initialized!", Self.Doodad.Title);

	var err = Self.AddAnimation("open", 100, [0, 1, 2, 3]);
	console.error("door error: %s", err)
	var animating = false;

	Events.OnCollide(function() {
		if (animating) {
			return;
		}

		animating = true;
		Self.PlayAnimation("open", null);
	});
}
