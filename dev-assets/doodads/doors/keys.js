function main() {
	Events.OnCollide(function(e) {
		console.log("%s picked up by %s", Self.Doodad.Title, e.Actor.Title);
		e.Actor.SetData("key:" + Self.Doodad.Title, "true");
		Self.Destroy();
	})
}
