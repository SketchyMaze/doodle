// Blue State Block
function main() {
	Self.SetHitbox(0, 0, 33, 33);

	// Blue block is ON by default.
	var state = true;

	Message.Subscribe("broadcast:state-change", function(newState) {
		state = !newState;
		console.warn("BLUE BLOCK Received state=%+v, set mine to %+v", newState, state);

		// Layer 0: ON
		// Layer 1: OFF
		if (state) {
			Self.ShowLayer(0);
		} else {
			Self.ShowLayer(1);
		}
	});

	Events.OnCollide(function(e) {
		if (e.Actor.IsMobile() && e.InHitbox) {
			if (state) {
				return false;
			}
		}
	});
}
