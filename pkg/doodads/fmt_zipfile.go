package doodads

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"

	"git.kirsle.net/apps/doodle/pkg/balance"
	"git.kirsle.net/apps/doodle/pkg/log"
)

// ToZipfile serializes the doodad into zipfile format.
func (d *Doodad) ToZipfile() ([]byte, error) {
	fh := bytes.NewBuffer([]byte{})
	zipper := zip.NewWriter(fh)
	defer zipper.Close()

	// Migrate the Chunker caches into the zipfile.
	for _, layer := range d.Layers {
		if err := layer.Chunker.MigrateZipfile(zipper); err != nil {
			return nil, fmt.Errorf("MigrateZipfile: %s", err)
		}
	}

	// Write the header json.
	{
		header, err := d.AsJSON()
		if err != nil {
			return nil, err
		}

		writer, err := zipper.Create("doodad.json")
		if err != nil {
			return nil, err
		}

		if n, err := writer.Write(header); err != nil {
			return nil, err
		} else {
			log.Debug("Written doodad.json to zipfile: %d bytes", n)
		}
	}

	zipper.Close()

	// Refresh our Zipfile reader from the zipper we just wrote.
	bin := fh.Bytes()
	if err := d.ReloadZipfile(bin); err != nil {
		log.Error("ReloadZipfile: %s", err)
	}

	return fh.Bytes(), nil
}

// FromZipfile reads a doodad from zipfile format.
func FromZipfile(data []byte) (*Doodad, error) {
	var (
		doodad = New(balance.DoodadSize)
		err    = doodad.populateFromZipfile(data)
	)
	return doodad, err
}

// ReloadZipfile re-reads the level's zipfile after a write.
func (d *Doodad) ReloadZipfile(data []byte) error {
	return d.populateFromZipfile(data)
}

// Common function between FromZipfile and ReloadZipFile.
func (d *Doodad) populateFromZipfile(data []byte) error {
	var (
		buf     = bytes.NewReader(data)
		zf      *zip.Reader
		decoder *json.Decoder
	)

	zf, err := zip.NewReader(buf, buf.Size())
	if err != nil {
		return err
	}

	// Read the doodad.json.
	file, err := zf.Open("doodad.json")
	if err != nil {
		return err
	}

	decoder = json.NewDecoder(file)
	err = decoder.Decode(d)

	// Keep the zipfile reader handy.
	d.Zipfile = zf
	for i, layer := range d.Layers {
		layer.Chunker.Layer = i
		layer.Chunker.Zipfile = zf
	}

	return err
}

// Loop may be called each loop to allow the level to maintain its
// memory usage, e.g., for chunks not requested recently from a zipfile
// level to free those from RAM.
func (d *Doodad) Loop() error {
	for _, layer := range d.Layers {
		layer.Chunker.FreeCaches()
	}
	return nil
}
