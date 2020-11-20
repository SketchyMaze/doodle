package balance

// Feature Flags to turn on/off experimental content.
var Feature = feature{
	Zoom: false,
}

// FeaturesOn turns on all feature flags, from CLI --experimental option.
func FeaturesOn() {
	Feature.Zoom = true
}

type feature struct {
	Zoom bool
}
