function main() {
	console.log("%s initialized!", Self.Doodad.Title);

	Self.AddAnimation("open", 100, [0, 1, 2, 3]);
	Self.AddAnimation("close", 100, [3, 2, 1, 0]);
	var animating = false;
	var opened = false;

	Events.OnCollide(function(e) {
		if (animating || opened) {
			return;
		}

		if (e.Overlap.X + e.Overlap.W >= 16 && e.Overlap.X < 48) {
			animating = true;
			Self.PlayAnimation("open", function() {
				opened = true;
				animating = false;
			});
		}
	});
	Events.OnLeave(function() {
		if (opened) {
			Self.PlayAnimation("close", function() {
				opened = false;
			});
		}
	})
}
