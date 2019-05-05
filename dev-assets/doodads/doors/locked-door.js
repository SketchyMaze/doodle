function main() {
	Self.AddAnimation("open", 0, [1]);
	var unlocked = false;

	Events.OnCollide(function(e) {
		if (unlocked) {
			return;
		}

		unlocked = true;
		Self.PlayAnimation("open", null);
	});
}
