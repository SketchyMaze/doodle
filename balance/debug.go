package balance

import (
	"os"

	"git.kirsle.net/apps/doodle/render"
)

// Debug related variables that can toggle on or off certain features and
// overlays within the game.
var (
	/***************
	 * Visualizers *
	 ***************/

	// Background color to use when exporting a drawing Chunk as a bitmap image
	// on disk. Default is white. Setting this to translucent yellow is a great
	// way to visualize the chunks loaded from cache on your screen.
	DebugChunkBitmapBackground = render.White // XXX: export $DEBUG_CHUNK_COLOR
)

func init() {
	if color := os.Getenv("DEBUG_CHUNK_COLOR"); color != "" {
		DebugChunkBitmapBackground = render.MustHexColor(color)
	}
}
