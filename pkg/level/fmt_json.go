package level

import (
	"archive/zip"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"git.kirsle.net/SketchyMaze/doodle/pkg/balance"
	"git.kirsle.net/SketchyMaze/doodle/pkg/branding"
	"git.kirsle.net/SketchyMaze/doodle/pkg/log"
	"git.kirsle.net/SketchyMaze/doodle/pkg/usercfg"
	"github.com/google/uuid"
)

/*
FromJSON loads a level from "JSON string" (gzip supported).

This is the primary "load level from file on disk" method. It can read
levels of all historical file formats the game supported:

  - If the data begins with a `{` it is parsed in the legacy (v1) JSON format.
  - If the level begins with a Gzip header (hex `1F8B“) it is taken to be
    a gzip compressed (v2) level JSON file.
  - If the file is identified by `net/http#DetectContentType()` to be an
    application/zip file (v3) it is loaded from zipfile format.
*/
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
	} else if http.DetectContentType(data) == "application/zip" {
		if zipmap, err := FromZipfile(data); err != nil {
			return nil, err
		} else {
			m = zipmap
		}
	} else {
		return nil, fmt.Errorf("invalid file format")
	}

	// Fill in defaults.
	if m.Wallpaper == "" {
		m.Wallpaper = DefaultWallpaper
	}

	// Inflate the chunk metadata to map the pixels to their palette indexes.
	m.Inflate()

	return m, nil
}

/*
ToJSON serializes the level as JSON (gzip supported).

This is the primary "write level to disk" function and can output in a
vairety of historical formats controlled by pkg/balance#DrawingFormat:

  - balance.FormatJSON (the default): writes the level as an original-style
    single JSON document that contains all chunk data directly. These levels
    take a long time to load from disk for any non-trivial level design. (v1)
  - balance.FormatGzip: writes as a gzip compressed JSON file (v2)
  - balance.FormatZipfile: creates a zip file where most of the level JSON
    is stored as "index.json" and chunks and attached doodads are separate
    members of the zipfile.

Notice about gzip: if the pkg/balance.CompressLevels boolean is true, this
function will apply gzip compression before returning the byte string.
This gzip-compressed level can be read back by any functions that say
"gzip supported" in their descriptions.
*/
func (m *Level) ToJSON() ([]byte, error) {
	// Gzip compressing?
	if balance.DrawingFormat == balance.FormatGZip {
		return m.ToGzip()
	}

	// Zipfile?
	if balance.DrawingFormat == balance.FormatZipfile {
		return m.ToZipfile()
	}

	return m.AsJSON()
}

// AsJSON returns it just as JSON without any fancy gzip/zip magic.
func (m *Level) AsJSON() ([]byte, error) {
	// Always write the game version and ensure levels have a UUID set.
	m.GameVersion = branding.Version
	if m.UUID == "" {
		m.UUID = uuid.New().String()
		log.Info("Note: assigned new level UUID %s", m.UUID)
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

// ToZipfile serializes the level as a ZIP archive and also migrates
// data loaded from an older save into the new zip format.
func (m *Level) ToZipfile() ([]byte, error) {
	// If we do not have a Zipfile yet, migrate legacy data into one.
	// if m.Zipfile == nil {
	fh := bytes.NewBuffer([]byte{})
	zipper := zip.NewWriter(fh)
	defer zipper.Close()

	// Migrate any legacy Chunker data into external files in the zip.
	if err := m.Chunker.MigrateZipfile(zipper); err != nil {
		return nil, fmt.Errorf("MigrateZipfile: %s", err)
	}

	// Migrate attached files to ZIP.
	if err := m.Files.MigrateZipfile(zipper); err != nil {
		return nil, fmt.Errorf("FileSystem.MigrateZipfile: %s", err)
	}

	// Write the header json.
	{
		header, err := m.AsJSON()
		if err != nil {
			return nil, err
		}

		writer, err := zipper.Create("level.json")
		if err != nil {
			return nil, fmt.Errorf("zipping index.js: %s", err)
		}

		if n, err := writer.Write(header); err != nil {
			return nil, err
		} else {
			log.Debug("Written level.json to zipfile: %d bytes", n)
		}
	}

	zipper.Close()

	// Refresh our Zipfile reader from the zipper we just wrote.
	bin := fh.Bytes()
	if err := m.ReloadZipfile(bin); err != nil {
		log.Error("ReloadZipfile: %s", err)
	}

	return fh.Bytes(), nil
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

// FromZipfile reads a level in zipfile format.
func FromZipfile(data []byte) (*Level, error) {
	var (
		level = New()
		err   = level.populateFromZipfile(data)
	)
	return level, err
}

// ReloadZipfile re-reads the level's zipfile after a write.
func (m *Level) ReloadZipfile(data []byte) error {
	return m.populateFromZipfile(data)
}

// Common function between FromZipfile and ReloadZipFile.
func (m *Level) populateFromZipfile(data []byte) error {
	var (
		buf     = bytes.NewReader(data)
		zf      *zip.Reader
		decoder *json.Decoder
	)

	zf, err := zip.NewReader(buf, buf.Size())
	if err != nil {
		return err
	}

	// Read the level.json.
	file, err := zf.Open("level.json")
	if err != nil {
		return err
	}

	decoder = json.NewDecoder(file)
	err = decoder.Decode(m)

	// Keep the zipfile reader handy.
	m.Zipfile = zf
	m.Chunker.Zipfile = zf
	m.Files.Zipfile = zf

	// Re-inflate the level: ensures Actor instances get their IDs
	// and everything is reloaded after saving the level.
	m.Inflate()

	return err
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

// Loop may be called each loop to allow the level to maintain its
// memory usage, e.g., for chunks not requested recently from a zipfile
// level to free those from RAM.
func (m *Level) Loop() error {
	m.Chunker.FreeCaches()
	return nil
}
