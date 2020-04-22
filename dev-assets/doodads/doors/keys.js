function main() {
	var color = Self.GetTag("color");

	Events.OnCollide(function(e) {
		if (e.Settled) {
			e.Actor.AddItem(Self.Filename, 0);
			Self.Destroy();
		}
	})
}
