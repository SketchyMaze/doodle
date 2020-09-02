// Package sound manages music and sound effects.
package sound

import "path/filepath"

// Package globals.
var (
	// If enabled is false, all sound functions are no-ops.
	Enabled bool

	// Root folder on disk where sound and music files should live.
	SoundRoot = filepath.Join("rtp", "sfx")
	MusicRoot = filepath.Join("rtp", "music")
)
