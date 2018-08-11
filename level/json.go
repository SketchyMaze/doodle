package level

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
)

// ToJSON serializes the level as JSON.
func (m *Level) ToJSON() ([]byte, error) {
	out := bytes.NewBuffer([]byte{})
	encoder := json.NewEncoder(out)
	encoder.SetIndent("", "\t")
	err := encoder.Encode(m)
	return out.Bytes(), err
}

// LoadJSON loads a map from JSON file.
func LoadJSON(filename string) (*Level, error) {
	fh, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer fh.Close()

	m := New()
	decoder := json.NewDecoder(fh)
	err = decoder.Decode(&m)
	if err != nil {
		return m, err
	}

	// Inflate the private instance values.
	for _, px := range m.Pixels {
		if int(px.PaletteIndex) > len(m.Palette.Swatches) {
			return nil, fmt.Errorf(
				"pixel %s references palette index %d but there are only %d swatches in the palette",
				px, px.PaletteIndex, len(m.Palette.Swatches),
			)
		}
		px.Palette = m.Palette
		px.Swatch = m.Palette.Swatches[px.PaletteIndex]
	}
	return m, err
}
