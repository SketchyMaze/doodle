package balance

// Feature Flags to turn on/off experimental content.
var Feature = feature{
	Zoom: false,
	CustomWallpaper: false,
	ChangePalette: false,
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
}
