// Orange State Block
function main() {
	Self.SetHitbox(0, 0, 33, 33);

	// Orange block is OFF by default.
	var state = false;

	Message.Subscribe("broadcast:state-change", function(newState) {
		state = newState;
		console.warn("ORANGE BLOCK Received state=%+v, set mine to %+v", newState, state);

		// Layer 0: OFF
		// Layer 1: ON
		if (state) {
			Self.ShowLayer(1);
		} else {
			Self.ShowLayer(0);
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
