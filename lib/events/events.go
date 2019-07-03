// Package events manages mouse and keyboard SDL events for Doodle.
package events

import (
	"strings"
)

// State keeps track of event states.
type State struct {
	// Mouse buttons.
	Button1 *BoolTick // left
	Button2 *BoolTick // right
	Button3 *BoolTick // middle

	EscapeKey     *BoolTick
	EnterKey      *BoolTick
	ShiftActive   *BoolTick
	ControlActive *BoolTick
	KeyName       *StringTick
	Up            *BoolTick
	Left          *BoolTick
	Right         *BoolTick
	Down          *BoolTick

	// Cursor positions.
	CursorX *Int32Tick
	CursorY *Int32Tick

	// Window events: window has changed size.
	Resized *BoolTick
}

// New creates a new event state manager.
func New() *State {
	return &State{
		Button1:       &BoolTick{},
		Button2:       &BoolTick{},
		Button3:       &BoolTick{},
		EscapeKey:     &BoolTick{},
		EnterKey:      &BoolTick{},
		ShiftActive:   &BoolTick{},
		ControlActive: &BoolTick{},
		KeyName:       &StringTick{},
		Up:            &BoolTick{},
		Left:          &BoolTick{},
		Right:         &BoolTick{},
		Down:          &BoolTick{},
		CursorX:       &Int32Tick{},
		CursorY:       &Int32Tick{},
		Resized:       &BoolTick{},
	}
}

// ReadKey returns the normalized key symbol being pressed,
// taking the Shift key into account. QWERTY keyboard only, probably.
func (ev *State) ReadKey() string {
	if key := ev.KeyName.Read(); key != "" {
		if ev.ShiftActive.Pressed() {
			if symbol, ok := shiftMap[key]; ok {
				return symbol
			}
			return strings.ToUpper(key)
		}
		return key
	}
	return ""
}

// shiftMap maps keys to their Shift versions.
var shiftMap = map[string]string{
	"`": "~",
	"1": "!",
	"2": "@",
	"3": "#",
	"4": "$",
	"5": "%",
	"6": "^",
	"7": "&",
	"8": "*",
	"9": "(",
	"0": ")",
	"-": "_",
	"=": "+",
	"[": "{",
	"]": "}",
	`\`: "|",
	";": ":",
	`'`: `"`,
	",": "<",
	".": ">",
	"/": "?",
}
