package level

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"

	"git.kirsle.net/go/render"
)

// MapAccessor implements a chunk accessor by using a map of points to their
// palette indexes. This is the simplest accessor and is best for sparse chunks.
type MapAccessor struct {
	grid map[render.Point]*Swatch
	mu   sync.RWMutex
}

// NewMapAccessor initializes a MapAccessor.
func NewMapAccessor() *MapAccessor {
	return &MapAccessor{
		grid: map[render.Point]*Swatch{},
	}
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
func (a *MapAccessor) MarshalJSON() ([]byte, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	dict := map[string]int{}
	for point, sw := range a.grid {
		dict[point.String()] = sw.Index()
	}

	out, err := json.Marshal(dict)
	return out, err
}

// UnmarshalJSON to convert the chunk map back from JSON.
func (a *MapAccessor) UnmarshalJSON(b []byte) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	var dict map[string]int
	err := json.Unmarshal(b, &dict)
	if err != nil {
		return err
	}

	for coord, index := range dict {
		point, err := render.ParsePoint(coord)
		if err != nil {
			return fmt.Errorf("MapAccessor.UnmarshalJSON: %s", err)
		}
		a.grid[point] = NewSparseSwatch(index)
	}

	return nil
}
