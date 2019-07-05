function main() {
	// What direction is the trapdoor facing?
	// - Titles are like "Trapdoor Left" or "Trapdoor Right"
	// - The default (down) is called just "Trapdoor"
	var direction = Self.Doodad.Title.split(" ")[1];
	if (!direction) {
		direction = "down";
	}
	direction = direction.toLowerCase();

	console.log("Trapdoor(%s) initialized", direction);

	var timer = 0;

	// Set our hitbox based on our orientation.
	var thickness = 6;
	var doodadSize = 72;
	if (direction === "left") {
		Self.SetHitbox(48, 0, doodadSize, doodadSize);
	} else if (direction === "right") {
		Self.SetHitbox(0, 0, thickness+4, doodadSize);
	} else if (direction === "up") {
		Self.SetHitbox(0, doodadSize - thickness, doodadSize, doodadSize);
	} else { // Down, default.
		Self.SetHitbox(0, 0, 72, 6);
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
			if (direction === "left" && (e.Overlap.X+e.Overlap.W) < (doodadSize-thickness)) {
				return false;
			} else if (direction === "right" && e.Overlap.X > 0) {
				return false;
			} else if (direction === "up" && (e.Overlap.Y+e.Overlap.H) < doodadSize) {
				return false;
			} else if (direction === "down" && e.Overlap.Y > 0) {
				return false;
			} else {
				opened = true;
				Self.PlayAnimation("open", null);
			}
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
