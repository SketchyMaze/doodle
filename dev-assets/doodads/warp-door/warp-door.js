// Warp Doors

const color = Self.GetTag("color"),
	isStateDoor = color === 'blue' || color === 'orange';

// State in case we're a blue warp door.
let state = color === 'blue',
	animating = false,
	collide = false;

function main() {
	// Declare animations and sprite names.
	let animSpeed = 100;
	let spriteDefault, spriteDisabled;  // the latter for state doors.
	if (color === 'blue') {
		Self.AddAnimation("open", animSpeed, ["blue-2", "blue-3", "blue-4"]);
		Self.AddAnimation("close", animSpeed, ["blue-4", "blue-3", "blue-2", "blue-1"]);
		spriteDefault = "blue-1";
		spriteDisabled = "blue-off";
	} else if (color === 'orange') {
		Self.AddAnimation("open", animSpeed, ["orange-2", "orange-3", "orange-4"]);
		Self.AddAnimation("close", animSpeed, ["orange-4", "orange-3", "orange-2", "orange-1"]);
		spriteDefault = "orange-1";
		spriteDisabled = "orange-off";
	} else if (color === 'invisible') {
		// Invisible Warp Door region.
		Self.Hide();
		Self.AddAnimation("open", animSpeed, [0, 0]);
		Self.AddAnimation("close", animSpeed, [0, 0]);
	} else {
		Self.AddAnimation("open", animSpeed, ["door-2", "door-3", "door-4"]);
		Self.AddAnimation("close", animSpeed, ["door-4", "door-3", "door-2", "door-1"]);
		spriteDefault = "door-1";
	}

	// Find our linked Warp Door.
	let linkedDoor = null;
	for (let link of Self.GetLinks()) {
		if (link.Title.indexOf("Warp Door") > -1) {
			linkedDoor = link;
			break;
		}
	}

	// Subscribe to the global state-change if we are a state door.
	if (isStateDoor) {
		Message.Subscribe("broadcast:state-change", (newState) => {
			state = color === 'blue' ? !newState : newState;

			// Activate or deactivate the door.
			Self.ShowLayerNamed(state ? spriteDefault : spriteDisabled);
		});
	}

	// For player groundedness work-around
	let playerLastY = []; // last sampling of Y values

	// The player Uses the door.
	let flashedCooldown = false; // "Locked Door" flashed message.
	Events.OnUse((e) => {
		if (animating) {
			return;
		}

		// Doors without linked exits are not usable.
		if (linkedDoor === null) {
			if (!flashedCooldown) {
				Flash("This door is locked.");
				flashedCooldown = true;
				setTimeout(function() {
					flashedCooldown = false;
				}, 1000);
			}
			return;
		}

		// Only players can use doors for now.
		if (e.Actor.IsPlayer()) {
			if (isStateDoor && !state) {
				// The state door is inactive (dotted outline).
				return;
			}

			// The player must be grounded or have no gravity to open the door.
			if (!e.Actor.Grounded() && e.Actor.HasGravity()) {
				// Work-around: if two Boxes are stacked atop each other the player can
				// get stuck if he jumps on top. He may not be Grounded but isn't changing
				// effective Y position and a warp door may work as a good way out.
				let yValue = e.Actor.Position().Y;

				// Collect a sampling of last few Y values. If the player Y position
				// is constant the last handful of frames, treat them as if they're
				// grounded (or else they can't activate the warp door).
				playerLastY.unshift(yValue);
				if (playerLastY.length < 6) {
					return;
				}

				// We have enough history.
				playerLastY.pop();

				// Hasn't moved?
				let isGrounded = true;
				for (let i = 0; i < playerLastY.length; i++) {
					if (yValue !== playerLastY[i]) {
						isGrounded = false;
						break;
					}
				}

				if (!isGrounded) {
					return;
				}

				// Player was effectively grounded! No change in Y position.
			}

			// Freeze the player.
			e.Actor.Freeze()

			// Play the open and close animation.
			animating = true;
			Self.PlayAnimation("open", () => {
				e.Actor.Hide()
				Self.PlayAnimation("close", () => {
					Self.ShowLayerNamed(isStateDoor && !state ? spriteDisabled : spriteDefault);
					animating = false;

					// Teleport the player to the linked door. Inform the target
					// door of the arrival of the player so it doesn't trigger
					// to send the player back here again on a loop.
					if (linkedDoor !== null) {
						Message.Publish("warp-door:incoming", e.Actor);
						e.Actor.MoveTo(linkedDoor.Position());
					}
				});
			});
		}
	});

	// Respond to incoming warp events.
	Message.Subscribe("warp-door:incoming", (player) => {
		animating = true;
		player.Unfreeze();
		Self.PlayAnimation("open", () => {
			player.Show();
			Self.PlayAnimation("close", () => {
				animating = false;

				// If the receiving door was a State Door, fix its state.
				if (isStateDoor) {
					Self.ShowLayerNamed(state ? spriteDefault : spriteDisabled);
				}
			});
		});
	});
}
