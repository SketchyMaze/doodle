// Package campaign contains types and functions for the single player campaigns.
package campaign

import (
	"io/ioutil"
	"path/filepath"
	"runtime"
	"sort"

	"git.kirsle.net/apps/doodle/pkg/bindata"
	"git.kirsle.net/apps/doodle/pkg/filesystem"
	"git.kirsle.net/apps/doodle/pkg/userdir"
)

// List returns the list of available campaign JSONs.
//
// It searches in:
// - The embedded bindata for built-in scenarios
// - Scenarios on disk at the assets/campaigns folder.
// - User-made scenarios at ~/doodle/campaigns.
func List() ([]string, error) {
	var names []string

	// List built-in bindata campaigns.
	if files, err := bindata.AssetDir("assets/campaigns"); err == nil {
		names = append(names, files...)
	}

	// WASM: only built-in campaigns, no filesystem access.
	if runtime.GOOS == "js" {
		return names, nil
	}

	// Read system-level doodads first. Ignore errors, if the system path is
	// empty we still go on to read the user directory.
	files, _ := ioutil.ReadDir(filesystem.SystemCampaignsPath)
	for _, file := range files {
		name := file.Name()
		if filepath.Ext(name) == ".json" {
			names = append(names, name)
		}
	}

	// Append user campaigns.
	userFiles, err := userdir.ListCampaigns()
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
