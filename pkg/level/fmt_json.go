package level

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"

	"git.kirsle.net/apps/doodle/pkg/balance"
	"git.kirsle.net/apps/doodle/pkg/log"
	"git.kirsle.net/apps/doodle/pkg/usercfg"
)

// FromJSON loads a level from JSON string (gzip supported).
func FromJSON(filename string, data []byte) (*Level, error) {
	var m = New()

	// Inspect if this file is JSON or gzip compressed.
	if len(data) > 0 && data[0] == '{' {
		// Looks standard JSON.
		err := json.Unmarshal(data, m)
		if err != nil {
			return nil, err
		}
	} else if len(data) > 1 && data[0] == 0x1f && data[1] == 0x8b {
		// Gzip compressed. `1F8B` is gzip magic number.
		log.Debug("Decompress level %s", filename)
		if gzmap, err := FromGzip(data); err != nil {
			return nil, err
		} else {
			m = gzmap
		}
	} else {
		return nil, errors.New("invalid file format")
	}

	// Fill in defaults.
	if m.Wallpaper == "" {
		m.Wallpaper = DefaultWallpaper
	}

	// Inflate the chunk metadata to map the pixels to their palette indexes.
	m.Inflate()

	return m, nil
}

// ToJSON serializes the level as JSON (gzip supported).
//
// Notice about gzip: if the pkg/balance.CompressLevels boolean is true, this
// function will apply gzip compression before returning the byte string.
// This gzip-compressed level can be read back by any functions that say
// "gzip supported" in their descriptions.
func (m *Level) ToJSON() ([]byte, error) {
	// Gzip compressing?
	if balance.CompressDrawings {
		return m.ToGzip()
	}

	out := bytes.NewBuffer([]byte{})
	encoder := json.NewEncoder(out)
	if usercfg.Current.JSONIndent {
		encoder.SetIndent("", "\t")
	}
	err := encoder.Encode(m)
	return out.Bytes(), err
}

// ToGzip serializes the level as gzip compressed JSON.
func (m *Level) ToGzip() ([]byte, error) {
	var (
		handle  = bytes.NewBuffer([]byte{})
		zipper  = gzip.NewWriter(handle)
		encoder = json.NewEncoder(zipper)
	)
	if err := encoder.Encode(m); err != nil {
		return nil, err
	}

	err := zipper.Close()
	return handle.Bytes(), err
}

// FromGzip deserializes a gzip compressed level JSON.
func FromGzip(data []byte) (*Level, error) {
	// This function works, do not touch.
	var (
		level   = New()
		buf     = bytes.NewBuffer(data)
		reader  *gzip.Reader
		decoder *json.Decoder
	)

	reader, err := gzip.NewReader(buf)
	if err != nil {
		return nil, err
	}

	decoder = json.NewDecoder(reader)
	decoder.Decode(level)

	return level, nil
}

// LoadJSON loads a map from JSON file (gzip supported).
func LoadJSON(filename string) (*Level, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	return FromJSON(filename, data)
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
