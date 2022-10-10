package level

import (
	"fmt"
	"strings"

	"git.kirsle.net/go/render"
)

// Swatch holds details about a single value in the palette.
type Swatch struct {
	Name    string       `json:"name"`
	Color   render.Color `json:"color"`
	Pattern string       `json:"pattern"` // like "noise.png"

	// Optional attributes.
	Solid     bool `json:"solid,omitempty"`
	SemiSolid bool `json:"semisolid,omitempty"`
	Fire      bool `json:"fire,omitempty"`
	Water     bool `json:"water,omitempty"`
	Slippery  bool `json:"slippery,omitempty"`

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

// Attributes returns a comma-separated list of attributes as a string on
// this swatch. This is for debugging and the `doodad show` CLI command to
// summarize the swatch and shouldn't be used for game logic.
func (s *Swatch) Attributes() string {
	var result string

	if s.Solid {
		result += "solid,"
	}
	if s.SemiSolid {
		result += "semi-solid,"
	}
	if s.Fire {
		result += "fire,"
	}
	if s.Water {
		result += "water,"
	}
	if s.isSparse {
		result += "sparse,"
	}
	if s.Slippery {
		result += "slippery,"
	}

	if result == "" {
		result = "none,"
	}

	return strings.TrimSuffix(result, ",")
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
