package balance

import "git.kirsle.net/go/render"

// Numbers.
var (
	// Window dimensions.
	Width  = 1024
	Height = 768

	// Speed to scroll a canvas with arrow keys in Edit Mode.
	CanvasScrollSpeed = 8

	// Window scrolling behavior in Play Mode.
	ScrollboxOffset = render.Point{ // from center of screen
		X: 40,
		Y: 100,
	}

	// Player speeds
	PlayerMaxVelocity   float64 = 6
	PlayerAcceleration  float64 = 0.9
	Gravity             float64 = 6
	GravityAcceleration float64 = 0.2
	SlopeMaxHeight              = 8 // max pixel height for player to walk up a slope

	// Default chunk size for canvases.
	ChunkSize = 128

	// Default size for a new Doodad.
	DoodadSize = 100

	// Size of Undo/Redo history for map editor.
	UndoHistory = 20

	// Options for brush size.
	BrushSizeOptions = []int{
		0,
		1,
		2,
		4,
		8,
		16,
		24,
		32,
		48,
		64,
	}
	DefaultEraserBrushSize = 8
	MaxEraserBrushSize     = 32 // the bigger, the slower

	// Default player character doodad in Play Mode.
	PlayerCharacterDoodad = "boy.doodad"

	// Level name for the title screen.
	DemoLevelName = "Tutorial 3.level"

	// Level attachment filename for the custom wallpaper.
	// NOTE: due to hard-coded "assets/wallpapers/" prefix in uix/canvas.go#LoadLevel.
	CustomWallpaperFilename  = "custom.b64img"
	CustomWallpaperEmbedPath = "assets/wallpapers/custom.b64img"

	// Publishing: Doodads-embedded-within-levels.
	EmbeddedDoodadsBasePath   = "assets/doodads/"
	EmbeddedWallpaperBasePath = "assets/wallpapers/"

	// File formats: save new levels and doodads gzip compressed
	CompressDrawings = true
)

// Edit Mode Values
var (
	// Number of Doodads per row in the palette.
	UIDoodadsPerRow = 2
)
