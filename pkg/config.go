package doodle

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"git.kirsle.net/apps/doodle/pkg/log"
	"git.kirsle.net/apps/doodle/pkg/userdir"
)

// Regexp to match simple filenames for maps and doodads.
var reSimpleFilename = regexp.MustCompile(`^([A-Za-z0-9-_.,+ '"\[\](){}]+)$`)

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
		if foundFilename := userdir.ResolvePath(filename, extension, false); foundFilename != "" {
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
