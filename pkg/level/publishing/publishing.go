/*
Package publishing contains functionality for "publishing" a Level, which
involves the writing and reading of custom doodads embedded inside
the levels.

Free tiers of the game will not read or write embedded doodads into
levels.
*/
package publishing

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"git.kirsle.net/SketchyMaze/doodle/pkg/balance"
	"git.kirsle.net/SketchyMaze/doodle/pkg/doodads"
	"git.kirsle.net/SketchyMaze/doodle/pkg/level"
	"git.kirsle.net/SketchyMaze/doodle/pkg/log"
	"git.kirsle.net/SketchyMaze/doodle/pkg/plus/dpp"
)

/*
Level "Publishing" functions, involving the writing and reading of embedded
doodads within level files.
*/

// Publish writes a published level file, with embedded doodads included.
func Publish(lvl *level.Level) error {
	// Not embedding doodads?
	if !lvl.SaveDoodads {
		if removed := lvl.DeleteFiles(balance.EmbeddedDoodadsBasePath); removed > 0 {
			log.Info("Note: removed %d attached doodads because SaveDoodads is false", removed)
		}
		return nil
	}

	// Registered games only.
	if !balance.DPP || !dpp.Driver.IsRegistered() {
		return errors.New("only registered versions of the game can attach doodads to levels")
	}

	// Get and embed the doodads.
	builtins, customs := GetUsedDoodadNames(lvl)
	var names = map[string]interface{}{}
	if lvl.SaveBuiltins {
		log.Debug("including builtins: %+v", builtins)
		customs = append(customs, builtins...)
	}
	for _, filename := range customs {
		log.Debug("Embed filename: %s", filename)
		names[filename] = nil

		doodad, err := dpp.Driver.LoadFromEmbeddable(filename, lvl, false)
		if err != nil {
			return fmt.Errorf("couldn't load doodad %s: %s", filename, err)
		}

		bin, err := doodad.Serialize()
		if err != nil {
			return fmt.Errorf("couldn't serialize doodad %s: %s", filename, err)
		}

		// Embed it.
		lvl.SetFile(balance.EmbeddedDoodadsBasePath+filename, bin)
	}

	// Trim any doodads not currently in the level.
	for _, filename := range lvl.ListFilesAt(balance.EmbeddedDoodadsBasePath) {
		basename := strings.TrimPrefix(filename, balance.EmbeddedDoodadsBasePath)
		if _, ok := names[basename]; !ok {
			log.Debug("Remove embedded doodad %s (cleanup)", basename)
			lvl.DeleteFile(filename)
		}
	}

	return nil
}

// GetUsedDoodadNames returns the lists of doodad filenames in use in a level,
// bucketed by built-in or custom user doodads.
func GetUsedDoodadNames(lvl *level.Level) (builtin []string, custom []string) {
	// Collect all the doodad names in use in this level.
	unique := map[string]interface{}{}
	names := []string{}
	if lvl != nil {
		for _, actor := range lvl.Actors {
			if _, ok := unique[actor.Filename]; ok {
				continue
			}
			unique[actor.Filename] = nil
			names = append(names, actor.Filename)
		}
	}

	// Identify which of the doodads are built-ins.
	// builtin = []string{}
	builtinMap := map[string]interface{}{}
	// custom := []string{}
	if builtins, err := doodads.ListBuiltin(); err == nil {
		for _, filename := range builtins {
			if _, ok := unique[filename]; ok {
				builtin = append(builtin, filename)
				builtinMap[filename] = nil
			}
		}
	}
	for _, name := range names {
		if _, ok := builtinMap[name]; ok {
			continue
		}
		custom = append(custom, name)
	}

	sort.Strings(builtin)
	sort.Strings(custom)

	return builtin, custom
}
