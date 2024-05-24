package level

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"sync"

	"git.kirsle.net/SketchyMaze/doodle/pkg/balance"
	"git.kirsle.net/SketchyMaze/doodle/pkg/log"
	"git.kirsle.net/go/render"
)

// MapAccessor implements a chunk accessor by using a map of points to their
// palette indexes. This is the simplest accessor and is best for sparse chunks.
type MapAccessor struct {
	chunk *Chunk // Pointer to parent struct, for its Size and Point
	grid  map[render.Point]*Swatch
	mu    sync.RWMutex
}

// NewMapAccessor initializes a MapAccessor.
func NewMapAccessor(chunk *Chunk) *MapAccessor {
	return &MapAccessor{
		chunk: chunk,
		grid:  map[render.Point]*Swatch{},
	}
}

// Reset the MapAccessor.
func (a *MapAccessor) Reset() {
	a.grid = map[render.Point]*Swatch{}
}

// Inflate the sparse swatches from their palette indexes.
func (a *MapAccessor) Inflate(pal *Palette) error {
	for point, swatch := range a.grid {
		if swatch.IsSparse() {
			// Replace this with the correct swatch from the palette.
			if swatch.paletteIndex >= len(pal.Swatches) {
				return fmt.Errorf("MapAccessor.Inflate: swatch for point %s has paletteIndex %d but palette has only %d colors",
					point,
					swatch.paletteIndex,
					len(pal.Swatches),
				)
			}

			a.mu.Lock()
			a.grid[point] = pal.Swatches[swatch.paletteIndex] // <- concurrent write
			a.mu.Unlock()
		}
	}
	return nil
}

// Len returns the current size of the map, or number of pixels registered.
func (a *MapAccessor) Len() int {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return len(a.grid)
}

// IterViewport returns a channel to loop over pixels in the viewport.
func (a *MapAccessor) IterViewport(viewport render.Rect) <-chan Pixel {
	pipe := make(chan Pixel)
	go func() {
		for px := range a.Iter() {
			if px.Point().Inside(viewport) {
				pipe <- px
			}
		}
		close(pipe)
	}()
	return pipe
}

// Iter returns a channel to loop over all points in this chunk.
func (a *MapAccessor) Iter() <-chan Pixel {
	pipe := make(chan Pixel)
	go func() {
		a.mu.Lock()
		for point, swatch := range a.grid {
			pipe <- Pixel{
				X:      point.X,
				Y:      point.Y,
				Swatch: swatch,
			}
		}
		a.mu.Unlock()
		close(pipe)
	}()
	return pipe
}

// Get a pixel from the map.
func (a *MapAccessor) Get(p render.Point) (*Swatch, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	pixel, ok := a.grid[p] // <- concurrent read and write
	if !ok {
		return nil, errors.New("no pixel")
	}
	return pixel, nil
}

// Set a pixel on the map.
func (a *MapAccessor) Set(p render.Point, sw *Swatch) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.grid[p] = sw
	return nil
}

// Delete a pixel from the map.
func (a *MapAccessor) Delete(p render.Point) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if _, ok := a.grid[p]; ok {
		delete(a.grid, p)
		return nil
	}

	return errors.New("pixel was not there")
}

// MarshalJSON to convert the chunk map to JSON.
//
// When serialized, the key is the "X,Y" coordinate and the value is the
// swatch index of the Palette, rather than redundantly serializing out the
// Swatch object for every pixel.
//
// DEPRECATED: in the Zipfile format chunks will be saved as binary files
// instead of with their JSON wrappers, so MarshalJSON will be phased out.
func (a *MapAccessor) MarshalJSON() ([]byte, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	// Write in the new compressed format.
	if balance.CompressMapAccessor {
		var compressed []byte
		for point, sw := range a.grid {
			var (
				x     = int64(point.X)
				y     = int64(point.Y)
				sw    = uint64(sw.index)
				entry = []byte{}
			)

			entry = binary.AppendVarint(entry, x)
			entry = binary.AppendVarint(entry, y)
			entry = binary.AppendUvarint(entry, sw)

			compressed = append(compressed, entry...)
		}

		out, err := json.Marshal(compressed)
		return out, err
	}

	dict := map[string]int{}
	for point, sw := range a.grid {
		dict[point.String()] = sw.Index()
	}

	out, err := json.Marshal(dict)
	return out, err
}

