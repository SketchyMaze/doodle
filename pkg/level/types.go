package level

import (
	"encoding/json"
	"fmt"

	"git.kirsle.net/apps/doodle/lib/render"
	"git.kirsle.net/apps/doodle/pkg/balance"
)

// Useful variables.
var (
	DefaultWallpaper = "notebook.png"
)

// Base provides the common struct keys that are shared between Levels and
// Doodads.
type Base struct {
	Version     int    `json:"version" msgpack:"0"`     // File format version spec.
	GameVersion string `json:"gameVersion" msgpack:"1"` // Game version that created the level.
	Title       string `json:"title" msgpack:"2"`
	Author      string `json:"author" msgpack:"3"`

	// Every drawing type is able to embed other files inside of itself.
	Files FileSystem `json:"files" msgpack:"4"`
}

// Level is the container format for Doodle map drawings.
type Level struct {
	Base
	Password string `json:"passwd" msgpack:"10"`
	Locked   bool   `json:"locked" msgpack:"11"`

	// Chunked pixel data.
	Chunker *Chunker `json:"chunks" msgpack:"12"`

	// The Palette holds the unique "colors" used in this map file, and their
	// properties (solid, fire, slippery, etc.)
	Palette *Palette `json:"palette" msgpack:"13"`

	// Page boundaries and wallpaper settings.
	PageType  PageType `json:"pageType" msgpack:"14"`
	MaxWidth  int64    `json:"boundedWidth" msgpack:"15"` // only if bounded or bordered
	MaxHeight int64    `json:"boundedHeight" msgpack:"16"`
	Wallpaper string   `json:"wallpaper" msgpack:"17"`

	// Actors keep a list of the doodad instances in this map.
	Actors ActorMap `json:"actors" msgpack:"18"`
}

// New creates a blank level object with all its members initialized.
func New() *Level {
	return &Level{
		Base: Base{
			Version: 1,
		},
		Chunker: NewChunker(balance.ChunkSize),
		Palette: &Palette{},
		Actors:  ActorMap{},

		PageType:  NoNegativeSpace,
		Wallpaper: DefaultWallpaper,
		MaxWidth:  2550,
		MaxHeight: 3300,
	}
}

// Pixel associates a coordinate with a palette index.
type Pixel struct {
	X            int32 `json:"x"`
	Y            int32 `json:"y"`
	PaletteIndex int32 `json:"p"`

	// Private runtime values.
	Swatch *Swatch `json:"-"` // pointer to its swatch, for when rendered.
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
