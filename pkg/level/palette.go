package level

import "git.kirsle.net/go/render"

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
				Color: render.RGBA(0, 0, 255, 180),
				Water: true,
			},
		},
	}
}

// NewBlueprintPalette returns the blueprint theme's color palette.
func NewBlueprintPalette() *Palette {
	return &Palette{
		Swatches: []*Swatch{
			&Swatch{
				Name:  "solid",
				Color: render.RGBA(254, 254, 254, 255),
				Solid: true,
			},
			&Swatch{
				Name:  "decoration",
				Color: render.Grey,
			},
			&Swatch{
				Name:  "fire",
				Color: render.RGBA(255, 80, 0, 255),
				Fire:  true,
			},
			&Swatch{
				Name:  "water",
				Color: render.RGBA(0, 153, 255, 255),
				Water: true,
			},
			&Swatch{
				Name:  "electric",
				Color: render.RGBA(255, 255, 0, 255),
				Solid: true,
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
