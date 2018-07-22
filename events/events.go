// Package events manages mouse and keyboard SDL events for Doodle.
package events

// State keeps track of event states.
type State struct {
	// Mouse buttons.
	Button1 *BoolTick
	Button2 *BoolTick

	// Screenshot key.
	ScreenshotKey *BoolTick
	EscapeKey     *BoolTick
	Up            *BoolTick
	Left          *BoolTick
	Right         *BoolTick
	Down          *BoolTick

	// Cursor positions.
	CursorX *Int32Tick
	CursorY *Int32Tick
}

// New creates a new event state manager.
func New() *State {
	return &State{
		Button1:       &BoolTick{},
		Button2:       &BoolTick{},
		ScreenshotKey: &BoolTick{},
		EscapeKey:     &BoolTick{},
		Up:            &BoolTick{},
		Left:          &BoolTick{},
		Right:         &BoolTick{},
		Down:          &BoolTick{},
		CursorX:       &Int32Tick{},
		CursorY:       &Int32Tick{},
	}
}
