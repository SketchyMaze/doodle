package level

import (
	"encoding/json"
	"fmt"

	"git.kirsle.net/apps/doodle/balance"
	"git.kirsle.net/apps/doodle/render"
)

// Level is the container format for Doodle map drawings.
type Level struct {
	Version     int    `json:"version"`     // File format version spec.
	GameVersion string `json:"gameVersion"` // Game version that created the level.
	Title       string `json:"title"`
	Author      string `json:"author"`
	Password    string `json:"passwd"`
	Locked      bool   `json:"locked"`

	// Chunked pixel data.
	Chunker *Chunker `json:"chunks"`

	// XXX: deprecated?
	Width  int32 `json:"w"`
	Height int32 `json:"h"`

	// The Palette holds the unique "colors" used in this map file, and their
	// properties (solid, fire, slippery, etc.)
	Palette *Palette `json:"palette"`

	// Pixels is a 2D array indexed by [X][Y]. The cell values are indexes into
	// the Palette.
	Pixels []*Pixel `json:"pixels"`
}

// New creates a blank level object with all its members initialized.
func New() *Level {
	return &Level{
		Version: 1,
		Chunker: NewChunker(balance.ChunkSize),
		Pixels:  []*Pixel{},
		Palette: &Palette{},
	}
}

// Pixel associates a coordinate with a palette index.
type Pixel struct {
	X            int32 `json:"x"`
	Y            int32 `json:"y"`
	PaletteIndex int32 `json:"p"`

	// Private runtime values.
	Palette *Palette `json:"-"` // pointer to its palette, TODO: needed?
	Swatch  *Swatch  `json:"-"` // pointer to its swatch, for when rendered.
}

func (p Pixel) String() string {
	return fmt.Sprintf("Pixel<%s '%s' (%d,%d)>", p.Swatch.Color, p.Swatch.Name, p.X, p.Y)
}

// Point returns the pixel's point.
func (p Pixel) Point() render.Point {
	return render.Point{
		X: p.X,
		Y: p.Y,
	}
}

// MarshalJSON serializes a Pixel compactly as a simple list.
func (p Pixel) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(
		`[%d, %d, %d]`,
		p.X, p.Y, p.PaletteIndex,
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
	if len(triplet) > 2 {
		p.PaletteIndex = triplet[2]
	}
	return nil
}
