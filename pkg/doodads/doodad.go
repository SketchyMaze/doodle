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
		Hitbox:  render.NewRect(size, size),
		Layers: []Layer{
			{
				Name:    "main",
				Chunker: level.NewChunker(size),
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
		chunker = level.NewChunker(d.ChunkSize())
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
	return d.Layers[0].Chunker.Size
}

// Rect returns a rect of the ChunkSize for scaling a Canvas widget.
func (d *Doodad) Rect() render.Rect {
	var size = d.ChunkSize()
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
