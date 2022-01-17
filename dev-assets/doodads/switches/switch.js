function main() {

	// Switch has two frames:
	// 0: Off
	// 1: On

	let state = false;
	let collide = false;

	Message.Subscribe("power", (powered) => {
		state = powered;
		showState(state);
	});

	Events.OnCollide((e) => {
		if (!e.Settled || !e.Actor.IsMobile()) {
			return;
		}

		if (collide === false) {
			Sound.Play("button-down.wav")
			state = !state;

			Message.Publish("switch:toggle", state);
			Message.Publish("power", state);
			showState(state);

			collide = true;
		}
	});

	Events.OnLeave((e) => {
		collide = false;
	});
}

// showState shows the on/off frame based on the boolean powered state.
function showState(state) {
	if (state) {
		Self.ShowLayer(1);
	} else {
		Self.ShowLayer(0);
	}
}
