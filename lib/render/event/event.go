package event

import "strings"

// State holds the current state of key/mouse events.
type State struct {
	// Mouse buttons.
	Button1 bool // Left
	Button2 bool // Middle
	Button3 bool // Right

	// Special keys
	Escape bool
	Enter  bool
	Shift  bool
	Ctrl   bool
	Alt    bool
	Up     bool
	Left   bool
	Right  bool
	Down   bool

	// Pressed keys.
	keydown map[string]interface{}

	// Cursor position
	CursorX int
	CursorY int

	// Window resized
	WindowResized bool
}

// NewState creates a new event.State.
func NewState() *State {
	return &State{
		keydown: map[string]interface{}{},
	}
}

// SetKeyDown sets that a named key is pressed down.
func (s *State) SetKeyDown(name string, down bool) {
	if down {
		s.keydown[name] = nil
	} else {
		delete(s.keydown, name)
	}
}

// KeyDown returns whether a named key is currently pressed.
func (s *State) KeyDown(name string) bool {
	_, ok := s.keydown[name]
	return ok
}

// KeysDown returns a list of all key names currently pressed down.
// Set shifted to True to return the key symbols correctly shifted
// (uppercase, or symbols on number keys, etc.)
func (s *State) KeysDown(shifted bool) []string {
	var (
		result = make([]string, len(s.keydown))
		i      = 0
	)

	for key := range s.keydown {
		if shifted && s.Shift {
			if symbol, ok := shiftMap[key]; ok {
				result[i] = symbol
			} else {
				result[i] = strings.ToUpper(key)
			}
		} else {
			result[i] = key
		}

		i++
	}
	return result
}

// ResetKeyDown clears all key-down states.
func (s *State) ResetKeyDown() {
	s.keydown = map[string]interface{}{}
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
