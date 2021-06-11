package balance

// Feature Flags to turn on/off experimental content.
var Feature = feature{
	Zoom: false, // enable the zoom in/out feature (very buggy rn)
	CustomWallpaper: true, // attach custom wallpaper img to levels
	ChangePalette: false, // reset your palette after level creation to a diff preset

	// Allow embedded doodads in levels.
	EmbeddableDoodads: true,
}

// FeaturesOn turns on all feature flags, from CLI --experimental option.
func FeaturesOn() {
	Feature.Zoom = true
	Feature.CustomWallpaper = true
	Feature.ChangePalette = true
}

type feature struct {
	Zoom bool
	CustomWallpaper bool
	ChangePalette bool
	EmbeddableDoodads bool
}
