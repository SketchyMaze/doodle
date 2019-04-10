package balance

// Numbers.
var (
	// Window dimensions.
	Width  = 1024
	Height = 768

	// Speed to scroll a canvas with arrow keys in Edit Mode.
	CanvasScrollSpeed int32 = 8

	// Window scrolling behavior in Play Mode.
	ScrollboxHoz      = 64 // horizontal px from window border to start scrol
	ScrollboxVert     = 128
	ScrollMaxVelocity = 24

	// Player speeds
	PlayerMaxVelocity  = 12
	PlayerAcceleration = 2
	Gravity            = 2

	// Default chunk size for canvases.
	ChunkSize = 128

	// Default size for a new Doodad.
	DoodadSize = 100
)

// Edit Mode Values
var (
	// Number of Doodads per row in the palette.
	UIDoodadsPerRow = 2
)
