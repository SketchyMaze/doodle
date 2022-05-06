// Azulian (Red and Blue)

const color = Self.GetTag("color");
var playerSpeed = color === 'blue' ? 2 : 4,
	swimSpeed = playerSpeed * 0.4,
	aggroX = 250,  // X/Y distance sensitivity from player
	aggroY = color === 'blue' ? 100 : 200,
	jumpSpeed = color === 'blue' ? 14 : 18,
	swimJumpSpeed = jumpSpeed * 0.4,
	animating = false,
	direction = "right",
	lastDirection = "right";

// white Azulian is faster yet than the red
if (color === 'white') {
	aggroX = 1000;
	aggroY = 400;
	playerSpeed = 8;
	swimSpeed = playerSpeed * 0.4;
	jumpSpeed = 20;
	swimJumpSpeed = jumpSpeed * 0.4;
}

function setupAnimations(color) {
	let left = color === 'blue' ? 'blu-wl' : color + '-wl',
		right = color === 'blue' ? 'blu-wr' : color + '-wr',
		leftFrames = [left + '1', left + '2', left + '3', left + '4'],
		rightFrames = [right + '1', right + '2', right + '3', right + '4'];

	Self.AddAnimation("walk-left", 100, leftFrames);
	Self.AddAnimation("walk-right", 100, rightFrames);
}

function main() {
	playerSpeed = color === 'blue' ? 2 : 4;

	let swimJumpCooldownTick = 0, // minimum Game Tick before we can jump while swimming
		swimJumpCooldown = 10;    // CONFIG: delta of ticks between jumps while swimming

	Self.SetMobile(true);
	Self.SetGravity(true);
	Self.SetInventory(true);
	Self.SetHitbox(0, 0, 24, 32);
	setupAnimations(color);

	if (Self.IsPlayer()) {
		return playerControls();
	}

	// A.I. pattern: walks back and forth, turning around
	// when it meets resistance.

	// Sample our X position every few frames and detect if we've hit a solid wall.
	let sampleTick = 0;
	let sampleRate = 5;
	let lastSampledX = 0;

	// Get the player on touch.
	Events.OnCollide((e) => {
		// If we're diving and we hit the player, game over!
		// Azulians are friendly to Thieves though!
		if (e.Settled && isPlayerFood(e.Actor)) {
			FailLevel("Watch out for the Azulians!");
			return;
		}
	});

	setInterval(() => {
		// If the player is nearby, walk towards them. Otherwise, default pattern
		// is to walk back and forth.
		let player = Actors.FindPlayer(),
			followPlayer = false,
			jump = false;

		// Don't follow boring players.
		if (player !== null && isPlayerFood(player)) {
			let playerPt = player.Position(),
				myPt = Self.Position();

			// If the player is within aggro range, move towards.
			if ((Math.abs(playerPt.X - myPt.X) < aggroX && Math.abs(playerPt.Y - myPt.Y) < aggroY)
				|| (Level.Difficulty > 0)) {
				direction = playerPt.X < myPt.X ? "left" : "right";
				followPlayer = true;

				if (playerPt.Y + player.Size().H < myPt.Y + Self.Size().H) {
					jump = true;
				}
			}
		}

		// Default AI: sample position so we turn around on obstacles.
		if (!followPlayer) {
			if (sampleTick % sampleRate === 0) {
				let curX = Self.Position().X;
				let delta = Math.abs(curX - lastSampledX);
				if (delta < 5) {
					direction = direction === "right" ? "left" : "right";
				}
				lastSampledX = curX;
			}
			sampleTick++;
		}

		// Handle being underwater.
		let canJump = Self.Grounded();
		if (Self.IsWet()) {
			let tick = GetTick();
			if (tick > swimJumpCooldownTick) {
				canJump = true;
				swimJumpCooldownTick = tick + swimJumpCooldown;
			}
		}

		// How speedy would our movement and jump be?
		let xspeed = playerSpeed, yspeed = jumpSpeed;
		if (Self.IsWet()) {
			xspeed = swimSpeed;
			yspeed = swimJumpSpeed;
		}

		let Vx = parseFloat(xspeed * (direction === "left" ? -1 : 1)),
			Vy = jump && canJump ? parseFloat(-yspeed) : Self.GetVelocity().Y;
		Self.SetVelocity(Vector(Vx, Vy));

		// If we changed directions, stop animating now so we can
		// turn around quickly without moonwalking.
		if (direction !== lastDirection) {
			Self.StopAnimation();
		}

		if (!Self.IsAnimating()) {
			Self.PlayAnimation("walk-" + direction, null);
		}

		lastDirection = direction;
	}, 10);
}

function playerControls() {
	// Note: player speed is controlled by the engine.
	Events.OnKeypress((ev) => {
		if (ev.Right) {
			if (!Self.IsAnimating()) {
				Self.PlayAnimation("walk-right", null);
			}
		} else if (ev.Left) {
			if (!Self.IsAnimating()) {
				Self.PlayAnimation("walk-left", null);
			}
		} else {
			Self.StopAnimation();
			animating = false;
		}
	})
}

// Logic to decide if the player is interesting to the Azulian (aka, if the Azulian
// will be hostile towards the player). Boring players will not be chased after and
// the Azulian will not harm them if they make contact.
function isPlayerFood(actor) {
	// Not a player or is invulnerable, or Peaceful difficulty.
	if (!actor.IsPlayer() || actor.Invulnerable() || Level.Difficulty < 0) {
		return false;
	}

	// On hard mode they are hostile to any player.
	if (Level.Difficulty > 0) {
		return true;
	}

	// Azulians are friendly to Thieves and other Azulians.
	if (actor.Doodad().Filename === "thief.doodad" || actor.Doodad().Title.indexOf("Azulian") > -1) {
		return false;
	}

	return true;
}
