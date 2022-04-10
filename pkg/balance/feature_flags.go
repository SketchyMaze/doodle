package balance

// Feature Flags to turn on/off experimental content.
var Feature = feature{
	/////////
	// Experimental features that are off by default
	ViewportWindow: false, // Open new viewport into your level

	/////////
	// Fully activated features

	// Attach custom wallpaper img to levels
	CustomWallpaper: true,

	// Allow embedded doodads in levels.
	EmbeddableDoodads: true,

	// Enable the zoom in/out feature (kinda buggy still)
	Zoom: true,

	// Reassign an existing level's palette to a different builtin.
	ChangePalette: true,

	// LoadUnloadChunk feature to better optimize memory. Set it to false and the
	// loadscreen will eager load all chunk bitmaps (stable, but uses a lot of
	// memory), set true and the Canvas will load/unload bitmaps + free SDL textures
	// for chunks falling outside the LoadingViewport (new, maybe unstable).
	LoadUnloadChunk: true,
}

// FeaturesOn turns on all feature flags, from CLI --experimental option.
func FeaturesOn() {
	Feature.ViewportWindow = true
}

type feature struct {
	Zoom              bool
	CustomWallpaper   bool
	ChangePalette     bool
	EmbeddableDoodads bool
	ViewportWindow    bool
	LoadUnloadChunk   bool
}
