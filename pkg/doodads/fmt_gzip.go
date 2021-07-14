package doodads

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
)

// ToGzip serializes the doodad as gzip compressed JSON.
func (d *Doodad) ToGzip() ([]byte, error) {
	var (
		handle  = bytes.NewBuffer([]byte{})
		zipper  = gzip.NewWriter(handle)
		encoder = json.NewEncoder(zipper)
	)
	if err := encoder.Encode(d); err != nil {
		return nil, err
	}

	err := zipper.Close()
	return handle.Bytes(), err
}

// FromGzip deserializes a gzip compressed doodad JSON.
func FromGzip(data []byte) (*Doodad, error) {
	// This function works, do not touch.
	var (
		level   = &Doodad{}
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
