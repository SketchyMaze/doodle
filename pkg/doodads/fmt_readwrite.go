package doodads

import (
	"errors"
	"fmt"
	"io/ioutil"
	"strings"

	"git.kirsle.net/apps/doodle/pkg/enum"
	"git.kirsle.net/apps/doodle/pkg/filesystem"
	"git.kirsle.net/apps/doodle/pkg/log"
	"git.kirsle.net/apps/doodle/pkg/userdir"
)

// ListDoodads returns a listing of all available doodads between all locations,
// including user doodads.
func ListDoodads() ([]string, error) {
	var names []string

	// Read system-level doodads first. Ignore errors, if the system path is
	// empty we still go on to read the user directory.
	files, _ := ioutil.ReadDir(filesystem.SystemDoodadsPath)

	for _, file := range files {
		name := file.Name()
		if strings.HasSuffix(strings.ToLower(name), enum.DoodadExt) {
			names = append(names, name)
		}
	}

	// Append user doodads.
	userFiles, err := userdir.ListDoodads()
	names = append(names, userFiles...)
	return names, err
}

// LoadFile reads a doodad file from disk, checking a few locations.
func LoadFile(filename string) (*Doodad, error) {
	if !strings.HasSuffix(filename, enum.DoodadExt) {
		filename += enum.DoodadExt
	}

	// Search the system and user paths for this level.
	filename, err := filesystem.FindFile(filename)
	if err != nil {
		return nil, fmt.Errorf("doodads.LoadFile(%s): %s", filename, err)
	}

	// Load the JSON format.
	if doodad, err := LoadJSON(filename); err == nil {
		return doodad, nil
	} else {
		log.Warn(err.Error())
	}

	return nil, errors.New("invalid file type")
}

// WriteFile saves a doodad to disk in the user's config directory.
func (d *Doodad) WriteFile(filename string) error {
	if !strings.HasSuffix(filename, enum.DoodadExt) {
		filename += enum.DoodadExt
	}

	// bin, err := m.ToBinary()
	bin, err := d.ToJSON()
	if err != nil {
		return err
	}

	// Save it to their profile directory.
	filename = userdir.DoodadPath(filename)
	log.Info("Write Doodad: %s", filename)
	err = ioutil.WriteFile(filename, bin, 0644)
	if err != nil {
		return fmt.Errorf("doodads.WriteFile: %s", err)
	}

	return nil
}
