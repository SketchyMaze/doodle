function main() {
	console.log("%s initialized!", Self.Doodad.Title);

	var timer = 0;

	Self.SetHitbox(0, 0, 72, 9);

	var animationSpeed = 100;
	var opened = false;
	Self.AddAnimation("open", animationSpeed, ["down1", "down2", "down3", "down4"]);
	Self.AddAnimation("close", animationSpeed, ["down4", "down3", "down2", "down1"]);

	Events.OnCollide( function(e) {
		if (opened) {
			return;
		}

		// Is the actor colliding our solid part?
		if (e.InHitbox) {
			// Touching the top or the bottom?
			if (e.Overlap.Y > 0) {
				return false; // solid wall when touched from below
			} else {
				opened = true;
				Self.PlayAnimation("open", function() {
				});
			}
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
