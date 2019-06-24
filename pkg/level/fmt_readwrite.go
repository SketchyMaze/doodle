package level

import (
	"errors"
	"fmt"
	"io/ioutil"
	"strings"

	"git.kirsle.net/apps/doodle/pkg/branding"
	"git.kirsle.net/apps/doodle/pkg/enum"
	"git.kirsle.net/apps/doodle/pkg/filesystem"
	"git.kirsle.net/apps/doodle/pkg/log"
	"git.kirsle.net/apps/doodle/pkg/userdir"
)

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
