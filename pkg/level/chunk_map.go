package level

import (
	"encoding/json"
	"errors"
	"fmt"

	"git.kirsle.net/apps/doodle/lib/render"
	"github.com/vmihailenco/msgpack"
)

// MapAccessor implements a chunk accessor by using a map of points to their
// palette indexes. This is the simplest accessor and is best for sparse chunks.
type MapAccessor map[render.Point]*Swatch

// NewMapAccessor initializes a MapAccessor.
func NewMapAccessor() MapAccessor {
	return MapAccessor{}
}

// Inflate the sparse swatches from their palette indexes.
func (a MapAccessor) Inflate(pal *Palette) error {
	for point, swatch := range a {
		if swatch.IsSparse() {
			// Replace this with the correct swatch from the palette.
			if swatch.paletteIndex >= len(pal.Swatches) {
				return fmt.Errorf("MapAccessor.Inflate: swatch for point %s has paletteIndex %d but palette has only %d colors",
					point,
					swatch.paletteIndex,
					len(pal.Swatches),
				)
			}
			a[point] = pal.Swatches[swatch.paletteIndex]
		}
	}
	return nil
}

// Len returns the current size of the map, or number of pixels registered.
func (a MapAccessor) Len() int {
	return len(a)
}

// IterViewport returns a channel to loop over pixels in the viewport.
func (a MapAccessor) IterViewport(viewport render.Rect) <-chan Pixel {
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
func (a MapAccessor) Iter() <-chan Pixel {
	pipe := make(chan Pixel)
	go func() {
		for point, swatch := range a {
			pipe <- Pixel{
				X:      point.X,
				Y:      point.Y,
				Swatch: swatch,
			}
		}
		close(pipe)
	}()
	return pipe
}

// Get a pixel from the map.
func (a MapAccessor) Get(p render.Point) (*Swatch, error) {
	pixel, ok := a[p]
	if !ok {
		return nil, errors.New("no pixel")
	}
	return pixel, nil
}

// Set a pixel on the map.
func (a MapAccessor) Set(p render.Point, sw *Swatch) error {
	a[p] = sw
	return nil
}

// Delete a pixel from the map.
func (a MapAccessor) Delete(p render.Point) error {
	if _, ok := a[p]; ok {
		delete(a, p)
		return nil
	}
	return errors.New("pixel was not there")
}

// MarshalJSON to convert the chunk map to JSON.
//
// When serialized, the key is the "X,Y" coordinate and the value is the
// swatch index of the Palette, rather than redundantly serializing out the
// Swatch object for every pixel.
func (a MapAccessor) MarshalJSON() ([]byte, error) {
	dict := map[string]int{}
	for point, sw := range a {
		dict[point.String()] = sw.Index()
	}

	out, err := json.Marshal(dict)
	return out, err
}

// UnmarshalJSON to convert the chunk map back from JSON.
func (a MapAccessor) UnmarshalJSON(b []byte) error {
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
		a[point] = NewSparseSwatch(index)
	}

	return nil
}

// // MarshalMsgpack serializes for msgpack.
// func (a MapAccessor) MarshalMsgpack() ([]byte, error) {
// 	dict := map[string]int{}
// 	for point, sw := range a {
// 		dict[point.String()] = sw.Index()
// 	}
// 	return msgpack.Marshal(dict)
// }
//
// // Serialize converts the chunk accessor to a map for serialization.
// func (a MapAccessor) Serialize() interface{} {
// 	dict := map[string]int{}
// 	for point, sw := range a {
// 		dict[point.String()] = sw.Index()
// 	}
// 	return dict
// }
//
// // UnmarshalMsgpack decodes from msgpack format.
// func (a MapAccessor) UnmarshalMsgpack(b []byte) error {
// 	var dict map[string]int
// 	err := msgpack.Unmarshal(b, &dict)
// 	if err != nil {
// 		return err
// 	}
//
// 	for coord, index := range dict {
// 		point, err := render.ParsePoint(coord)
// 		if err != nil {
// 			return fmt.Errorf("MapAccessor.UnmarshalJSON: %s", err)
// 		}
// 		a[point] = NewSparseSwatch(index)
// 	}
//
// 	return nil
// }

func (a MapAccessor) EncodeMsgpack(enc *msgpack.Encoder) error {
	dict := map[string]int{}
	for point, sw := range a {
		dict[point.String()] = sw.Index()
	}
	return enc.Encode(dict)
}

func (a MapAccessor) DecodeMsgpack(dec *msgpack.Decoder) error {
	v, err := dec.DecodeMap()
	if err != nil {
		return fmt.Errorf("MapAccessor.DecodeMsgpack: %s", err)
	}
	dict := v.(map[string]int)

	for coord, index := range dict {
		point, err := render.ParsePoint(coord)
		if err != nil {
			return fmt.Errorf("MapAccessor.UnmarshalJSON: %s", err)
		}
		a[point] = NewSparseSwatch(index)
	}
	return nil
}
