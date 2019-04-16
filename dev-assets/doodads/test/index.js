// Test Doodad Script
function main() {
	console.log("I am actor ID " + Self.ID());

	// Set our doodad's background color to pink. It will be turned
	// red whenever something collides with us.
	Self.Canvas.SetBackground(RGBA(255, 153, 255, 153));

	Events.OnCollide( function(e) {
		console.log("Collided with something!");
		Self.Canvas.SetBackground(RGBA(255, 0, 0, 153));
	});
}
