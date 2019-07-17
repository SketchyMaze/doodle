function main() {
	Self.AddAnimation("open", 0, [1]);
	var unlocked = false;

	// Map our door names to key names.
	var KeyMap = {
		"Blue Door": "Blue Key",
		"Red Door": "Red Key",
		"Green Door": "Green Key",
		"Yellow Door": "Yellow Key"
	}

	Self.SetHitbox(16, 0, 32, 64);

	Events.OnCollide(function(e) {
		if (unlocked) {
			return;
		}

		if (e.InHitbox) {
			var data = e.Actor.GetData("key:" + KeyMap[Self.Doodad.Title]);
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
