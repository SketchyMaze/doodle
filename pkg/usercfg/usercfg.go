/*
Package usercfg has functions around the user's Game Settings.

Other places in the codebase to look for its related functionality:

- pkg/windows/settings.go: the Settings Window is the UI owner of
  this feature, it adjusts the usercfg.Current struct and Saves the
  changes to disk.
*/
package usercfg

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"git.kirsle.net/apps/doodle/pkg/userdir"
)

// Settings are the available game settings.
type Settings struct {
	// Initialized is set true the first time the settings are saved to
	// disk, so the game may decide some default settings for first-time
	// user experience, e.g. set horizontal toolbars for mobile.
	Initialized bool

	// Configurable settings (pkg/windows/settings.go)
	HorizontalToolbars bool `json:",omitempty"`

	// Secret boolprops from balance/boolprops.go
	ShowHiddenDoodads bool `json:",omitempty"`
	WriteLockOverride bool `json:",omitempty"`
	JSONIndent        bool `json:",omitempty"`

	// Bookkeeping.
	UpdatedAt time.Time
}

// Current loaded settings, good defaults by default.
var Current = Defaults()

// Defaults returns sensible default user settings.
func Defaults() *Settings {
	return &Settings{}
}

// Filepath returns the path to the settings file.
func Filepath() string {
	return filepath.Join(userdir.ProfileDirectory, "settings.json")
}

// Save the settings to disk.
func Save() error {
	var (
		filename = Filepath()
		bin      = bytes.NewBuffer([]byte{})
		enc      = json.NewEncoder(bin)
	)
	enc.SetIndent("", "\t")
	Current.Initialized = true
	Current.UpdatedAt = time.Now()
	if err := enc.Encode(Current); err != nil {
		return err
	}
	err := ioutil.WriteFile(filename, bin.Bytes(), 0644)
	return err
}

// Load the settings from disk. The loaded settings will be available
// at usercfg.Current.
func Load() error {
	var (
		filename = Filepath()
		settings = Defaults()
	)
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return nil // no file, no problem
	}

	fh, err := os.Open(filename)
	if err != nil {
		return err
	}

	// Decode JSON from file.
	dec := json.NewDecoder(fh)
	err = dec.Decode(settings)
	if err != nil {
		return err
	}

	Current = settings
	return nil
}
