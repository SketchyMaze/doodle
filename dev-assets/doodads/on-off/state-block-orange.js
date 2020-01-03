// Orange State Block
function main() {
	Self.SetHitbox(0, 0, 33, 33);

	// Orange block is OFF by default.
	var state = false;

	Message.Subscribe("broadcast:state-change", function(newState) {
		state = newState;

		// Layer 0: OFF
		// Layer 1: ON
		Self.ShowLayer(state ? 1 : 0);
	});

	Events.OnCollide(function(e) {
		if (e.Actor.IsMobile() && e.InHitbox) {
			if (state) {
				return false;
			}
		}
	});
}
