// State Block Control Button
function main() {
	console.log("%s initialized!", Self.Doodad.Title);
	Self.SetHitbox(0, 0, 33, 33);

	// When the button is activated, don't keep toggling state until we're not
	// being touched again.
	var colliding = false;

	// Button is "OFF" by default.
	var state = false;

	Events.OnCollide(function(e) {
		if (colliding) {
			return false;
		}

		// Only trigger for mobile characters.
		if (e.Actor.IsMobile()) {
			console.log("Mobile actor %s touched the on/off button!", e.Actor.Actor.Filename);

			// Only activate if touched from the bottom or sides.
			if (e.Overlap.Y === 0) {
				console.log("... but touched the top!");
				return false;
			}

			colliding = true;
			console.log("   -> emit state change");
			state = !state;
			Message.Broadcast("broadcast:state-change", state);

			if (state) {
				Self.ShowLayer(1);
			} else {
				Self.ShowLayer(0);
			}
		}

		// Always a solid button.
		return false;
	});

	Events.OnLeave(function(e) {
		colliding = false;
	})
}
