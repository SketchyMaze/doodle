/*
Package publishing contains functionality for "publishing" a Level, which
involves the writing and reading of custom doodads embedded inside
the levels.

Free tiers of the game will not read or write embedded doodads into
levels.
*/
package publishing

import (
	"fmt"
	"sort"

	"git.kirsle.net/apps/doodle/pkg/balance"
	"git.kirsle.net/apps/doodle/pkg/doodads"
	"git.kirsle.net/apps/doodle/pkg/level"
	"git.kirsle.net/apps/doodle/pkg/log"
)

/*
Level "Publishing" functions, involving the writing and reading of embedded
doodads within level files.
*/

// Publish writes a published level file, with embedded doodads included.
func Publish(lvl *level.Level, filename string, includeBuiltins bool) (*level.Level, error) {
	// Get and embed the doodads.
	builtins, customs := GetUsedDoodadNames(lvl)
	if includeBuiltins {
		log.Debug("including builtins: %+v", builtins)
		customs = append(customs, builtins...)
	}
	for _, filename := range customs {
		log.Debug("Embed filename: %s", filename)
		doodad, err := doodads.LoadFromEmbeddable(filename, lvl)
		if err != nil {
			return nil, fmt.Errorf("couldn't load doodad %s: %s", filename, err)
		}

		bin, err := doodad.Serialize()
		if err != nil {
			return nil, fmt.Errorf("couldn't serialize doodad %s: %s", filename, err)
		}

		// Embed it.
		lvl.SetFile(balance.EmbeddedDoodadsBasePath+filename, bin)
	}

	log.Info("Publish: write file to %s", filename)
	err := lvl.WriteFile(filename)
	return lvl, err
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
