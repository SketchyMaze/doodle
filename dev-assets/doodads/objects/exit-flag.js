// Exit Flag.
function main() {
	Self.SetHitbox(22+16, 16, 75-16, 86);

	Events.OnCollide(function(e) {
		if (!e.Settled) {
			return;
		}

		if (e.InHitbox) {
			EndLevel();
		}
	});
}
