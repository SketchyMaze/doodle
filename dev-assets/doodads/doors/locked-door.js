// DEPRECATED: old locked door script. Superceded by colored-door.js.
function main() {
	Self.AddAnimation("open", 0, [1]);
	var unlocked = false;
	var color = Self.Doodad().Tag("color");

	Self.SetHitbox(16, 0, 32, 64);

	Events.OnCollide(function(e) {
		if (unlocked) {
			return;
		}

		if (e.InHitbox) {
			var data = e.Actor.GetData("key:" + color);
			if (data === "") {
				// Door is locked.
				return false;
			}

			if (e.Settled) {
				unlocked = true;
				Self.PlayAnimation("open", null);
			}
		}
	});
}
