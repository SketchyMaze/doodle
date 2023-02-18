package doodads

import (
	"git.kirsle.net/SketchyMaze/doodle/pkg/balance"
	"git.kirsle.net/SketchyMaze/doodle/pkg/drawtool"
	"git.kirsle.net/SketchyMaze/doodle/pkg/level"
	"git.kirsle.net/SketchyMaze/doodle/pkg/log"
	"git.kirsle.net/go/render"
)

// Doodad is a reusable component for Levels that have scripts and graphics.
type Doodad struct {
	level.Base
	Filename string             `json:"-"` // used internally, not saved in json
	Hidden   bool               `json:"hidden,omitempty"`
	Palette  *level.Palette     `json:"palette"`
	Size     render.Rect        `json:"size"` // doodad dimensions
	Script   string             `json:"script"`
	Hitbox   render.Rect        `json:"hitbox"`
	Layers   []Layer            `json:"layers"`
	Tags     map[string]string  `json:"data"`    // arbitrary key/value data storage
	Options  map[string]*Option `json:"options"` // runtime options for a doodad

	// Undo history, temporary live data not persisted to the level file.
	UndoHistory *drawtool.History `json:"-"`
}

// Layer holds a layer of drawing data for a Doodad.
type Layer struct {
	Name    string         `json:"name"`
	Chunker *level.Chunker `json:"chunks"`
}

/*
New creates a new Doodad.

You can give it one or two values for dimensions:

- New(size int) creates a square doodad (classic)
- New(width, height int) lets you have a different width x height.
*/
func New(dimensions ...int) *Doodad {
	var (
		// Defaults
		size      int
		chunkSize uint8
		width     int
		height    int
	)

	switch len(dimensions) {
	case 1:
		width, height = dimensions[0], dimensions[0]
	case 2:
		width, height = dimensions[0], dimensions[1]
	}

	// Set the desired chunkSize to be the largest dimension.
	if width > height {
		size = width
	} else {
		size = height
	}

	// If no size at all, fall back on the default.
	if size == 0 {
		size = int(balance.ChunkSize)
	}

	// Pick an optimal chunk size - if our doodad size
	// is under 256 use only one chunk per layer by matching
	// that size.
	if size <= 255 {
		chunkSize = uint8(size)
	}

	return &Doodad{
		Base: level.Base{
			Version: 1,
		},
		Palette: level.DefaultPalette(),
		Hitbox:  render.NewRect(width, height),
		Size:    render.NewRect(width, height),
		Layers: []Layer{
			{
				Name:    "main",
				Chunker: level.NewChunker(chunkSize),
			},
		},
		Tags:        map[string]string{},
		Options:     map[string]*Option{},
		UndoHistory: drawtool.NewHistory(balance.UndoHistory),
	}
}

// AddLayer adds a new layer to the doodad. Call this rather than appending
// your own layer so it points the Zipfile and layer number in. The chunker
// is optional - pass nil and a new blank chunker is created.
func (d *Doodad) AddLayer(name string, chunker *level.Chunker) Layer {
	if chunker == nil {
		chunker = level.NewChunker(d.ChunkSize8())
	}

	layer := Layer{
		Name:    name,
		Chunker: chunker,
	}
	layer.Chunker.Layer = len(d.Layers)
	d.Layers = append(d.Layers, layer)
	d.Inflate()

	return layer
}

// Teardown cleans up texture cache memory when the doodad is no longer needed by the game.
func (d *Doodad) Teardown() {
	var (
		chunks   int
		textures int
	)

	for _, layer := range d.Layers {
		for coord := range layer.Chunker.IterChunks() {
			if chunk, ok := layer.Chunker.GetChunk(coord); ok {
				freed := chunk.Teardown()
				chunks++
				textures += freed
			}
		}
	}

	// Debug log if any textures were actually freed.
	if textures > 0 {
		log.Debug("Teardown doodad (%s): Freed %d textures across %d chunks", d.Title, textures, chunks)
	}
}

// Tag gets a value from the doodad's tags.
func (d *Doodad) Tag(name string) string {
	if v, ok := d.Tags[name]; ok {
		return v
	}
	log.Warn("Doodad(%s).Tag(%s): tag not defined", d.Title, name)
	return ""
}

// ChunkSize returns the chunk size of the Doodad's first layer.
func (d *Doodad) ChunkSize() int {
	return int(d.Layers[0].Chunker.Size)
}

// ChunkSize8 returns the chunk size of the Doodad's first layer as its actual uint8 value.
func (d *Doodad) ChunkSize8() uint8 {
	return d.Layers[0].Chunker.Size
}

// Rect returns a rect of the ChunkSize for scaling a Canvas widget.
func (d *Doodad) Rect() render.Rect {
	var size = int(d.ChunkSize())
	return render.Rect{
		W: size,
		H: size,
	}
}

// Inflate attaches the pixels to their swatches after loading from disk.
func (d *Doodad) Inflate() {
	d.Palette.Inflate()
	for i, layer := range d.Layers {
		layer.Chunker.Layer = i
		layer.Chunker.Inflate(d.Palette)
	}
}
