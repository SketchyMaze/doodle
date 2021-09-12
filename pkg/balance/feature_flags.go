package balance

// Feature Flags to turn on/off experimental content.
var Feature = feature{
	/////////
	// Experimental features that are off by default
	Zoom:          false, // enable the zoom in/out feature (very buggy rn)
	ChangePalette: false, // reset your palette after level creation to a diff preset

	/////////
	// Fully activated features

	// Attach custom wallpaper img to levels
	CustomWallpaper: true,

	// Allow embedded doodads in levels.
	EmbeddableDoodads: true,
}

// FeaturesOn turns on all feature flags, from CLI --experimental option.
func FeaturesOn() {
	Feature.Zoom = true
	Feature.ChangePalette = true
}

type feature struct {
	Zoom              bool
	CustomWallpaper   bool
	ChangePalette     bool
	EmbeddableDoodads bool
}
