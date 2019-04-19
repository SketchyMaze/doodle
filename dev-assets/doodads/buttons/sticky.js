function main() {
	console.log("%s initialized!", Self.Doodad.Title);

	Events.OnCollide( function() {
		Self.ShowLayer(1);
	})
}
