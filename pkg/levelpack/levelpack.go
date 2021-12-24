// Package levelpack handles ZIP archives for level packs.
package levelpack

import (
	"archive/zip"
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"runtime"
	"strings"
	"time"

	"git.kirsle.net/apps/doodle/assets"
	"git.kirsle.net/apps/doodle/pkg/enum"
	"git.kirsle.net/apps/doodle/pkg/filesystem"
	"git.kirsle.net/apps/doodle/pkg/userdir"
)

// LevelPack describes the contents of a levelpack file.
type LevelPack struct {
	Title       string    `json:"title`
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
}

// Level holds metadata about the levels in the levelpack.
type Level struct {
	Title    string `json:"title"`
	Author   string `json:"author"`
	Filename string `json:"filename"`
}

// LoadFile reads a .levelpack zip file.
func LoadFile(filename string) (LevelPack, error) {
	stat, err := os.Stat(filename)
	if err != nil {
		return LevelPack{}, err
	}

	fh, err := os.Open(filename)
	if err != nil {
		return LevelPack{}, err
	}

	reader, err := zip.NewReader(fh, stat.Size())
	if err != nil {
		return LevelPack{}, err
	}

	lp := LevelPack{
		Zipfile: reader,
	}

	// Read the index.json.
	lp.GetJSON(&lp, "index.json")

	return lp, nil
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

	return names, nil
}

// WriteFile saves the metadata to a .json file on disk.
func (l LevelPack) WriteFile(filename string) error {
	out, err := json.Marshal(l)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(filename, out, 0655)
}

// GetData returns file data from inside the loaded zipfile of a levelpack.
func (l LevelPack) GetData(filename string) ([]byte, error) {
	if l.Zipfile == nil {
		return []byte{}, errors.New("zipfile not loaded")
	}

	file, err := l.Zipfile.Open(filename)
	if err != nil {
		return []byte{}, err
	}

	return ioutil.ReadAll(file)
}

// GetJSON loads a JSON file from the zipfile and marshals it into your struct.
func (l LevelPack) GetJSON(v interface{}, filename string) error {
	data, err := l.GetData(filename)
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
