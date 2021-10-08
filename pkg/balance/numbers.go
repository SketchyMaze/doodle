package balance

import (
	"time"

	"git.kirsle.net/go/render"
)

// Numbers.
var (
	// Window dimensions.
	Width  = 1024
	Height = 768

	// Speed to scroll a canvas with arrow keys in Edit Mode.
	CanvasScrollSpeed         = 8
	FollowActorMaxScrollSpeed = 64

	// Window scrolling behavior in Play Mode.
	ScrollboxOffset = render.Point{ // from center of screen
		X: 60,
		Y: 60,
	}

	// Player speeds
	PlayerMaxVelocity   float64 = 7
	PlayerJumpVelocity  float64 = -23
	PlayerAcceleration  float64 = 0.12
	Gravity             float64 = 7
	GravityAcceleration float64 = 0.1
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

	// Play Mode Touchscreen controls.
	PlayModeIdleTimeout = 2200 * time.Millisecond
	PlayModeAlphaStep   = 8 // 0-255 alpha, steps per tick for fade in
	PlayModeAlphaMax    = 220
)

// Edit Mode Values
var (
	// Number of Doodads per row in the palette.
	UIDoodadsPerRow = 2
)
