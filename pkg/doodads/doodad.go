package doodads

import (
	"git.kirsle.net/go/render"
	"git.kirsle.net/apps/doodle/pkg/balance"
	"git.kirsle.net/apps/doodle/pkg/level"
)

// Doodad is a reusable component for Levels that have scripts and graphics.
type Doodad struct {
	level.Base
	Filename string            `json:"-"` // used internally, not saved in json
	Hidden   bool              `json:"hidden,omitempty"`
	Palette  *level.Palette    `json:"palette"`
	Script   string            `json:"script"`
	Layers   []Layer           `json:"layers"`
	Tags     map[string]string `json:"data"` // arbitrary key/value data storage
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
		Tags: map[string]string{},
	}
}

// ChunkSize returns the chunk size of the Doodad's first layer.
func (d *Doodad) ChunkSize() int {
	return d.Layers[0].Chunker.Size
}

// Rect returns a rect of the ChunkSize for scaling a Canvas widget.
func (d *Doodad) Rect() render.Rect {
	var size = d.ChunkSize()
	return render.Rect{
		W: int32(size),
		H: int32(size),
	}
}

// Inflate attaches the pixels to their swatches after loading from disk.
func (d *Doodad) Inflate() {
	d.Palette.Inflate()
	for _, layer := range d.Layers {
		layer.Chunker.Inflate(d.Palette)
	}
}
