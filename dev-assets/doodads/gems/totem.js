// Gem stone totem socket.

/*
The Totem is a type of key-door that holds onto its corresponding
Gemstone. When a doodad holding the right Gemstone touches the
totem, the totem takes the gemstone and activates.

If the Totem is not linked to any other Totems, it immediately
sends a power(true) signal upon activation.

If the Totem is linked to other Totems, it waits until all totems
have been activated before it will emit a power signal. Only one
such totem needs to be linked to e.g. an Electric Door - no matter
which totem is solved last, they'll all emit a power signal when
all of their linked totems are activated.
*/

let color = Self.GetTag("color"),
    keyname = "gem-"+color+".doodad",
    activated = false,
    linkedReceiver = false, // is linked to a non-totem which might want power
    totems = {}, // linked totems
    shimmerFreq = 1000;

function main() {
    // Show the hollow socket on level load (last layer)
    Self.ShowLayer(4);

    // Find any linked totems.
    for (let link of Self.GetLinks()) {
        if (link.Filename.indexOf("gem-totem") > -1) {
            totems[link.ID()] = false;
        } else {
            linkedReceiver = true;
        }
    }

    // Shimmer animation is just like the gemstones: first 4 frames
    // are the filled socket sprites.
    Self.AddAnimation("shimmer", 100, [0, 1, 2, 3, 0]);

	Events.OnCollide((e) => {
        if (activated) return;

		if (e.Actor.IsMobile() && e.Settled) {
            // Do they have our gemstone?
            let hasKey = e.Actor.HasItem(keyname) >= 0;
            if (!hasKey) {
                return;
            }

            // Take the gemstone.
            e.Actor.RemoveItem(keyname, 1);
			Self.ShowLayer(0);

            // Emit to our linked totem neighbors.
            activated = true;
            Message.Publish("gem-totem:activated", Self.ID());
            tryPower();
		}
	});

    Message.Subscribe("gem-totem:activated", (totemId) => {
        totems[totemId] = true;
        tryPower();
    })

    setInterval(() => {
        if (activated) {
            Self.PlayAnimation("shimmer", null);
        }
    }, shimmerFreq);
}

// Try to send a power signal for an activated totem.
function tryPower() {
    // Only emit power if we are linked to something other than a totem.
    if (!linkedReceiver) {
        return;
    }

    // Can't if any of our linked totems aren't activated.
    try {
        for (let totemId of Object.keys(totems)) {
            if (totems[totemId] === false) {
                return;
            }
        }
    } catch(e) {
        console.error("Caught: %s", e);
    }

    // Can't if we aren't powered.
    if (activated === false) {
        return;
    }

    // Emit power!
    Message.Publish("power", true);
}