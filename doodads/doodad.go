package doodads

import (
	"git.kirsle.net/apps/doodle/balance"
	"git.kirsle.net/apps/doodle/level"
)

// Doodad is a reusable component for Levels that have scripts and graphics.
type Doodad struct {
	level.Base
	Palette *level.Palette `json:"palette"`
	Script  string         `json:"script"`
	Layers  []Layer        `json:"layers"`
}

// Layer holds a layer of drawing data for a Doodad.
type Layer struct {
	Name    string         `json:"name"`
	Chunker *level.Chunker `json:"chunks"`
}

// New creates a new Doodad.
func New(size int) *Doodad {
	if size == 0 {
		size = balance.DoodadSize
	}

	return &Doodad{
		Base: level.Base{
			Version: 1,
		},
		Palette: level.DefaultPalette(),
		Layers: []Layer{
			{
				Name:    "main",
				Chunker: level.NewChunker(size),
			},
		},
	}
}

// Inflate attaches the pixels to their swatches after loading from disk.
func (d *Doodad) Inflate() {
	d.Palette.Inflate()
	for _, layer := range d.Layers {
		layer.Chunker.Inflate(d.Palette)
	}
}
