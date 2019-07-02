// Exit Flag.
function main() {
	console.log("%s initialized!", Self.Doodad.Title);
	Self.SetHitbox(22+16, 16, 75-16, 86);

	Events.OnCollide(function(e) {
		if (e.InHitbox) {
			EndLevel();
		}
	});
}
