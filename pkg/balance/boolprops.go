package balance

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

// Fun bool props to wreak havoc in the game.
var (
	// Force show hidden doodads in the palette in Editor Mode.
	ShowHiddenDoodads bool

	// Force ability to edit Locked levels and doodads.
	WriteLockOverride bool

	// Pretty-print JSON files when writing.
	JSONIndent bool
)

// Human friendly names for the boolProps. Not necessarily the long descriptive
// variable names above.
var props = map[string]*bool{
	"showAllDoodads":    &ShowHiddenDoodads,
	"writeLockOverride": &WriteLockOverride,
	"prettyJSON":        &JSONIndent,

	// WARNING: SLOW!
	"disableChunkTextureCache": &DisableChunkTextureCache,
}

// GetBoolProp reads the current value of a boolProp.
// Special value "list" will error out with a list of available props.
func GetBoolProp(name string) (bool, error) {
	if name == "list" {
		var keys []string
		for k := range props {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		return false, fmt.Errorf(
			"Boolprops: %s",
			strings.Join(keys, ", "),
		)
	}
	if prop, ok := props[name]; ok {
		return *prop, nil
	}
	return false, errors.New("no such boolProp")
}

// BoolProp allows easily setting a boolProp by name.
func BoolProp(name string, v bool) error {
	if prop, ok := props[name]; ok {
		*prop = v
		return nil
	}
	return errors.New("no such boolProp")
}
