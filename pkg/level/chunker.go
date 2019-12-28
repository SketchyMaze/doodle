package level

import (
	"encoding/json"
	"fmt"
	"math"

	"git.kirsle.net/apps/doodle/pkg/log"
	"git.kirsle.net/go/render"
	"github.com/vmihailenco/msgpack"
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
		chunk.Point = coord
		chunk.Size = c.Size
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

						// Verify this pixel is also in range.
						if px.Point().Inside(viewport) {
							pipe <- px
						}
					}
				}
			}
		}
		close(pipe)
	}()
	return pipe
}

// IterViewportChunks returns a channel to iterate over the Chunk objects that
// appear within the viewport rect, instead of the pixels in each chunk.
func (c *Chunker) IterViewportChunks(viewport render.Rect) <-chan render.Point {
	pipe := make(chan render.Point)
	go func() {
		sent := make(map[render.Point]interface{})

		for x := viewport.X; x < viewport.W; x += (c.Size / 4) {
			for y := viewport.Y; y < viewport.H; y += (c.Size / 4) {

				// Constrain this chunksize step to a point within the bounds
				// of the viewport. This can yield partial chunks on the edges
				// of the viewport.
				point := render.NewPoint(x, y)
				if point.X < viewport.X {
					point.X = viewport.X
				} else if point.X > viewport.X+viewport.W {
					point.X = viewport.X + viewport.W
				}
				if point.Y < viewport.Y {
					point.Y = viewport.Y
				} else if point.Y > viewport.Y+viewport.H {
					point.Y = viewport.Y + viewport.H
				}

				// Translate to a chunk coordinate, dedupe and send it.
				coord := c.ChunkCoordinate(render.NewPoint(x, y))
				if _, ok := sent[coord]; ok {
					continue
				}
				sent[coord] = nil

				if _, ok := c.GetChunk(coord); ok {
					pipe <- coord
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

// WorldSize returns the bounding coordinates that the Chunker has chunks to
// manage: the lowest pixels from the lowest chunks to the highest pixels of
// the highest chunks.
func (c *Chunker) WorldSize() render.Rect {
	// Lowest and highest chunks.
	var (
		chunkLowest  render.Point
		chunkHighest render.Point
		size         = c.Size
	)

	for coord := range c.Chunks {
		if coord.X < chunkLowest.X {
			chunkLowest.X = coord.X
		}
		if coord.Y < chunkLowest.Y {
			chunkLowest.Y = coord.Y
		}

		if coord.X > chunkHighest.X {
			chunkHighest.X = coord.X
		}
		if coord.Y > chunkHighest.Y {
			chunkHighest.Y = coord.Y
		}
	}

	return render.Rect{
		X: chunkLowest.X * size,
		Y: chunkLowest.Y * size,
		W: (chunkHighest.X * size) + (size - 1),
		H: (chunkHighest.Y * size) + (size - 1),
	}
}

// WorldSizePositive returns the WorldSize anchored to 0,0 with only positive
// coordinates.
func (c *Chunker) WorldSizePositive() render.Rect {
	S := c.WorldSize()
	return render.Rect{
		X: 0,
		Y: 0,
		W: int(math.Abs(float64(S.X))) + S.W,
		H: int(math.Abs(float64(S.Y))) + S.H,
	}
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
		chunk.Point = coord
		chunk.Size = c.Size
	}

	return chunk.Set(p, sw)
}

// SetRect sets a rectangle of pixels to a color all at once.
func (c *Chunker) SetRect(r render.Rect, sw *Swatch) error {
	var (
		xMin = r.X
		yMin = r.Y
		xMax = r.X + r.W
		yMax = r.Y + r.H
	)
	for x := xMin; x < xMax; x++ {
		for y := yMin; y < yMax; y++ {
			c.Set(render.NewPoint(x, y), sw)
		}
	}

	return nil
}

// Delete a pixel at the given coordinate.
func (c *Chunker) Delete(p render.Point) error {
	coord := c.ChunkCoordinate(p)
	defer c.pruneChunk(coord)

	if chunk, ok := c.Chunks[coord]; ok {
		return chunk.Delete(p)
	}
	return fmt.Errorf("no chunk %s exists for point %s", coord, p)
}

// DeleteRect deletes a rectangle of pixels between two points.
// The rect is a relative one with a width and height, and the X,Y values are
// an absolute world coordinate.
func (c *Chunker) DeleteRect(r render.Rect) error {
	var (
		xMin = r.X
		yMin = r.Y
		xMax = r.X + r.W
		yMax = r.Y + r.H
	)
	for x := xMin; x < xMax; x++ {
		for y := yMin; y < yMax; y++ {
			c.Delete(render.NewPoint(x, y))
		}
	}

	return nil
}

// pruneChunk will remove an empty chunk from the chunk map, called after
// delete operations.
func (c *Chunker) pruneChunk(coord render.Point) {
	if chunk, ok := c.Chunks[coord]; ok {
		if chunk.Len() == 0 {
			log.Info("Chunker.pruneChunk: %s has become empty", coord)
			delete(c.Chunks, coord)
		}
	}
}

// ChunkCoordinate computes a chunk coordinate from an absolute coordinate.
func (c *Chunker) ChunkCoordinate(abs render.Point) render.Point {
	if c.Size == 0 {
		return render.Point{}
	}

	size := float64(c.Size)
	return render.NewPoint(
		int(math.Floor(float64(abs.X)/size)),
		int(math.Floor(float64(abs.Y)/size)),
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

// MarshalMsgpack to convert the chunk map to binary.
func (c ChunkMap) MarshalMsgpack() ([]byte, error) {
	dict := map[string]*Chunk{}
	for point, chunk := range c {
		dict[point.String()] = chunk
	}

	out, err := msgpack.Marshal(dict)
	return out, err
}
