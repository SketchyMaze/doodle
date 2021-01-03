// Warp Doors
function main() {
	Self.SetHitbox(0, 0, 34, 76);

	// Are we a blue or orange door? Regular warp door will be 'none'
	var color = Self.GetTag("color");
	var isStateDoor = color === 'blue' || color === 'orange';
	var state = color === 'blue';  // Blue door is ON by default.

	var animating = false;
	var collide = false;

	// Declare animations and sprite names.
	var animSpeed = 100;
	var spriteDefault, spriteDisabled;  // the latter for state doors.
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
	} else {
		Self.AddAnimation("open", animSpeed, ["door-2", "door-3", "door-4"]);
		Self.AddAnimation("close", animSpeed, ["door-4", "door-3", "door-2", "door-1"]);
		spriteDefault = "door-1";
	}

	// Find our linked Warp Door.
	var links = Self.GetLinks()
	var linkedDoor = null;
	for (var i = 0; i < links.length; i++) {
		if (links[i].Title.indexOf("Warp Door") > -1) {
			linkedDoor = links[i];
		}
	}

	// Subscribe to the global state-change if we are a state door.
	if (isStateDoor) {
		Message.Subscribe("broadcast:state-change", function(newState) {
			state = color === 'blue' ? !newState : newState;

			// Activate or deactivate the door.
			Self.ShowLayerNamed(state ? spriteDefault : spriteDisabled);
		});
	}

	// The player Uses the door.
	var flashedCooldown = false; // "Locked Door" flashed message.
	Events.OnUse(function(e) {
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

			// Freeze the player.
			e.Actor.Freeze()

			// Play the open and close animation.
			animating = true;
			Self.PlayAnimation("open", function() {
				e.Actor.Hide()
				Self.PlayAnimation("close", function() {
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
	Message.Subscribe("warp-door:incoming", function(player) {
		animating = true;
		player.Unfreeze();
		Self.PlayAnimation("open", function() {
			player.Show();
			Self.PlayAnimation("close", function() {
				animating = false;

				// If the receiving door was a State Door, fix its state.
				if (isStateDoor) {
					Self.ShowLayerNamed(state ? spriteDefault : spriteDisabled);
				}
			});
		});
	});
}
