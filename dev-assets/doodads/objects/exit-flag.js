// Exit Flag.
function main() {
	Self.SetHitbox(22 + 16, 16, 75 - 16, 86);

	Events.OnCollide((e) => {
		if (!e.Settled) {
			return;
		}

		// Only care if it's the player.
		if (!e.Actor.IsPlayer()) {
			return;
		}

		if (e.InHitbox) {
			EndLevel();
		}
	});
}
