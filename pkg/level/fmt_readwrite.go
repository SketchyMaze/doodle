package level

import (
	"fmt"
	"io/ioutil"
	"runtime"
	"strings"

	"git.kirsle.net/SketchyMaze/doodle/assets"
	"git.kirsle.net/SketchyMaze/doodle/pkg/branding"
	"git.kirsle.net/SketchyMaze/doodle/pkg/enum"
	"git.kirsle.net/SketchyMaze/doodle/pkg/filesystem"
	"git.kirsle.net/SketchyMaze/doodle/pkg/log"
	"git.kirsle.net/SketchyMaze/doodle/pkg/userdir"
	"git.kirsle.net/SketchyMaze/doodle/pkg/wasm"
)

// ListSystemLevels returns a list of built-in levels.
func ListSystemLevels() ([]string, error) {
	var names = []string{}

	// Add the levels embedded inside the binary.
	if levels, err := assets.AssetDir("assets/levels"); err == nil {
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
	if jsonData, err := assets.Asset(filename); err == nil {
		log.Debug("Level %s: loaded from embedded bindata", filename)
		return FromJSON(filename, jsonData)
	}

	// WASM: try the file from localStorage or HTTP ajax request.
	if runtime.GOOS == "js" {
		if result, ok := wasm.GetSession(filename); ok {
			log.Info("recall level data from localStorage")
			return FromJSON(filename, []byte(result))
		}

		// Ajax request.
		jsonData, err := wasm.HTTPGet(filename)
		if err != nil {
			return nil, err
		}

		return FromJSON(filename, jsonData)
	}

	// Load as JSON.
	if level, err := LoadJSON(filename); err == nil {
		return level, nil
	} else {
		log.Warn(err.Error())
		return nil, err
	}
}

// WriteFile saves a level to disk in the user's config directory.
func (m *Level) WriteFile(filename string) error {
	if !strings.HasSuffix(filename, enum.LevelExt) {
		filename += enum.LevelExt
	}

	// Set the version information.
	m.Version = 1
	m.GameVersion = branding.Version

	// Maintenance functions, clean up cruft before save.
	m.PruneLinks()

	bin, err := m.ToJSON()
	if err != nil {
		return err
	}

	// Save it to their profile directory.
	filename = userdir.LevelPath(filename)
	log.Info("Write Level: %s", filename)

	// WASM: place in localStorage.
	if runtime.GOOS == "js" {
		log.Info("wasm: write %s to localStorage", filename)
		wasm.SetSession(filename, string(bin))
		return nil
	}

	// Desktop: write to disk.
	err = ioutil.WriteFile(filename, bin, 0644)
	if err != nil {
		return fmt.Errorf("level.WriteFile: %s", err)
	}

	return nil
}
