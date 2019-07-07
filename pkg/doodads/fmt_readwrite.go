package doodads

import (
	"fmt"
	"io/ioutil"
	"runtime"
	"sort"
	"strings"

	"git.kirsle.net/apps/doodle/pkg/bindata"
	"git.kirsle.net/apps/doodle/pkg/branding"
	"git.kirsle.net/apps/doodle/pkg/enum"
	"git.kirsle.net/apps/doodle/pkg/filesystem"
	"git.kirsle.net/apps/doodle/pkg/log"
	"git.kirsle.net/apps/doodle/pkg/userdir"
	"git.kirsle.net/apps/doodle/pkg/wasm"
)

// ListDoodads returns a listing of all available doodads between all locations,
// including user doodads.
func ListDoodads() ([]string, error) {
	var names []string

	// List doodads embedded into the binary.
	if files, err := bindata.AssetDir("assets/doodads"); err == nil {
		names = append(names, files...)
	}

	// WASM
	if runtime.GOOS == "js" {
		// Return the array of doodads embedded in the bindata.
		// TODO: append user doodads to the list.
		return names, nil
	}

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

	// Deduplicate names.
	var uniq = map[string]interface{}{}
	var result []string
	for _, name := range names {
		if _, ok := uniq[name]; !ok {
			uniq[name] = nil
			result = append(result, name)
		}
	}

	sort.Strings(result)

	return result, err
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

	// Do we have the file in bindata?
	if jsonData, err := bindata.Asset(filename); err == nil {
		return FromJSON(filename, jsonData)
	}

	// WASM: try the file over HTTP ajax request.
	if runtime.GOOS == "js" {
		if result, ok := wasm.GetSession(filename); ok {
			log.Info("recall doodad data from localStorage")
			return FromJSON(filename, []byte(result))
		}

		// TODO: ajax load for doodads might not work, filesystem.FindFile returns
		// the base file for WASM but for now force it to system doodads path
		filename = "assets/doodads/" + filename
		jsonData, err := wasm.HTTPGet(filename)
		if err != nil {
			return nil, err
		}

		return FromJSON(filename, jsonData)
	}

	// Load the JSON file from the filesystem.
	return LoadJSON(filename)
}

// WriteFile saves a doodad to disk in the user's config directory.
func (d *Doodad) WriteFile(filename string) error {
	if !strings.HasSuffix(filename, enum.DoodadExt) {
		filename += enum.DoodadExt
	}

	// Set the version information.
	d.Version = 1
	d.GameVersion = branding.Version

	bin, err := d.ToJSON()
	if err != nil {
		return err
	}

	// WASM: place in localStorage.
	if runtime.GOOS == "js" {
		log.Info("wasm: write %s to localStorage", filename)
		wasm.SetSession(filename, string(bin))
		return nil
	}

	// Desktop: write to disk.
	filename = userdir.DoodadPath(filename)
	log.Debug("Write Doodad: %s", filename)
	err = ioutil.WriteFile(filename, bin, 0644)
	if err != nil {
		return fmt.Errorf("doodads.WriteFile: %s", err)
	}

	return nil
}
