function main() {
	console.log("%s initialized!", Self.Doodad.Title);

	Self.AddAnimation("open", 100, [0, 1, 2, 3]);
	Self.AddAnimation("close", 100, [3, 2, 1, 0]);
	var animating = false;
	var opened = false;

	Self.SetHitbox(16, 0, 32, 64);

	Message.Subscribe("power", function(powered) {
		console.log("%s got power=%+v", Self.Doodad.Title, powered);

		if (powered) {
			if (animating || opened) {
				return;
			}

			animating = true;
			Self.PlayAnimation("open", function() {
				opened = true;
				animating = false;
			});
		} else {
			animating = true;
			opened = false;
			Self.PlayAnimation("close", function() {
				animating = false;
			})
		}
	});

	Events.OnCollide(function(e) {
		if (e.InHitbox) {
			if (!opened) {
				return false;
			}
		}
	});
	Events.OnLeave(function() {
		// if (opened) {
		// 	Self.PlayAnimation("close", function() {
		// 		opened = false;
		// 	});
		// }
	})
}
