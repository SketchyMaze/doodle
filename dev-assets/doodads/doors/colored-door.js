function main() {
	var color = Self.GetTag("color");
	var keyname = color === "small" ? "small-key.doodad" : "key-" + color + ".doodad";

	// Layers in the doodad image.
	var layer = {
		closed: 0,
		unlocked: 1,
		right: 2,
		left: 3,
	};

	// Variables that change in event handler.
	var unlocked = false;  // Key has been used to unlock the door (one time).
	var opened = false;    // If door is currently showing its opened state.
	var enterSide = 0;     // Side of player entering the door, -1 or 1, left or right.

	Self.SetHitbox(34, 0, 13, 76);

	Events.OnCollide(function(e) {
		// Record the side that this actor has touched us, in case the door
		// needs to open.
		if (enterSide === 0) {
			enterSide = e.Overlap.X > 0 ? 1 : -1;
		}

		if (opened) {
			return;
		}

		if (e.InHitbox) {
			if (unlocked) {
				Self.ShowLayer(enterSide < 0 ? layer.right : layer.left);
				opened = true;
				Sound.Play("door-open.wav")
				return;
			}

			// Do they have our key?
			var hasKey = e.Actor.HasItem(keyname) >= 0;
			if (!hasKey) {
				// Door is locked.
				return false;
			}

			if (e.Settled) {
				unlocked = true;
				Self.ShowLayer(enterSide < 0 ? layer.right : layer.left);
				Sound.Play("unlock.wav");

				// If a Small Key door, consume a small key.
				if (color === "small") {
					e.Actor.RemoveItem(keyname, 1)
				}
			}
		}
	});
	Events.OnLeave(function(e) {
		Self.ShowLayer(unlocked ? layer.unlocked : layer.closed);
		// Sound.Play("door-close.wav")

		// Reset collision state.
		opened = false;
		enterSide = 0;
	});
}
