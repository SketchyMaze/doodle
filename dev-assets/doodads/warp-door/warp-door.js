// Warp Doors
function main() {
	console.log("Warp Door %s Initialized", Self.Title);

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

	console.log("Warp %s: default=%s  disabled=%+v  color=%s  isState=%+v  state=%+v", Self.Title, spriteDefault, spriteDisabled, color, isStateDoor, state);

	// Subscribe to the global state-change if we are a state door.
	if (isStateDoor) {
		Message.Subscribe("broadcast:state-change", function(newState) {
			console.log("Warp %s: received state to %+v", Self.Title, newState);
			state = color === 'blue' ? !newState : newState;

			// Activate or deactivate the door.
			Self.ShowLayerNamed(state ? spriteDefault : spriteDisabled);
		});
	}

	// TODO: respond to a "Use" button instead of a Collide to open the door.
	Events.OnCollide(function(e) {
		if (!e.Settled) {
			return;
		}

		if (animating || collide) {
			return;
		}

		// Only players can use doors for now.
		if (e.Actor.IsPlayer() && e.InHitbox) {
			if (isStateDoor && !state) {
				// The state door is inactive (dotted outline).
				return;
			}

			// Play the open and close animation.
			animating = true;
			collide = true;
			Self.PlayAnimation("open", function() {
				e.Actor.Hide()
				Self.PlayAnimation("close", function() {
					Self.ShowLayerNamed(isStateDoor && !state ? spriteDisabled : spriteDefault);
					e.Actor.Show()
					animating = false;
				});
			});
		}
	});

	Events.OnLeave(function(e) {
		collide = false;
	});
}
