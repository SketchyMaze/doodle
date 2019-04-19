function main() {
	console.log("%s initialized!", Self.Doodad.Title);

	console.log(Object.keys(console));
	console.log(Object.keys(log));
	console.log(Object.keys(log.Config));
	console.log(Object.keys(Self.Canvas.Palette));
	console.log(Object.keys(Self.Canvas.Palette.Swatches[0]));

	Self.Canvas.Palette.Swatches[0].Color = RGBA(255, 0, 255, 255);
	Self.Canvas.Palette.Swatches[1].Color = RGBA(0, 255, 255, 255);
	console.log(Self.Canvas.Palette.Swatches);
	log.Config.TimeFormat = "haha";

	var colors = [
		RGBA(255, 0, 0, 255),
		RGBA(255, 153, 0, 255),
		RGBA(255, 255, 0, 255),
		RGBA(0, 255, 0, 255),
		RGBA(0, 153, 255, 255),
		RGBA(0, 0, 255, 255),
		RGBA(255, 0, 255, 255)
	];
	var colorIndex = 0;
	setInterval(function() {
		console.log("sticky tick");
		Self.Canvas.MaskColor = colors[colorIndex];
		colorIndex++;
		if (colorIndex == colors.length) {
			colorIndex = 0;
		}
	}, 100);

	// log.Config.Colors = 0; // panics, can't set a golog.Color

	Events.OnCollide( function() {

		Self.ShowLayer(1);
		setTimeout(function() {
			Self.ShowLayer(0);
		}, 200);
	})
}
