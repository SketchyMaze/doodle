// Crumbly Floor.
function main() {
	Self.SetHitbox(0, 0, 65, 7);

	Self.AddAnimation("shake", 100, ["shake1", "shake2", "floor", "shake1", "shake2", "floor"]);
	Self.AddAnimation("fall", 100, ["fall1", "fall2", "fall3", "fall4"]);

	// Recover time for the floor to respawn.
	var recover = 5000;

	// States of the floor.
	var stateSolid = 0;
	var stateShaking = 1;
	var stateFalling = 2;
	var stateFallen = 3;
	var state = stateSolid;

	// Started the animation?
	var startedAnimation = false;

	Events.OnCollide(function(e) {
		// Only trigger for mobile characters.
		if (!e.Actor.IsMobile()) {
			return;
		}

		// If the floor is falling, the player passes right thru.
		if (state === stateFalling || state === stateFallen) {
			return;
		}

		// Floor is solid until it begins to fall.
		if (e.InHitbox && (state === stateSolid || state === stateShaking)) {
			// Only activate when touched from the top.
			if (e.Overlap.Y > 0) {
				return false;
			}

			// Begin the animation sequence if we're in the solid state.
			if (state === stateSolid) {
				state = stateShaking;
				Self.PlayAnimation("shake", function() {
					state = stateFalling;
					Self.PlayAnimation("fall", function() {
						state = stateFallen;
						Self.ShowLayerNamed("fallen");

						// Recover after a while.
						setTimeout(function() {
							Self.ShowLayer(0);
							state = stateSolid;
						}, recover);
					});
				})
			}

			return false;
		}
	});
}
