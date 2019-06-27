package doodads

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

// ToJSON serializes the doodad as JSON.
func (d *Doodad) ToJSON() ([]byte, error) {
	out := bytes.NewBuffer([]byte{})
	encoder := json.NewEncoder(out)
	encoder.SetIndent("", "\t")
	err := encoder.Encode(d)
	return out.Bytes(), err
}

// FromJSON loads a doodad from JSON string.
func FromJSON(filename string, data []byte) (*Doodad, error) {
	var doodad = &Doodad{}
	err := json.Unmarshal(data, doodad)

	// Inflate the chunk metadata to map the pixels to their palette indexes.
	doodad.Filename = filepath.Base(filename)
	doodad.Inflate()

	return doodad, err
}

// WriteJSON writes a Doodad to JSON on disk.
func (d *Doodad) WriteJSON(filename string) error {
	json, err := d.ToJSON()
	if err != nil {
		return fmt.Errorf("Doodad.WriteJSON: JSON encode error: %s", err)
	}

	err = ioutil.WriteFile(filename, json, 0755)
	if err != nil {
		return fmt.Errorf("Doodad.WriteJSON: WriteFile error: %s", err)
	}

	d.Filename = filepath.Base(filename)
	return nil
}

// LoadJSON loads a map from JSON file.
func LoadJSON(filename string) (*Doodad, error) {
	fh, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer fh.Close()

	// Decode the JSON file from disk.
	d := New(0)
	decoder := json.NewDecoder(fh)
	err = decoder.Decode(&d)
	if err != nil {
		return d, fmt.Errorf("doodad.LoadJSON: JSON decode error: %s", err)
	}

	// Inflate the chunk metadata to map the pixels to their palette indexes.
	d.Filename = filepath.Base(filename)
	d.Inflate()
	return d, err
}
