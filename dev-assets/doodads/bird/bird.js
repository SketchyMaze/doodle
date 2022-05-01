// Bird (red and blue)

/*
Base A.I. behaviors (red bird) are:

- Tries to maintain original altitude in level and flies left/right.
- Divebombs the player to attack, then climbs back to its original
  altitude when it hits a floor/wall.

Blue bird:

- Flies in a sine wave pattern (its target altitude fluctuates
  around the bird's original placement on the level).
- Longer aggro radius to dive.
*/

let speed = 4,
	Vx = Vy = 0,
	color = Self.GetTag("color"), // informs our A.I. behaviors
	searchStep = 12 // pixels to step while searching for a player
	searchLimit = color === "blue" ? 24 : 12, // multiples of searchStep for aggro radius
	altitude = Self.Position().Y, // original height in level
	targetAltitude = altitude;    // bird's target height to maintain

let direction = "left",
	lastDirection = "left";
let states = {
	flying: 0,
	diving: 1,
};
let state = states.flying;

function main() {
	Self.SetMobile(true);
	Self.SetGravity(false);
	Self.SetHitbox(0, 0, 46, 32);
	Self.AddAnimation("fly-left", 100, ["left-1", "left-2"]);
	Self.AddAnimation("fly-right", 100, ["right-1", "right-2"]);

	// Player Character controls?
	if (Self.IsPlayer()) {
		return player();
	}

	Events.OnCollide((e) => {
		// If we're diving and we hit the player, game over!
		if (e.Settled && state === states.diving && e.Actor.IsPlayer()) {
			FailLevel("Watch out for birds!");
			return;
		}

		if (e.Actor.IsMobile() && e.Actor.HasGravity() && e.InHitbox) {
			return false;
		}
	});

	// Sample our X position every few frames and detect if we've hit a solid wall.
	let sampleTick = 0,
		sampleRate = 2,
		lastSampled = Point(0, 0);

	setInterval(() => {
		// Run blue bird A.I. if we are blue: moves our target altitude along a sine wave.
		if (color === "blue") {
			AI_BlueBird();
		}

		// Sample how far we've moved to detect hitting a wall.
		if (sampleTick % sampleRate === 0) {
			let curP = Self.Position()
			let delta = Math.abs(curP.X - lastSampled.X);
			if (delta < 5) {
				direction = direction === "right" ? "left" : "right";
			}

			// If we were diving, check Y delta too for if we hit floor
			if (state === states.diving && Math.abs(curP.Y - lastSampled.Y) < 5) {
				state = states.flying;
			}
			lastSampled = curP
		}
		sampleTick++;

		// Are we diving?
		if (state === states.diving) {
			Vy = speed
		} else {
			// If we are not flying at our original altitude, correct for that.
			let curV = Self.Position();
			Vy = 0.0;
			if (curV.Y != targetAltitude) {
				Vy = curV.Y < targetAltitude ? 1 : -1;
			}

			// Scan for the player character and dive.
			try {
				AI_ScanForPlayer()
			} catch (e) {
				console.error("Error in AI_ScanForPlayer: %s", e);
			}
		}

		// TODO: Vector() requires floats, pain in the butt for JS,
		// the JS API should be friendlier and custom...
		let Vx = parseFloat(speed * (direction === "left" ? -1 : 1));
		Self.SetVelocity(Vector(Vx, Vy));

		// If diving, exit - don't edit animation.
		if (state === states.diving) {
			Self.ShowLayerNamed("dive-" + direction);
			lastDirection = direction;
			return;
		}

		// If we changed directions, stop animating now so we can
		// turn around quickly without moonwalking.
		if (direction !== lastDirection) {
			Self.StopAnimation();
		}

		if (!Self.IsAnimating()) {
			Self.PlayAnimation("fly-" + direction, null);
		}

		lastDirection = direction;
	}, 100);
}

