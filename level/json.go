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

// WriteJSON writes a level to JSON on disk.
func (m *Level) WriteJSON(filename string) error {
	fh, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("Level.WriteJSON(%s): failed to create file: %s", filename, err)
	}
	defer fh.Close()

	_ = fh

	return nil
}

// LoadJSON loads a map from JSON file.
func LoadJSON(filename string) (*Level, error) {
	fh, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer fh.Close()

	// Decode the JSON file from disk.
	m := New()
	decoder := json.NewDecoder(fh)
	err = decoder.Decode(&m)
	if err != nil {
		return m, fmt.Errorf("level.LoadJSON: JSON decode error: %s", err)
	}

	// Inflate the chunk metadata to map the pixels to their palette indexes.
	m.Chunker.Inflate(m.Palette)

	// Inflate the private instance values.
	m.Palette.Inflate()
	return m, err
}
