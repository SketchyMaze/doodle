function main() {
	console.log("%s initialized!", Self.Doodad.Title);

	var timer = 0;

	var animationSpeed = 100;
	var animating = false;
	Self.AddAnimation("open", animationSpeed, ["down1", "down2", "down3", "down4"]);
	Self.AddAnimation("close", animationSpeed, ["down4", "down3", "down2", "down1"]);

	Events.OnCollide( function() {
		if (animating) {
			return;
		}

		animating = true;
		Self.PlayAnimation("open", function() {
			setTimeout(function() {
				Self.PlayAnimation("close", function() {
					animating = false;
				});
			}, 3000);
		})
	});
}
