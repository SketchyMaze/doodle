package level

import (
	"encoding/json"
	"fmt"
	"math"

	"git.kirsle.net/apps/doodle/render"
)

// Chunker is the data structure that manages the chunks of a level, and
// provides the API to interact with the pixels using their absolute coordinates
// while abstracting away the underlying details.
type Chunker struct {
	Size   int      `json:"size"`
	Chunks ChunkMap `json:"chunks"`
}

// NewChunker creates a new chunk manager with a given chunk size.
func NewChunker(size int) *Chunker {
	return &Chunker{
		Size:   size,
		Chunks: ChunkMap{},
	}
}

// Inflate iterates over the pixels in the (loaded) chunks and expands any
// Sparse Swatches (which have only their palette index, from the file format
// on disk) to connect references to the swatches in the palette.
func (c *Chunker) Inflate(pal *Palette) error {
	for coord, chunk := range c.Chunks {
		log.Debug("Chunker.Inflate: expanding chunk %s %+v", coord, chunk)
		chunk.Inflate(pal)
	}
	return nil
}

// IterViewport returns a channel to iterate every point that exists within
// the viewport rect.
func (c *Chunker) IterViewport(viewport render.Rect) <-chan Pixel {
	pipe := make(chan Pixel)
	go func() {
		// Get the chunk box coordinates.
		var (
			topLeft     = c.ChunkCoordinate(render.NewPoint(viewport.X, viewport.Y))
			bottomRight = c.ChunkCoordinate(render.Point{
				X: viewport.X + viewport.W,
				Y: viewport.Y + viewport.H,
			})
		)
		for cx := topLeft.X; cx <= bottomRight.X; cx++ {
			for cy := topLeft.Y; cy <= bottomRight.Y; cy++ {
				if chunk, ok := c.GetChunk(render.NewPoint(cx, cy)); ok {
					for px := range chunk.Iter() {
						pipe <- px
					}
				}
			}
		}
		close(pipe)
	}()
	return pipe
}

// IterPixels returns a channel to iterate over every pixel in the entire
// chunker.
func (c *Chunker) IterPixels() <-chan Pixel {
	pipe := make(chan Pixel)
	go func() {
		for _, chunk := range c.Chunks {
			for px := range chunk.Iter() {
				pipe <- px
			}
		}
		close(pipe)
	}()
	return pipe
}

// GetChunk gets a chunk at a certain position. Returns false if not found.
func (c *Chunker) GetChunk(p render.Point) (*Chunk, bool) {
	chunk, ok := c.Chunks[p]
	return chunk, ok
}

// Get a pixel at the given coordinate. Returns the Palette entry for that
// pixel or else returns an error if not found.
func (c *Chunker) Get(p render.Point) (*Swatch, error) {
	// Compute the chunk coordinate.
	coord := c.ChunkCoordinate(p)
	if chunk, ok := c.Chunks[coord]; ok {
		return chunk.Get(p)
	}
	return nil, fmt.Errorf("no chunk %s exists for point %s", coord, p)
}

// Set a pixel at the given coordinate.
func (c *Chunker) Set(p render.Point, sw *Swatch) error {
	coord := c.ChunkCoordinate(p)
	chunk, ok := c.Chunks[coord]
	if !ok {
		chunk = NewChunk()
		c.Chunks[coord] = chunk
	}

	return chunk.Set(p, sw)
}

// Delete a pixel at the given coordinate.
func (c *Chunker) Delete(p render.Point) error {
	coord := c.ChunkCoordinate(p)
	if chunk, ok := c.Chunks[coord]; ok {
		return chunk.Delete(p)
	}
	return fmt.Errorf("no chunk %s exists for point %s", coord, p)
}

// ChunkCoordinate computes a chunk coordinate from an absolute coordinate.
func (c *Chunker) ChunkCoordinate(abs render.Point) render.Point {
	if c.Size == 0 {
		return render.Point{}
	}

	size := float64(c.Size)
	return render.NewPoint(
		int32(math.Floor(float64(abs.X)/size)),
		int32(math.Floor(float64(abs.Y)/size)),
	)
}

// ChunkMap maps a chunk coordinate to its chunk data.
type ChunkMap map[render.Point]*Chunk

// MarshalJSON to convert the chunk map to JSON. This is needed for writing so
// the JSON encoder knows how to serializes a `map[Point]*Chunk` but the inverse
// is not necessary to implement.
func (c ChunkMap) MarshalJSON() ([]byte, error) {
	dict := map[string]*Chunk{}
	for point, chunk := range c {
		dict[point.String()] = chunk
	}

	out, err := json.Marshal(dict)
	return out, err
}
