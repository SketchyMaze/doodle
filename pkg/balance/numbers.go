package balance

import (
	"time"

	"git.kirsle.net/go/render"
)

// Format for level and doodad files.
type Format int

const (
	FormatJSON    Format = iota // v0: plain json files
	FormatGZip                  // v1: gzip compressed json files
	FormatZipfile               // v2: zip archive with external chunks
)

// Numbers.
var (
	// Window dimensions.
	Width  = 1024
	Height = 768

	// Title screen height needed for the main menu. Phones in landscape
	// mode will switch to the horizontal layout if less than this height.
	TitleScreenResponsiveHeight = 600

	// Speed to scroll a canvas with arrow keys in Edit Mode.
	CanvasScrollSpeed         = 8
	FollowActorMaxScrollSpeed = 64

	// Window scrolling behavior in Play Mode.
	ScrollboxOffset = render.Point{ // from center of screen
		X: 60,
		Y: 60,
	}

	// Player speeds
	PlayerMaxVelocity    float64 = 7
	PlayerJumpVelocity   float64 = -23
	PlayerAcceleration   float64 = 0.04 // 0.12
	PlayerFriction       float64 = 0.1
	SlipperyAcceleration float64 = 0.02
	SlipperyFriction     float64 = 0.02
	Gravity              float64 = 7
	GravityAcceleration  float64 = 0.1
	SwimGravity          float64 = 3
	SwimJumpVelocity     float64 = -12
	SwimJumpCooldown     uint64  = 24 // number of frames of cooldown between swim-jumps
	SlopeMaxHeight               = 8  // max pixel height for player to walk up a slope

	// Number of game ticks to insist the canvas follows the player at the start
	// of a level - to overcome Anvils settling into their starting positions so
	// they don't steal the camera focus straight away.
	FollowPlayerFirstTicks uint64 = 20

	// Default chunk size for canvases.
	ChunkSize uint8 = 128

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

	// Default font filename selected for Text Tool in the editor.
	// TODO: better centralize font filenames, here and in theme.go
	TextToolDefaultFont = SansSerifFont

	// Interval for auto-save in the editor
	AutoSaveInterval = 5 * time.Minute

	// Default player character doodad in Play Mode.
	PlayerCharacterDoodad = "boy.doodad"

	// Levelpack and level names for the title screen.
	DemoLevelPack = "assets/levelpacks/builtin-Tutorial.levelpack"
	DemoLevelName = []string{
		"Tutorial 1.level",
		"Tutorial 2.level",
		"Tutorial 3.level",
		"Tutorial 5.level",
	}

	// Level attachment filename for the custom wallpaper.
	// NOTE: due to hard-coded "assets/wallpapers/" prefix in uix/canvas.go#LoadLevel.
	CustomWallpaperFilename  = "custom.b64img"
	CustomWallpaperEmbedPath = "assets/wallpapers/custom.b64img"

	// Publishing: Doodads-embedded-within-levels.
	EmbeddedDoodadsBasePath   = "assets/doodads/"
	EmbeddedWallpaperBasePath = "assets/wallpapers/"

	// File formats: save new levels and doodads gzip compressed
	DrawingFormat = FormatZipfile

	// Zipfile drawings: max size of the LRU cache for loading chunks from
	// a zip file. Normally the chunker discards chunks not loaded in a
	// recent tick, but when iterating the full level this limits the max
	// size of loaded chunks before some will be freed to make room.
	// 0 = do not cap the cache.
	ChunkerLRUCacheMax = 0

	// Play Mode Touchscreen controls.
	PlayModeIdleTimeout = 2200 * time.Millisecond
	PlayModeAlphaStep   = 8 // 0-255 alpha, steps per tick for fade in
	PlayModeAlphaMax    = 220

	// Invulnerability time in seconds at respawn from checkpoint, in case
	// enemies are spawn camping.
	RespawnGodModeTimer = 3 * time.Second

	// GameController thresholds.
	GameControllerMouseMoveMax float64 = 20  // Max pixels per tick to simulate mouse movement.
	GameControllerScrollMin    float64 = 0.3 // Minimum threshold for a right-stick scroll event.

	// Limits on the Flood Fill tool so it doesn't run away on us.
	FloodToolVoidLimit = 600  // If clicking the void, +- 1000 px limit
	FloodToolLimit     = 1200 // If clicking a valid color on the level

	// Eager render level chunks to images during the load screen.
	// Originally chunks rendered to image and SDL texture on-demand, the loadscreen was
	// added to eager load (to image) the whole entire level at once (SDL textures were
	// still on demand, as they scroll into screen). Control this in-game with
	// `boolProp eager-render false` and the loadscreen will go quicker cuz it won't
	// load the whole entire level. Maybe useful to explore memory issues.
	EagerRenderLevelChunks = true

	// Number of chunks margin outside the Canvas Viewport for the LoadingViewport.
	LoadingViewportMarginChunks              = render.NewPoint(10, 8) // hoz, vert
	CanvasLoadUnloadModuloTicks       uint64 = 4
	CanvasChunkFreeChoppingBlockTicks uint64 = 128 // number of ticks old a chunk is to free it

	// For bounded levels, the game will try and keep actors inside the boundaries. But
	// in case e.g. the player is teleported far out of the boundaries (going thru a warp
	// door into an interior room "off the map"), allow them to be out of bounds. This
	// variable is the tolerance offset - if they are only this far out of bounds, put them
	// back in bounds but further out and they're OK.
	OutOfBoundsMargin = 40
)

// Edit Mode Values
var (
	// Number of Doodads per row in the palette.
	UIDoodadsPerRow = 2

	// Size of the DoodadButtons on actor canvas mouseover.
	UICanvasDoodadButtonSize = 16

	// Threshold for "very small doodad" where the buttons take up too big a proportion
	// and the doodad can't drag/drop easily.. tiny doodads will show the DoodadButtons
	// 50% off the top/right edge.
	UICanvasDoodadButtonSpaceNeeded = 20
)
