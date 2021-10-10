var animating = false;
var opened = false;
var powerState = false;

// Function to handle the door opening or closing.
function setPoweredState(powered) {
	powerState = powered;

	if (powered) {
		if (animating || opened) {
			return;
		}

		animating = true;
		Sound.Play("electric-door.wav")
		Self.PlayAnimation("open", function () {
			opened = true;
			animating = false;
		});
	} else {
		animating = true;
		Sound.Play("electric-door.wav")
		Self.PlayAnimation("close", function () {
			opened = false;
			animating = false;
		})
	}
}

function main() {
	Self.AddAnimation("open", 100, [0, 1, 2, 3]);
	Self.AddAnimation("close", 100, [3, 2, 1, 0]);


	Self.SetHitbox(0, 0, 34, 76);

	// A linked Switch that activates the door will send the Toggle signal
	// immediately before the Power signal. The door can just invert its
	// state on this signal, and ignore the very next Power signal. Ordinary
	// power sources like Buttons will work as normal, as they emit only a power
	// signal.
	var ignoreNextPower = false;
	Message.Subscribe("switch:toggle", function (powered) {
		ignoreNextPower = true;
		setPoweredState(!powerState);
	})

	Message.Subscribe("power", function (powered) {
		if (ignoreNextPower) {
			ignoreNextPower = false;
			return;
		}

		setPoweredState(powered);
	});

	Events.OnCollide(function (e) {
		if (e.InHitbox) {
			if (!opened) {
				return false;
			}
		}
	});
}
