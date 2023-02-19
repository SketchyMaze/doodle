// Package levelpack handles ZIP archives for level packs.
package levelpack

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"runtime"
	"sort"
	"strings"
	"time"

	"git.kirsle.net/SketchyMaze/doodle/assets"
	"git.kirsle.net/SketchyMaze/doodle/pkg/balance"
	"git.kirsle.net/SketchyMaze/doodle/pkg/enum"
	"git.kirsle.net/SketchyMaze/doodle/pkg/filesystem"
	"git.kirsle.net/SketchyMaze/doodle/pkg/log"
	"git.kirsle.net/SketchyMaze/doodle/pkg/userdir"
)

// LevelPack describes the contents of a levelpack file.
type LevelPack struct {
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Author      string    `json:"author"`
	Created     time.Time `json:"created"`

	// Cached metadata about the (displayed) levels.
	Levels []Level `json:"levels"`

	// Number of levels unlocked by default.
	// 0 = all levels unlocked
	FreeLevels int `json:"freeLevels"`

	// The loaded zip file for reading an existing levelpack.
	Zipfile *zip.Reader `json:"-"`

	// A reference to the original filename, not stored in json.
	Filename string `json:"-"`

	// Signature to allow free versions of the game to load embedded
	// custom doodads inside this levelpack for its levels.
	Signature []byte `json:"signature,omitempty"`
}

// Level holds metadata about the levels in the levelpack.
type Level struct {
	Title    string `json:"title"`
	Author   string `json:"author"`
	Filename string `json:"filename"`
}

// LoadFile reads a .levelpack zip file.
func LoadFile(filename string) (*LevelPack, error) {
	var (
		fh       io.ReaderAt
		filesize int64
	)

	// Look in embedded bindata.
	if data, err := assets.Asset(filename); err == nil {
		filesize = int64(len(data))
		fh = bytes.NewReader(data)
	}

	// Try the filesystem.
	if fh == nil {
		stat, err := os.Stat(filename)
		if err != nil {
			return nil, err
		}
		filesize = stat.Size()

		fh, err = os.Open(filename)
		if err != nil {
			return nil, err
		}
	}

	// No luck?
	if fh == nil {
		return nil, errors.New("no file found")
	}

	reader, err := zip.NewReader(fh, filesize)
	if err != nil {
		return nil, err
	}

	lp := &LevelPack{
		Filename: filename,
		Zipfile:  reader,
	}

	// Read the index.json.
	lp.GetJSON(lp, "index.json")

	return lp, nil
}

// LoadAllAvailable loads every levelpack visible to the game. Returns
// the sorted list of filenames as from ListFiles, plus a deeply loaded
// hash map associating the filenames with their data.
func LoadAllAvailable() ([]string, map[string]*LevelPack, error) {
	filenames, err := ListFiles()
	if err != nil {
		return filenames, nil, err
	}

	var dictionary = map[string]*LevelPack{}
	for _, filename := range filenames {
		// Resolve the filename to a definite path on disk.
		path, err := filesystem.FindFile(filename)
		if err != nil {
			log.Error("LoadAllAvailable: FindFile(%s): %s", path, err)
			return filenames, nil, err
		}

		lp, err := LoadFile(path)
		if err != nil {
			return filenames, nil, fmt.Errorf("LoadAllAvailable: LoadFile(%s): %s", path, err)
		}

		dictionary[filename] = lp
	}

	return filenames, dictionary, nil
}

// ListFiles lists all the discoverable levelpack files, starting from
// the game's built-ins all the way to user levelpacks.
func ListFiles() ([]string, error) {
	var names []string

	// List levelpacks embedded into the binary.
	if files, err := assets.AssetDir("assets/levelpacks"); err == nil {
		names = append(names, files...)
	}

	// WASM stops here, no filesystem access.
	if runtime.GOOS == "js" {
		return names, nil
	}

	// Read system-level levelpacks.
	files, _ := ioutil.ReadDir(filesystem.SystemLevelPacksPath)
	for _, file := range files {
		name := file.Name()
		if strings.HasSuffix(name, enum.LevelPackExt) {
			names = append(names, name)
		}
	}

	// Append user levelpacks.
	files, _ = ioutil.ReadDir(userdir.LevelPackDirectory)
	for _, file := range files {
		name := file.Name()
		if strings.HasSuffix(name, enum.LevelPackExt) {
			names = append(names, name)
		}
	}

	// Deduplicate strings. Can happen e.g. because assets/ is baked
	// in to bindata but files also exist there locally.
	var (
		dedupe []string
		seen   = map[string]interface{}{}
	)
	for _, value := range names {
		if _, ok := seen[value]; !ok {
			seen[value] = nil
			dedupe = append(dedupe, value)
		}
	}

	sort.Strings(dedupe)
	return dedupe, nil
}

// WriteFile saves the metadata to a .json file on disk.
func (l LevelPack) WriteFile(filename string) error {
	out, err := json.Marshal(l)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(filename, out, 0655)
}

// WriteZipfile saves a levelpack back into a zip file.
func (l LevelPack) WriteZipfile(filename string) error {
	fh, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer fh.Close()

	// Copy all of the levels and other files from the old zip to new zip.
	zf := zip.NewWriter(fh)
	defer zf.Close()

	// Copy attached doodads and levels.
	for _, file := range l.Zipfile.File {
		if !strings.HasPrefix(file.Name, "doodads/") &&
			!strings.HasPrefix(file.Name, "levels/") {
			continue
		}

		if err := zf.Copy(file); err != nil {
			return err
		}
	}

	// Write the index.json metadata.
	meta, err := json.Marshal(l)
	if err != nil {
		return err
	}

	writer, err := zf.Create("index.json")
	if err != nil {
		return err
	}
	_, err = writer.Write(meta)
	return err
}

// GetFile returns file data from inside the loaded zipfile of a levelpack.
//
// This also implements the Embeddable interface.
func (l LevelPack) GetFile(filename string) ([]byte, error) {
	if l.Zipfile == nil {
		return []byte{}, errors.New("zipfile not loaded")
	}

	// NOTE: levelpacks don't have an "assets/" prefix but the game
	// might come looking for "assets/doodads"
	if strings.HasPrefix(filename, balance.EmbeddedDoodadsBasePath) {
		filename = strings.Replace(filename, balance.EmbeddedDoodadsBasePath, "doodads/", 1)
	}

	file, err := l.Zipfile.Open(filename)
	if err != nil {
		return []byte{}, err
	}

	return ioutil.ReadAll(file)
}

// GetJSON loads a JSON file from the zipfile and marshals it into your struct.
func (l LevelPack) GetJSON(v interface{}, filename string) error {
	data, err := l.GetFile(filename)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, v)
}

// ListFiles returns the files in the zipfile that match the prefix given.
func (l LevelPack) ListFiles(prefix string) []string {
	var result []string

	if l.Zipfile != nil {
		for _, file := range l.Zipfile.File {
			if strings.HasPrefix(file.Name, prefix) {
				result = append(result, file.Name)
			}
		}
	}

	return result
}
