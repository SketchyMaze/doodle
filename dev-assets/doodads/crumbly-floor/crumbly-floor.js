// Crumbly Floor.
function main() {
	Self.SetHitbox(0, 0, 98, 11);

	Self.AddAnimation("shake", 100, ["shake1", "shake2", "floor", "shake1", "shake2", "floor"]);
	Self.AddAnimation("fall", 100, ["fall1", "fall2", "fall3", "fall4"]);

	// Recover time for the floor to respawn.
	let recover = 5000;

	// States of the floor.
	let stateSolid = 0;
	let stateShaking = 1;
	let stateFalling = 2;
	let stateFallen = 3;
	let state = stateSolid;

	Events.OnCollide((e) => {

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

			// If movement is not settled, be solid.
			if (!e.Settled) {
				return false;
			}

			// Begin the animation sequence if we're in the solid state.
			if (state === stateSolid) {
				state = stateShaking;
				Self.PlayAnimation("shake", () => {
					state = stateFalling;
					Self.PlayAnimation("fall", () => {
						Sound.Play("crumbly-break.wav")
						state = stateFallen;
						Self.ShowLayerNamed("fallen");

						// Recover after a while.
						setTimeout(() => {
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
