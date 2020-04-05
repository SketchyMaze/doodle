function main() {
	var color = Self.Doodad().Tag("color");

	Events.OnCollide(function(e) {
		if (e.Settled) {
			e.Actor.AddItem(Self.Doodad().Filename, 0);
			Self.Destroy();
		}
	})
}
