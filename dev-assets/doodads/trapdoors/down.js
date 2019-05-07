function main() {
	console.log("%s initialized!", Self.Doodad.Title);

	var timer = 0;

	var animationSpeed = 100;
	var opened = false;
	Self.AddAnimation("open", animationSpeed, ["down1", "down2", "down3", "down4"]);
	Self.AddAnimation("close", animationSpeed, ["down4", "down3", "down2", "down1"]);

	Events.OnCollide( function(e) {
		if (opened) {
			return;
		}

		// Not touching the top of the door means door doesn't open.
		if (e.Overlap.Y > 9) {
			return;
		}

		opened = true;
		Self.PlayAnimation("open", function() {
		});
	});
	Events.OnLeave(function() {
		Self.PlayAnimation("close", function() {
			opened = false;
		});
	})
}
