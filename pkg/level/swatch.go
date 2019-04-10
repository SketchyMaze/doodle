package level

import (
	"fmt"

	"git.kirsle.net/apps/doodle/lib/render"
)

// Swatch holds details about a single value in the palette.
type Swatch struct {
	Name  string       `json:"name"`
	Color render.Color `json:"color"`

	// Optional attributes.
	Solid bool `json:"solid,omitempty"`
	Fire  bool `json:"fire,omitempty"`
	Water bool `json:"water,omitempty"`

	// Private runtime attributes.
	index int // position in the Palette, for reverse of `Palette.byName`

	// When the swatch is loaded from JSON we only get the index number, and
	// need to expand out the swatch later when the palette is loaded.
	paletteIndex int
	isSparse     bool
}

// NewSparseSwatch creates a sparse Swatch from a palette index that will need
// later expansion, when loading drawings from disk.
func NewSparseSwatch(paletteIndex int) *Swatch {
	return &Swatch{
		isSparse:     true,
		paletteIndex: paletteIndex,
	}
}

func (s Swatch) String() string {
	if s.isSparse {
		return fmt.Sprintf("Swatch<sparse:%d>", s.paletteIndex)
	}
	if s.Name == "" {
		return s.Color.String()
	}
	return s.Name
}

// IsSparse returns whether this Swatch is sparse (has only a palette index) and
// requires inflation.
func (s *Swatch) IsSparse() bool {
	return s.isSparse
}

// Index returns the Swatch's position in the palette.
func (s *Swatch) Index() int {
	return s.index
}
