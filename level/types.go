package level

import (
	"encoding/json"
	"fmt"
)

// Level is the container format for Doodle map drawings.
type Level struct {
	Version     int32  `json:"version"`     // File format version spec.
	GameVersion string `json:"gameVersion"` // Game version that created the level.
	Title       string `json:"title"`
	Author      string `json:"author"`
	Password    string `json:"passwd"`
	Locked      bool   `json:"locked"`

	// Level size.
	Width  int32 `json:"w"`
	Height int32 `json:"h"`

	// The Palette holds the unique "colors" used in this map file, and their
	// properties (solid, fire, slippery, etc.)
	Palette []Palette `json:"palette"`

	// Pixels is a 2D array indexed by [X][Y]. The cell values are indexes into
	// the Palette.
	Pixels []Pixel `json:"pixels"`
}

// Pixel associates a coordinate with a palette index.
type Pixel struct {
	X       int32 `json:"x"`
	Y       int32 `json:"y"`
	Palette int32 `json:"p"`
}

// MarshalJSON serializes a Pixel compactly as a simple list.
func (p Pixel) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(
		`[%d, %d, %d]`,
		p.X, p.Y, p.Palette,
	)), nil
}

// UnmarshalJSON loads a Pixel from JSON again.
func (p *Pixel) UnmarshalJSON(text []byte) error {
	var triplet []int32
	err := json.Unmarshal(text, &triplet)
	if err != nil {
		return err
	}

	p.X = triplet[0]
	p.Y = triplet[1]
	p.Palette = triplet[2]
	return nil
}

// Palette are the unique pixel attributes that this map uses, and serves
// as a lookup table for the Pixels.
type Palette struct {
	// Required attributes.
	Color string `json:"color"`

	// Optional attributes.
	Solid bool `json:"solid,omitempty"`
	Fire  bool `json:"fire,omitempty"`
}
