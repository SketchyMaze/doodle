function main() {
	// What direction is the trapdoor facing?
	var direction = Self.Doodad.Tag("direction");
	console.log("Trapdoor(%s) initialized", direction);

	var timer = 0;

	// Set our hitbox based on our orientation.
	var thickness = 6;
	var doodadSize = 72;
	if (direction === "left") {
		Self.SetHitbox(48, 0, doodadSize, doodadSize);
	} else if (direction === "right") {
		Self.SetHitbox(0, 0, thickness, doodadSize);
	} else if (direction === "up") {
		Self.SetHitbox(0, doodadSize - thickness, doodadSize, doodadSize);
	} else { // Down, default.
		Self.SetHitbox(0, 0, doodadSize, thickness);
	}

	var animationSpeed = 100;
	var opened = false;

	// Register our animations.
	var frames = [];
	for (var i = 1; i <= 4; i++) {
		frames.push(direction + i);
	}

	Self.AddAnimation("open", animationSpeed, frames);
	frames.reverse();
	Self.AddAnimation("close", animationSpeed, frames);

	Events.OnCollide( function(e) {
		if (opened) {
			return;
		}

		// Is the actor colliding our solid part?
		if (e.InHitbox) {
			// Are they touching our opening side?
			if (direction === "left") {
				if (doodadSize - e.Overlap.X < thickness) {
					// Touching the right edge, open the door.
					opened = true;
					Self.PlayAnimation("open", null);
					return;
				}
				if (e.Overlap.W === doodadSize - thickness) {
					return false;
				}
			} else if (direction === "right") {
				if (e.Overlap.X > 0) {
					return false;
				} else if (e.Settled) {
					opened = true;
					Self.PlayAnimation("open", null);
				}
			} else if (direction === "up") {
				if (doodadSize - e.Overlap.Y < thickness) {
					// Touching the bottom edge, open the door.
					opened = true;
					Self.PlayAnimation("open", null);
					return;
				}
				if (e.Overlap.H === doodadSize - thickness) {
					return false;
				}
			} else if (direction === "down") {
				if (e.Overlap.Y > 0) {
					return false;
				} else if (e.Settled) {
					opened = true;
					Self.PlayAnimation("open", null);
				}
			}

			return true;
		}
	});

	Events.OnLeave(function() {
		if (opened) {
			Self.PlayAnimation("close", function() {
				opened = false;
			});
		}
	})
}
