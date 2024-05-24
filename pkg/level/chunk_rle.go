package level

import (
	"bytes"
	"encoding/binary"
	"errors"

	"git.kirsle.net/SketchyMaze/doodle/pkg/log"
	"git.kirsle.net/go/render"
)

// RLEAccessor implements a chunk accessor which stores its on-disk format using
// Run Length Encoding (RLE), but in memory behaves equivalently to the MapAccessor.
type RLEAccessor struct {
	acc *MapAccessor
}

// NewRLEAccessor initializes a RLEAccessor.
func NewRLEAccessor() *RLEAccessor {
	return &RLEAccessor{
		acc: NewMapAccessor(),
	}
}

// SetChunkCoordinate receives our chunk's coordinate from the Chunker.
func (a *RLEAccessor) SetChunkCoordinate(p render.Point, size uint8) {
	a.acc.coord = p
	a.acc.size = size
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

// Make2DChunkGrid creates a 2D map of uint64 pointers matching the square dimensions of the given size.
//
// It is used by the RLEAccessor to flatten a chunk into a grid for run-length encoding.
func Make2DChunkGrid(size int) ([][]*uint64, error) {
	// Sanity check if the chunk was properly initialized.
	if size == 0 {
		return nil, errors.New("chunk not initialized correctly with its size and coordinate")
	}

	var grid = make([][]*uint64, size)
	for i := 0; i < size; i++ {
		grid[i] = make([]*uint64, size)
	}

	return grid, nil
}

/*
MarshalBinary converts the chunk data to a binary representation.

This accessor uses Run Length Encoding (RLE) in its binary format. Starting
with the top-left pixel of this chunk, the binary format is a stream of bytes
formatted as such:

- UVarint for the palette index number (0-255), with 0xFF meaning void
- UVarint for the length of repetition of that palette index
*/
func (a *RLEAccessor) MarshalBinary() ([]byte, error) {
	// Flatten the chunk out into a full 2D array of all its points.
	var (
		size      = int(a.acc.size)
		grid, err = Make2DChunkGrid(size)
	)
	if err != nil {
		return nil, err
	}

	// Populate the dense 2D array of its pixels.
	for px := range a.Iter() {
		var (
			point    = render.NewPoint(px.X, px.Y)
			relative = RelativeCoordinate(point, a.acc.coord, a.acc.size)
			ptr      = uint64(px.PaletteIndex)
		)
		grid[relative.Y][relative.X] = &ptr
	}

	// log.Error("2D GRID:\n%+v", grid)

	// Run-length encode the grid.
	var (
		compressed []byte
		firstColor = true
		lastColor  uint64
		runLength  uint64
	)
	for _, row := range grid {
		for _, color := range row {
			var index uint64
			if color == nil {
				index = 0xFF
			}

			if firstColor {
				lastColor = index
				runLength = 1
				firstColor = false
				continue
			}

			if index != lastColor {
				compressed = binary.AppendUvarint(compressed, index)
				compressed = binary.AppendUvarint(compressed, runLength)
				lastColor = index
				runLength = 1
				continue
			}

			runLength++
		}
	}

	log.Error("RLE compressed: %v", compressed)

	return compressed, nil
}

// UnmarshalBinary will decode a compressed RLEAccessor byte stream.
func (a *RLEAccessor) UnmarshalBinary(compressed []byte) error {
	a.acc.mu.Lock()
	defer a.acc.mu.Unlock()

	// New format: decompress the byte stream.
	log.Debug("RLEAccessor.Unmarshal: Reading %d bytes of compressed chunk data", len(compressed))

	// Prepare the 2D grid to decompress the RLE stream into.
	var (
		size         = int(a.acc.size)
		_, err       = Make2DChunkGrid(size)
		x, y, cursor int
	)
	if err != nil {
		return err
	}

	var reader = bytes.NewBuffer(compressed)

	for {
		var (
			paletteIndex, err1 = binary.ReadUvarint(reader)
			repeatCount, err2  = binary.ReadUvarint(reader)
		)

		if err1 != nil || err2 != nil {
			log.Error("reading Uvarints from compressed data: {%s, %s}", err1, err2)
			break
		}

		for i := uint64(0); i < repeatCount; i++ {
			cursor++
			if cursor%size == 0 {
				y++
				x = 0
			} else {
				x++
			}

			point := render.NewPoint(int(x), int(y))
			if paletteIndex != 0xFF {
				a.acc.grid[point] = NewSparseSwatch(int(paletteIndex))
			}
		}
	}

	return nil
}
