function main() {
	console.log("%s initialized!", Self.Title);

	Self.AddAnimation("open", 100, [0, 1, 2, 3]);
	Self.AddAnimation("close", 100, [3, 2, 1, 0]);
	var animating = false;
	var opened = false;

	Self.SetHitbox(0, 0, 34, 76);

	Message.Subscribe("power", function(powered) {
		console.log("%s got power=%+v", Self.Title, powered);

		if (powered) {
			if (animating || opened) {
				return;
			}

			animating = true;
			Sound.Play("electric-door.wav")
			Self.PlayAnimation("open", function() {
				opened = true;
				animating = false;
			});
		} else {
			animating = true;
			Sound.Play("electric-door.wav")
			Self.PlayAnimation("close", function() {
				opened = false;
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
}
