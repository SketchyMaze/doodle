package level

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"git.kirsle.net/apps/doodle/pkg/balance"
)

// ToJSON serializes the level as JSON.
func (m *Level) ToJSON() ([]byte, error) {
	out := bytes.NewBuffer([]byte{})
	encoder := json.NewEncoder(out)
	if balance.JSONIndent {
		encoder.SetIndent("", "\t")
	}
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

	// Decode the JSON file from disk.
	m := New()
	decoder := json.NewDecoder(fh)
	err = decoder.Decode(&m)
	if err != nil {
		return m, fmt.Errorf("level.LoadJSON: JSON decode error: %s", err)
	}

	// Fill in defaults.
	if m.Wallpaper == "" {
		m.Wallpaper = DefaultWallpaper
	}

	// Inflate the chunk metadata to map the pixels to their palette indexes.
	m.Chunker.Inflate(m.Palette)
	m.Actors.Inflate()

	// Inflate the private instance values.
	m.Palette.Inflate()
	return m, err
}

// WriteJSON writes a level to JSON on disk.
func (m *Level) WriteJSON(filename string) error {
	json, err := m.ToJSON()
	if err != nil {
		return fmt.Errorf("Level.WriteJSON: JSON encode error: %s", err)
	}

	err = ioutil.WriteFile(filename, json, 0755)
	if err != nil {
		return fmt.Errorf("Level.WriteJSON: WriteFile error: %s", err)
	}

	return nil
}
