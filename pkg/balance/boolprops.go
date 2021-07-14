package balance

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"git.kirsle.net/apps/doodle/pkg/usercfg"
)

/*
Boolprop is a boolean setting that can be toggled in the game using the
developer console. Many of these consist of usercfg settings that are
not exposed to the Settings UI window, and secret testing functions.
Where one points to usercfg, check usercfg.Settings for documentation
about what that boolean does.
*/
type Boolprop struct {
	Name string
	Get  func() bool
	Set  func(bool)
}

// Boolprops are the map of available boolprops, shown in the dev
// console when you type: "boolProp list"
var Boolprops = map[string]Boolprop{
	"show-hidden-doodads": {
		Get: func() bool { return usercfg.Current.ShowHiddenDoodads },
		Set: func(v bool) { usercfg.Current.ShowHiddenDoodads = v },
	},
	"write-lock-override": {
		Get: func() bool { return usercfg.Current.WriteLockOverride },
		Set: func(v bool) { usercfg.Current.WriteLockOverride = v },
	},
	"pretty-json": {
		Get: func() bool { return usercfg.Current.JSONIndent },
		Set: func(v bool) { usercfg.Current.JSONIndent = v },
	},
	"horizontal-toolbars": {
		Get: func() bool { return usercfg.Current.HorizontalToolbars },
		Set: func(v bool) { usercfg.Current.HorizontalToolbars = v },
	},
	"compress-drawings": {
		Get: func() bool { return CompressDrawings },
		Set: func(v bool) { CompressDrawings = v },
	},
}

// GetBoolProp reads the current value of a boolProp.
// Special value "list" will error out with a list of available props.
func GetBoolProp(name string) (bool, error) {
	if name == "list" {
		var keys []string
		for k := range Boolprops {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		return false, fmt.Errorf(
			"boolprops: %s",
			strings.Join(keys, ", "),
		)
	}
	if prop, ok := Boolprops[name]; ok {
		return prop.Get(), nil
	}
	return false, errors.New("no such boolProp")
}

// BoolProp allows easily setting a boolProp by name.
func BoolProp(name string, v bool) error {
	if prop, ok := Boolprops[name]; ok {
		prop.Set(v)
		return nil
	}
	return errors.New("no such boolProp")
}
