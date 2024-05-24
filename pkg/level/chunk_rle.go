package level

import (
	"git.kirsle.net/SketchyMaze/doodle/pkg/level/rle"
	"git.kirsle.net/go/render"
)

// RLEAccessor implements a chunk accessor which stores its on-disk format using
// Run Length Encoding (RLE), but in memory behaves equivalently to the MapAccessor.
type RLEAccessor struct {
	chunk *Chunk // parent Chunk, for its Size and Point
	acc   *MapAccessor
}

// NewRLEAccessor initializes a RLEAccessor.
func NewRLEAccessor(chunk *Chunk) *RLEAccessor {
	return &RLEAccessor{
		chunk: chunk,
		acc:   NewMapAccessor(chunk),
	}
}

// Inflate the sparse swatches from their palette indexes.
func (a *RLEAccessor) Inflate(pal *Palette) error {
	return a.acc.Inflate(pal)
}

// Len returns the current size of the map, or number of pixels registered.
func (a *RLEAccessor) Len() int {
	return a.acc.Len()
}

// IterViewport returns a channel to loop over pixels in the viewport.
func (a *RLEAccessor) IterViewport(viewport render.Rect) <-chan Pixel {
	return a.acc.IterViewport(viewport)
}

// Iter returns a channel to loop over all points in this chunk.
func (a *RLEAccessor) Iter() <-chan Pixel {
	return a.acc.Iter()
}

// Get a pixel from the map.
func (a *RLEAccessor) Get(p render.Point) (*Swatch, error) {
	return a.acc.Get(p)
}

// Set a pixel on the map.
func (a *RLEAccessor) Set(p render.Point, sw *Swatch) error {
	return a.acc.Set(p, sw)
}

// Delete a pixel from the map.
func (a *RLEAccessor) Delete(p render.Point) error {
	return a.acc.Delete(p)
}

/*
MarshalBinary converts the chunk data to a binary representation.

This accessor uses Run Length Encoding (RLE) in its binary format. Starting
with the top-left pixel of this chunk, the binary format is a stream of bytes
formatted as such:

- UVarint for the palette index number (0-255), with 0xFFFF meaning void
- UVarint for the length of repetition of that palette index
*/
func (a *RLEAccessor) MarshalBinary() ([]byte, error) {
	// Flatten the chunk out into a full 2D array of all its points.
	var (
		size      = int(a.chunk.Size)
		grid, err = rle.NewGrid(size)
	)
	if err != nil {
		return nil, err
	}

	// Populate the dense 2D array of its pixels.
	for y, row := range grid {
		for x := range row {
			var (
				relative    = render.NewPoint(x, y)
				absolute    = FromRelativeCoordinate(relative, a.chunk.Point, a.chunk.Size)
				swatch, err = a.Get(absolute)
			)

			if err != nil {
				continue
			}

			var ptr = uint64(swatch.Index())
			grid[relative.Y][relative.X] = &ptr
		}
	}

	return grid.Compress()
}

// UnmarshalBinary will decode a compressed RLEAccessor byte stream.
func (a *RLEAccessor) UnmarshalBinary(compressed []byte) error {
	a.acc.mu.Lock()
	defer a.acc.mu.Unlock()

	// New format: decompress the byte stream.
	// log.Debug("RLEAccessor.Unmarshal: Reading %d bytes of compressed chunk data", len(compressed))

	grid, err := rle.NewGrid(int(a.chunk.Size))
	if err != nil {
		return err
	}

	if err := grid.Decompress(compressed); err != nil {
		return err
	}

	// Load the grid into our MapAccessor.
	a.acc.Reset()
	for y, row := range grid {
		for x, col := range row {
			if col == nil {
				continue
			}

			abs := FromRelativeCoordinate(render.NewPoint(x, y), a.chunk.Point, a.chunk.Size)
			a.acc.grid[abs] = NewSparseSwatch(int(*col))
		}
	}

	return nil
}
