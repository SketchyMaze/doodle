package balance

// Numbers.
var (
	// Window dimensions.
	Width  = 1024
	Height = 768

	// Speed to scroll a canvas with arrow keys in Edit Mode.
	CanvasScrollSpeed = 8

	// Window scrolling behavior in Play Mode.
	ScrollboxHoz      = 256 // horizontal px from window border to start scrol
	ScrollboxVert     = 128
	ScrollMaxVelocity = 8 // 24

	// Player speeds
	PlayerMaxVelocity  = 8
	PlayerAcceleration = 2
	Gravity            = 2

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
)

// Edit Mode Values
var (
	// Number of Doodads per row in the palette.
	UIDoodadsPerRow = 2
)
