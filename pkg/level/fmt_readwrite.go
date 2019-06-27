package level

import (
	"errors"
	"fmt"
	"io/ioutil"
	"runtime"
	"strings"

	"git.kirsle.net/apps/doodle/pkg/bindata"
	"git.kirsle.net/apps/doodle/pkg/branding"
	"git.kirsle.net/apps/doodle/pkg/enum"
	"git.kirsle.net/apps/doodle/pkg/filesystem"
	"git.kirsle.net/apps/doodle/pkg/log"
	"git.kirsle.net/apps/doodle/pkg/userdir"
	"git.kirsle.net/apps/doodle/pkg/wasm"
)

// ListSystemLevels returns a list of built-in levels.
func ListSystemLevels() ([]string, error) {
	var names = []string{}

	// Add the levels embedded inside the binary.
	if levels, err := bindata.AssetDir("assets/levels"); err == nil {
		names = append(names, levels...)
	}

	// WASM
	if runtime.GOOS == "js" {
		// Return just the embedded ones, no filesystem access.
		return names, nil
	}

	// Read filesystem for system levels.
	files, err := ioutil.ReadDir(filesystem.SystemLevelsPath)
	for _, file := range files {
		name := file.Name()
		if strings.HasSuffix(strings.ToLower(name), enum.DoodadExt) {
			names = append(names, name)
		}
	}

	return names, err
}

// LoadFile reads a level file from disk, checking a few locations.
func LoadFile(filename string) (*Level, error) {
	if !strings.HasSuffix(filename, enum.LevelExt) {
		filename += enum.LevelExt
	}

	// Search the system and user paths for this level.
	filename, err := filesystem.FindFile(filename)
	if err != nil {
		return nil, err
	}

	// Do we have the file in bindata?
	if jsonData, err := bindata.Asset(filename); err == nil {
		log.Info("loaded from embedded bindata")
		return FromJSON(filename, jsonData)
	}

	// WASM: try the file over HTTP ajax request.
	if runtime.GOOS == "js" {
		jsonData, err := wasm.HTTPGet(filename)
		if err != nil {
			return nil, err
		}

		return FromJSON(filename, jsonData)
	}

	// Try the binary format.
	if level, err := LoadBinary(filename); err == nil {
		return level, nil
	} else {
		log.Warn(err.Error())
	}

	// Then the JSON format.
	if level, err := LoadJSON(filename); err == nil {
		return level, nil
	} else {
		log.Warn(err.Error())
	}

	return nil, errors.New("invalid file type")
}

// WriteFile saves a level to disk in the user's config directory.
func (m *Level) WriteFile(filename string) error {
	if !strings.HasSuffix(filename, enum.LevelExt) {
		filename += enum.LevelExt
	}

	// Set the version information.
	m.Version = 1
	m.GameVersion = branding.Version

	// bin, err := m.ToBinary()
	bin, err := m.ToJSON()
	if err != nil {
		return err
	}

	// Save it to their profile directory.
	filename = userdir.LevelPath(filename)
	log.Info("Write Level: %s", filename)
	err = ioutil.WriteFile(filename, bin, 0644)
	if err != nil {
		return fmt.Errorf("level.WriteFile: %s", err)
	}

	return nil
}
