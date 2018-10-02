package doodle

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"git.kirsle.net/apps/doodle/render"
	"github.com/kirsle/configdir"
)

// Configuration constants.
var (
	DebugTextPadding int32 = 8
	DebugTextSize          = 24
	DebugTextColor         = render.SkyBlue
	DebugTextStroke        = render.Grey
	DebugTextShadow        = render.Black
)

// Profile Directory settings.
var (
	ConfigDirectoryName = "doodle"
	ProfileDirectory    string
	LevelDirectory      string
	DoodadDirectory     string

	// Regexp to match simple filenames for maps and doodads.
	reSimpleFilename = regexp.MustCompile(`^([A-Za-z0-9-_.,+ '"\[\](){}]+)$`)
)

// File extensions
const (
	extLevel  = ".level"
	extDoodad = ".doodad"
)

func init() {
	ProfileDirectory = configdir.LocalConfig(ConfigDirectoryName)
	LevelDirectory = configdir.LocalConfig(ConfigDirectoryName, "levels")
	DoodadDirectory = configdir.LocalConfig(ConfigDirectoryName, "doodads")
	configdir.MakePath(LevelDirectory, DoodadDirectory)
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

// resolvePath is the inner logic for LevelPath and DoodadPath.
func resolvePath(directory, filename, extension string) string {
	if strings.Contains(filename, "/") {
		return filename
	}

	// Attach the file extension?
	if strings.ToLower(filepath.Ext(filename)) != extension {
		filename += extension
	}

	return filepath.Join(directory, filename)
}

/*
EditFile opens a drawing file (Level or Doodad) in the EditorScene.

The filename can be one of the following:

	- A simple filename with no path separators in it and/or no file extension.
	- An absolute path beginning with "/"
	- A relative path beginning with "./"

If the filename has an extension (`.level` or `.doodad`), that will disambiguate
how to find the file and which mode to start the EditorMode in. Otherwise, the
"levels" folder is checked first and the "doodads" folder second.
*/
func (d *Doodle) EditFile(filename string) error {
	var absPath string

	// Is it a simple filename?
	if m := reSimpleFilename.FindStringSubmatch(filename); len(m) > 0 {
		log.Debug("EditFile: simple filename %s", filename)
		extension := strings.ToLower(filepath.Ext(filename))
		if foundFilename := d.ResolvePath(filename, extension, false); foundFilename != "" {
			log.Info("EditFile: resolved name '%s' to path %s", filename, foundFilename)
			absPath = foundFilename
		} else {
			return fmt.Errorf("EditFile: %s: no level or doodad found", filename)
		}
	} else {
		log.Debug("Not a simple: %s %+v", filename, reSimpleFilename)
		if _, err := os.Stat(filename); !os.IsNotExist(err) {
			log.Debug("EditFile: verified path %s exists", filename)
			absPath = filename
		}
	}

	d.EditDrawing(absPath)

	return nil
}

// ResolvePath takes an ambiguous simple filename and searches for a Level or
// Doodad that matches. Returns a blank string if no files found.
//
// Pass a true value for `one` if you are intending to create the file. It will
// only test one filepath and return the first one, regardless if the file
// existed. So the filename should have a ".level" or ".doodad" extension and
// then this path will resolve the ProfileDirectory of the file.
func (d *Doodle) ResolvePath(filename, extension string, one bool) string {
	// If the filename exists outright, return it.
	if _, err := os.Stat(filename); !os.IsNotExist(err) {
		return filename
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
		log.Debug("findFilename: try to find '%s' as %s", filename, test)
		if _, err := os.Stat(test); os.IsNotExist(err) {
			continue
		}
		return test
	}

	return ""
}
