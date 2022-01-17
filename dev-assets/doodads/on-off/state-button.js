// State Block Control Button

// Button is "OFF" by default.
let state = false;

function main() {
	Self.SetHitbox(0, 0, 42, 42);

	// When the button is activated, don't keep toggling state until we're not
	// being touched again.
	let colliding = false;

	// If we receive a state change event from a DIFFERENT on/off button, update
	// ourself to match the state received.
	Message.Subscribe("broadcast:state-change", (value) => {
		state = value;
		showSprite();
	});

	Events.OnCollide((e) => {
		if (colliding) {
			return false;
		}

		// Only trigger for mobile characters.
		if (e.Actor.IsMobile()) {
			// Only activate if touched from the bottom or sides.
			if (e.Overlap.Y === 0) {
				return false;
			}

			colliding = true;
			state = !state;
			Message.Broadcast("broadcast:state-change", state);

			showSprite();
		}

		// Always a solid button.
		return false;
	});

	Events.OnLeave((e) => {
		colliding = false;
	})
}

// Update the active layer based on the current button state.
function showSprite() {
	if (state) {
		Self.ShowLayer(1);
	} else {
		Self.ShowLayer(0);
	}
}
