// Blue State Block
function main() {
	Self.SetHitbox(0, 0, 33, 33);

	// Blue block is ON by default.
	var state = true;

	Message.Subscribe("broadcast:state-change", function(newState) {
		state = !newState;

		// Layer 0: ON
		// Layer 1: OFF
		Self.ShowLayer(state ? 0 : 1);
	});

	Events.OnCollide(function(e) {
		if (e.Actor.IsMobile() && e.InHitbox) {
			if (state) {
				return false;
			}
		}
	});
}