// A.I. subroutine: scan for the player character.
// The bird scans in a 45 degree angle downwards, if the
// player is seen nearby in that scan it will begin a dive.
// It's not hostile towards characters that can fly (having
// no gravity).
function AI_ScanForPlayer() {
	// If Peaceful difficulty, do not attack.
	if (Level.Difficulty < 0) {
		return
	}

	let stepY = searchStep, // number of pixels to skip
		stepX = stepY,
		limit = stepX * searchLimit, // furthest we'll scan
		scanX = scanY = 0,
		size = Self.Size(),
		fromPoint = Self.Position();

	// From what point do we begin the scan?
	if (direction === 'left') {
		stepX = -stepX;
		fromPoint.Y += size.H;
	} else {
		fromPoint.Y += size.H;
		fromPoint.X += size.W;
	}

	scanX = fromPoint.X;
	scanY = fromPoint.Y;

	for (let i = 0; i < limit; i += stepY) {
		scanX += stepX;
		scanY += stepY;
		for (let actor of Actors.At(Point(scanX, scanY))) {
			if (actor.IsPlayer() && actor.HasGravity()) {
				state = states.diving;
				return;
			}
		}
	}

	return;
}

// Sine wave controls for the Blue bird.
var sineLimit = 32,
	sineCounter = 0,
	sineDirection = 1,
	sineFrequency = 5, // every 500ms
	sineStep = 16;

// A.I. Subroutine: sine wave pattern (Blue bird).
function AI_BlueBird() {
	// The main loop runs on a 100ms interval, control how frequently
	// to adjust the bird's target velocity.
	if (sineCounter > 0 && (sineCounter % sineFrequency) > 1) {
		sineCounter++;
		return;
	}
	sineCounter++;

	targetAltitude += sineStep*sineDirection;

	// Cap the distance between our starting altitude and sine-wave target.
	if (targetAltitude < altitude && altitude - targetAltitude >= sineLimit) {
		targetAltitude = altitude - sineLimit;
		sineDirection = 1
	} else if (targetAltitude > altitude && targetAltitude - altitude >= sineLimit) {
		targetAltitude = altitude + sineLimit;
		sineDirection = -1
	}
}

// If under control of the player character.
function player() {
	let playerSpeed = 12,
		diving = false,
		falling = false;

	// The player can dive by moving downwards and laterally, but
	// de-cheese their ability to just sweep across the ground; if
	// they aren't seen to be moving downwards, cancel the dive.
	let lastPoint = Self.Position();
	setInterval(() => {
		let nowAt = Self.Position();
		if (nowAt.Y > lastPoint.Y) {
			falling = true;
		} else {
			falling = false;
		}
		lastPoint = nowAt;
	}, 100);

	Events.OnKeypress((ev) => {
		Vx = 0;
		Vy = 0;

		if (ev.Right) {
			direction = "right";
		} else if (ev.Left) {
			direction = "left";
		}

		// Dive!
		if (ev.Down && ev.Right && falling) {
			Self.StopAnimation();
			Self.ShowLayerNamed("dive-right");
			diving = falling;
		} else if (ev.Down && ev.Left && falling) {
			Self.StopAnimation();
			Self.ShowLayerNamed("dive-left");
			diving = falling;
		} else if (ev.Right) {
			// Fly right.
			if (!Self.IsAnimating()) {
				Self.PlayAnimation("fly-right", null);
			}
			Vx = playerSpeed;
			diving = false;
		} else if (ev.Left) {
			// Fly left.
			if (!Self.IsAnimating()) {
				Self.PlayAnimation("fly-left", null);
			}
			Vx = -playerSpeed;
			diving = false;
		} else {
			// Hover in place.
			if (!Self.IsAnimating()) {
				Self.PlayAnimation("fly-" + direction);
			}
			diving = false;
		}

		// Player is invulnerable while diving.
		Self.SetInvulnerable(diving);
	});

	Events.OnCollide((e) => {
		// If the player is diving at an enemy mob, destroy it.
		if (diving && e.Settled && e.Actor.IsMobile() && !e.Actor.Invulnerable()) {
			Sound.Play("crumbly-break.wav");
			e.Actor.Destroy();
		}
	});
}
