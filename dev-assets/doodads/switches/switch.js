function main() {
	console.log("%s initialized!", Self.Doodad.Title);

	// Switch has two frames:
	// 0: Off
	// 1: On

	var state = false;
	var collide = false;

	Message.Subscribe("power", function(powered) {
		state = powered;
		showState(state);
	});

	Events.OnCollide(function(e) {
		if (collide === false) {
			state = !state;
			Message.Publish("power", state);
			showState(state);

			collide = true;
		}
	});

	Events.OnLeave(function(e) {
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
