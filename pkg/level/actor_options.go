package level

import (
	"fmt"
	"strconv"

	"git.kirsle.net/SketchyMaze/doodle/pkg/log"
)

// Option for runtime, user configurable overrides of Doodad Options.
type Option struct {
	Type  string      `json:"type"` // bool, str, int
	Name  string      `json:"name"`
	Value interface{} `json:"value"`
}

// SetOption sets an actor option, safely.
func (a *Actor) SetOption(name, dataType, v string) string {
	if _, ok := a.Options[name]; !ok {
		a.Options[name] = &Option{
			Type: dataType,
			Name: name,
		}
	}

	return a.Options[name].Set(v)
}

// Set an option value. Generally do not call this yourself - use SetOption
// to safely set an option which will create the map value the first time.
func (o *Option) Set(v string) string {
	switch o.Type {
	case "bool":
		o.Value = v == "true"
	case "str":
		o.Value = v
	case "int":
		if val, err := strconv.Atoi(v); err != nil {
			log.Error("Actor Option.Set: not an int: %v", val)
			o.Value = 0
		} else {
			o.Value = val
		}
	default:
		log.Error("Actor Option.Set: don't know how to set a %s type", o.Type)
	}
	return fmt.Sprintf("%v", o.Value)
}
