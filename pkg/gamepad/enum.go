package gamepad

// Enums and constants.

// Mode of controller behavior (Player One)
type Mode int

// Style of controller (Player One)
type Style int

// Controller mode options.
const (
	// MouseMode: the joystick moves a mouse cursor around and
	// the face buttons emulate mouse click events.
	MouseMode Mode = iota

	// GameplayMode: to control the player character during Play Mode.
	GameplayMode

	// EditorMode: to support the Level Editor.
	EditorMode
)

// Controller style options.
const (
	XStyle      Style = iota // Xbox 360 layout (A button on bottom)
	NStyle                   // Nintendo style (A button on right)
	CustomStyle              // Custom style (TODO)
)
