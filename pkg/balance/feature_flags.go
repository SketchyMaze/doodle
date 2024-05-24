package balance

// Hard-coded feature flags.
const (
	// Enable "v1.5" compression in the MapAccessor Chunker.
	//
	// The original MapAccessor encodes a chunk to json using syntax like
	// {"x,y": index} mapping coordinates to palette swatches.
	//
	// With compression on, it is encoded to a byte stream of x,y,index
	// triplets. The game can read both formats and will follow this flag
	// on all saves. NOTE: this applies to when we still use JSON format.
	// If BinaryChunkerEnabled, map accessors are always compressed as they
	// are written to .bin files instead of .json.
	CompressMapAccessor = true

	// Enable "v2" binary storage of Chunk data in Zipfiles.
	//
	// This is a separate toggle to the CompressMapAccessor. Some possible
	// variations of these flags includes:
	//
	// - CompressMapAccessor=true alone, will write the compressed bytes
	//   still wrapped in the JSON format as a Base64 encoded string.
	// - With BinaryChunkerEnabled=true: all chunks are encoded to
	//   binary and put in the zip as .bin instead of as .json files.
	//   MapAccessor is always compressed in binary mode.
	//
	// If you set both flags to false, level zipfiles will use the classic
	// json chunk format as before on save.
	BinaryChunkerEnabled = true

	// Enable "v3" Run-Length Encoding for level chunker.
	//
	// This only supports Zipfile levels and will use the ".bin" format
	// enabled by the previous setting.
	RLEBinaryChunkerEnabled = true
)

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
