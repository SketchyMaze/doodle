function main() {
	console.log("Sticky Button initialized!");

	Events.OnCollide( function() {
		console.log("Touched!");
		Self.Canvas.SetBackground(RGBA(255, 153, 0, 153))
	})
}
