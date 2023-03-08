package doodads

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"

	"git.kirsle.net/SketchyMaze/doodle/pkg/balance"
	"git.kirsle.net/SketchyMaze/doodle/pkg/branding"
	"git.kirsle.net/SketchyMaze/doodle/pkg/usercfg"
)

// ToJSON serializes the doodad as JSON (gzip supported).
//
// If balance.CompressLevels=true the doodad will be gzip compressed
// and the return value is gz bytes and not the raw JSON.
func (d *Doodad) ToJSON() ([]byte, error) {
	// Gzip compressing?
	if balance.DrawingFormat == balance.FormatGZip {
		return d.ToGzip()
	}

	// Zipfile?
	if balance.DrawingFormat == balance.FormatZipfile {
		return d.ToZipfile()
	}

	return d.AsJSON()
}

// AsJSON returns it just as JSON without any fancy gzip/zip magic.
func (d *Doodad) AsJSON() ([]byte, error) {
	// Always write the game version that last saved this doodad.
	d.GameVersion = branding.Version

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
		if gzd, err := FromGzip(data); err != nil {
			return nil, err
		} else {
			doodad = gzd
		}
	} else if http.DetectContentType(data) == "application/zip" {
		if zipdoodad, err := FromZipfile(data); err != nil {
			return nil, err
		} else {
			doodad = zipdoodad
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
