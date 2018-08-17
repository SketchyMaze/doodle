package level

import (
	"fmt"

	"git.kirsle.net/apps/doodle/render"
)

// DefaultPalette returns a sensible default palette.
func DefaultPalette() *Palette {
	return &Palette{
		Swatches: []*Swatch{
			&Swatch{
				Name:  "solid",
				Color: render.Black,
				Solid: true,
			},
			&Swatch{
				Name:  "decoration",
				Color: render.Grey,
			},
			&Swatch{
				Name:  "fire",
				Color: render.Red,
				Fire:  true,
			},
			&Swatch{
				Name:  "water",
				Color: render.Blue,
				Water: true,
			},
		},
	}
}

// NewPalette initializes a blank palette.
func NewPalette() *Palette {
	return &Palette{
		Swatches: []*Swatch{},
		byName:   map[string]int{},
	}
}

// Palette holds an index of colors used in a drawing.
type Palette struct {
	Swatches []*Swatch `json:"swatches"`

	// Private runtime values
	ActiveSwatch *Swatch        `json:"-"` // name of the actively selected color
	byName       map[string]int // Cache map of swatches by name
}

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
}

func (s Swatch) String() string {
	return s.Name
}

// Index returns the Swatch's position in the palette.
func (s Swatch) Index() int {
	fmt.Printf("%+v index: %d", s, s.index)
	return s.index
}

// Inflate the palette swatch caches. Always call this method after you have
// initialized the palette (i.e. loaded it from JSON); this will update the
// "color by name" cache and assign the index numbers to each swatch.
func (p *Palette) Inflate() {
	p.update()
}

// Get a swatch by name.
func (p *Palette) Get(name string) (result *Swatch, exists bool) {
	p.update()

	if index, ok := p.byName[name]; ok && index < len(p.Swatches) {
		result = p.Swatches[index]
		exists = true
	}

	return
}

// update the internal caches and such.
func (p *Palette) update() {
	// Initialize the name cache if nil or if the size disagrees with the
	// length of the swatches available.
	if p.byName == nil || len(p.byName) != len(p.Swatches) {
		// Initialize the name cache.
		p.byName = map[string]int{}
		for i, swatch := range p.Swatches {
			swatch.index = i
			p.byName[swatch.Name] = i
		}
	}
}
