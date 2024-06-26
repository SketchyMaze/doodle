package userdir

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"git.kirsle.net/SketchyMaze/doodle/pkg/wasm"
	"github.com/kirsle/configdir"
)

// Profile Directory settings.
var (
	ConfigDirectoryName = "doodle"

	ProfileDirectory    string
	LevelDirectory      string
	LevelPackDirectory  string
	DoodadDirectory     string
	CampaignDirectory   string
	ScreenshotDirectory string
	SaveFile            string
	LogFile             string

	CacheDirectory string
	FontDirectory  string
)

// File extensions
const (
	extLevel     = ".level"
	extDoodad    = ".doodad"
	extLevelPack = ".levelpack"
)

func init() {
	// Profile directory contains the user's levels and doodads.
	ProfileDirectory = configdir.LocalConfig(ConfigDirectoryName)
	LevelDirectory = configdir.LocalConfig(ConfigDirectoryName, "levels")
	LevelPackDirectory = configdir.LocalConfig(ConfigDirectoryName, "levelpacks")
	DoodadDirectory = configdir.LocalConfig(ConfigDirectoryName, "doodads")
	CampaignDirectory = configdir.LocalConfig(ConfigDirectoryName, "campaigns")
	ScreenshotDirectory = configdir.LocalConfig(ConfigDirectoryName, "screenshots")
	SaveFile = configdir.LocalConfig(ConfigDirectoryName, "savegame.json")
	LogFile = configdir.LocalConfig(ConfigDirectoryName, "logfile.txt")

	// Cache directory to extract font files to.
	CacheDirectory = configdir.LocalCache(ConfigDirectoryName)
	FontDirectory = configdir.LocalCache(ConfigDirectoryName, "fonts")

	// Ensure all the directories exist.
	// WASM: do not make paths in wasm.
	if runtime.GOOS != "js" {
		configdir.MakePath(LevelDirectory)
		configdir.MakePath(LevelPackDirectory)
		configdir.MakePath(DoodadDirectory)
		configdir.MakePath(CampaignDirectory)
		configdir.MakePath(FontDirectory)
		configdir.MakePath(ScreenshotDirectory)
	}
}

// LevelPath will turn a "simple" filename into an absolute path in the user's
// local levels folder. If the filename already contains slashes, it is returned
// as-is as an absolute or relative path.
func LevelPath(filename string) string {
	return resolvePath(LevelDirectory, filename, extLevel)
}

// DoodadPath is like LevelPath but for Doodad files.
func DoodadPath(filename string) string {
	return resolvePath(DoodadDirectory, filename, extDoodad)
}

// LevelPackPath returns the user's levelpacks directory.
func LevelPackPath(filename string) string {
	return resolvePath(LevelPackDirectory, filename, extLevelPack)
}

// CacheFilename returns a path to a file in the cache folder. Send in path
// components and not literal slashes, like
// CacheFilename("images", "chunks", "id.bmp")
func CacheFilename(filename ...string) string {
	paths := append([]string{CacheDirectory}, filename...)
	dir := paths[:len(paths)-1]

	if runtime.GOOS != "js" {
		configdir.MakePath(filepath.Join(dir...))
	}
	return filepath.Join(paths[0], filepath.Join(paths[1:]...))
}

// ListDoodads returns a listing of all available doodads.
func ListDoodads() ([]string, error) {
	var names []string

	// WASM: list from localStorage.
	if runtime.GOOS == "js" {
		return wasm.StorageKeys(DoodadDirectory + "/"), nil
	}

	files, err := ioutil.ReadDir(DoodadDirectory)
	if err != nil {
		return names, err
	}

	for _, file := range files {
		name := file.Name()
		if strings.HasSuffix(strings.ToLower(name), extDoodad) {
			names = append(names, name)
		}
	}

	return names, nil
}

// ListLevels returns a listing of all available levels.
func ListLevels() ([]string, error) {
	var names []string

	// WASM: list from localStorage.
	if runtime.GOOS == "js" {
		return wasm.StorageKeys(LevelDirectory + "/"), nil
	}

	files, err := ioutil.ReadDir(LevelDirectory)
	if err != nil {
		return names, err
	}

	for _, file := range files {
		name := file.Name()
		if strings.HasSuffix(strings.ToLower(name), extLevel) {
			names = append(names, name)
		}
	}

	return names, nil
}

// ListCampaigns returns a listing of all available campaigns.
func ListCampaigns() ([]string, error) {
	var names []string

	// WASM: list from localStorage.
	if runtime.GOOS == "js" {
		return wasm.StorageKeys(CampaignDirectory + "/"), nil
	}

	files, err := ioutil.ReadDir(CampaignDirectory)
	if err != nil {
		return names, err
	}

	for _, file := range files {
		name := file.Name()
		if filepath.Ext(name) == ".json" {
			names = append(names, name)
		}
	}

	return names, nil
}

// resolvePath is the inner logic for LevelPath and DoodadPath.
func resolvePath(directory, filename, extension string) string {
	if strings.Contains(filename, string(filepath.Separator)) {
		return filename
	}

	// Attach the file extension?
	if strings.ToLower(filepath.Ext(filename)) != extension {
		filename += extension
	}

	return filepath.Join(directory, filename)
}

// ResolvePath takes an ambiguous simple filename and searches for a Level or
// Doodad that matches. Returns a blank string if no files found.
//
// Pass a true value for `one` if you are intending to create the file. It will
// only test one filepath and return the first one, regardless if the file
// existed. So the filename should have a ".level" or ".doodad" extension and
// then this path will resolve the ProfileDirectory of the file.
func ResolvePath(filename, extension string, one bool) string {
	// If the filename exists outright, return it.
	if !(runtime.GOOS == "js") {
		if _, err := os.Stat(filename); !os.IsNotExist(err) {
			return filename
		}
	}

	var paths []string
	if extension == extLevel {
		paths = append(paths, filepath.Join(LevelDirectory, filename))
	} else if extension == extDoodad {
		paths = append(paths, filepath.Join(DoodadDirectory, filename))
	} else {
		paths = append(paths,
			filepath.Join(LevelDirectory, filename+".level"),
			filepath.Join(DoodadDirectory, filename+".doodad"),
		)
	}

	for _, test := range paths {
		// WASM: check the path in localStorage.
		if runtime.GOOS == "js" {
			if _, ok := wasm.GetSession(test); ok {
				return test
			}
			continue
		}

		// Desktop: test the filesystem.
		if _, err := os.Stat(test); os.IsNotExist(err) {
			continue
		}
		return test
	}

	return ""
}
