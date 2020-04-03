
function main() {
	var color = Self.Doodad.Tag("color");

	// Layers in the doodad image.
	var layer = {
		closed: 0,
		right: 1,
		left: 2,
	};

	// Variables that change in event handler.
	var unlocked = false;  // Key has been used to unlock the door (one time).
	var opened = false;    // If door is currently showing its opened state.
	var enterSide = 0;     // Side of player entering the door, -1 or 1, left or right.

	Self.SetHitbox(23, 0, 23, 64);

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
				return;
			}

			var data = e.Actor.GetData("key:" + color);
			if (data === "") {
				// Door is locked.
				return false;
			}

			if (e.Settled) {
				unlocked = true;
				Self.ShowLayer(enterSide < 0 ? layer.right : layer.left);
			}
		}
	});
	Events.OnLeave(function(e) {
		Self.ShowLayer(layer.closed);

		// Reset collision state.
		opened = false;
		enterSide = 0;
	});
}
