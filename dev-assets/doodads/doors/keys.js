function main() {
	var color = Self.Doodad.Tag("color");

	Events.OnCollide(function(e) {
		e.Actor.SetData("key:" + color, "true");
		Self.Destroy();
	})
}