// UnmarshalJSON to convert the chunk map back from JSON.
//
// DEPRECATED: in the Zipfile format chunks will be saved as binary files
// instead of with their JSON wrappers, so MarshalJSON will be phased out.
func (a *MapAccessor) UnmarshalJSON(b []byte) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	// Transparently upgrade the compression algorithm for this level.
	// - Old style was a map[string]int like {"123,456": 4} mapping
	//   a coordinate to a palette index.
	// - Now, coords and palettes are uint8 constrained so we can
	//   really tighten this up.
	// For transparent upgrade, try and parse it the old way first.
	var (
		dict       map[string]int // old-style
		compressed []byte         // new-style
	)
	err := json.Unmarshal(b, &dict)
	if err != nil {
		// Now try the new way.
		err = json.Unmarshal(b, &compressed)
		if err != nil {
			return err
		}
	}

	// New format: decompress the byte stream.
	if compressed != nil {
		// log.Debug("MapAccessor.Unmarshal: Reading %d bytes of compressed chunk data", len(compressed))

		var (
			reader = bytes.NewBuffer(compressed)
		)

		for {
			var (
				x, err1  = binary.ReadVarint(reader)
				y, err2  = binary.ReadVarint(reader)
				sw, err3 = binary.ReadUvarint(reader)
			)

			point := render.NewPoint(int(x), int(y))
			a.grid[point] = NewSparseSwatch(int(sw))

			if err1 != nil || err2 != nil || err3 != nil {
				// log.Error("Break read loop: %s; %s; %s", err1, err2, err3)
				break
			}
		}
		return nil
	}

	// Old format: read the dict in.
	for coord, index := range dict {
		point, err := render.ParsePoint(coord)
		if err != nil {
			return fmt.Errorf("MapAccessor.UnmarshalJSON: %s", err)
		}
		a.grid[point] = NewSparseSwatch(index)
	}

	return nil
}

/*
MarshalBinary converts the chunk data to a binary representation, for
better compression compared to JSON.

In the binary format each chunk begins with one Varint (the chunk Type)
followed by whatever wire format the chunk needs given its type.

This function is related to the CompressMapAccessor config constant:
the MapAccessor compression boils down each point to a series if packed
varints: the X, Y coord (varint) followed by palette index (Uvarint).

The output of this function is just the compressed MapAccessor stream.
*/
func (a *MapAccessor) MarshalBinary() ([]byte, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	// Write in the new compressed format.
	var compressed []byte
	for point, sw := range a.grid {
		var (
			x     = int64(point.X)
			y     = int64(point.Y)
			sw    = uint64(sw.index)
			entry = []byte{}
		)

		entry = binary.AppendVarint(entry, x)
		entry = binary.AppendVarint(entry, y)
		entry = binary.AppendUvarint(entry, sw)

		compressed = append(compressed, entry...)
	}

	return compressed, nil
}

// UnmarshalBinary will decode a compressed MapAccessor byte stream.
func (a *MapAccessor) UnmarshalBinary(compressed []byte) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	// New format: decompress the byte stream.
	log.Debug("MapAccessor.Unmarshal: Reading %d bytes of compressed chunk data", len(compressed))

	var reader = bytes.NewBuffer(compressed)

	for {
		var (
			x, err1  = binary.ReadVarint(reader)
			y, err2  = binary.ReadVarint(reader)
			sw, err3 = binary.ReadUvarint(reader)
		)

		// We expect all 3 errors to be EOF together if the binary is formed correctly.
		if err1 != nil || err2 != nil || err3 != nil {
			if err1 == nil || err2 == nil || err3 == nil {
				log.Error("MapAccessor.UnmarshalBinary: found odd number of varints!")
			}
			break
		}

		point := render.NewPoint(int(x), int(y))
		a.grid[point] = NewSparseSwatch(int(sw))
	}

	return nil
}
