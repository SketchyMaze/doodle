package doodads

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"

	"git.kirsle.net/apps/doodle/pkg/balance"
	"git.kirsle.net/apps/doodle/pkg/log"
	"git.kirsle.net/apps/doodle/pkg/usercfg"
)

// ToJSON serializes the doodad as JSON (gzip supported).
//
// If balance.CompressLevels=true the doodad will be gzip compressed
// and the return value is gz bytes and not the raw JSON.
func (d *Doodad) ToJSON() ([]byte, error) {
	// Gzip compressing?
	if balance.CompressDrawings {
		return d.ToGzip()
	}

	out := bytes.NewBuffer([]byte{})
	encoder := json.NewEncoder(out)
	if usercfg.Current.JSONIndent {
		encoder.SetIndent("", "\t")
	}
	err := encoder.Encode(d)
	return out.Bytes(), err
}

// FromJSON loads a doodad from JSON string (gzip supported).
func FromJSON(filename string, data []byte) (*Doodad, error) {
	var doodad = &Doodad{}

	// Inspect the headers of the file to see how it was encoded.
	if len(data) > 0 && data[0] == '{' {
		// Looks standard JSON.
		err := json.Unmarshal(data, doodad)
		if err != nil {
			return nil, err
		}
	} else if len(data) > 1 && data[0] == 0x1f && data[1] == 0x8b {
		// Gzip compressed. `1F8B` is gzip magic number.
		log.Debug("Decompress doodad %s", filename)
		if gzd, err := FromGzip(data); err != nil {
			return nil, err
		} else {
			doodad = gzd
		}
	}

	// Inflate the chunk metadata to map the pixels to their palette indexes.
	doodad.Filename = filepath.Base(filename)
	doodad.Inflate()

	return doodad, nil
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
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	return FromJSON(filename, data)
}
