package doodads

import (
	"fmt"
	"strconv"

	"git.kirsle.net/SketchyMaze/doodle/pkg/log"
)

// Options for runtime, user configurable.
type Option struct {
	Type    string      `json:"type"` // bool, str, int
	Name    string      `json:"name"`
	Default interface{} `json:"default"`
}

// SetOption sets an actor option, safely.
func (d *Doodad) SetOption(name, dataType, v string) string {
	if _, ok := d.Options[name]; !ok {
		d.Options[name] = &Option{
			Type: dataType,
			Name: name,
		}
	}

	return d.Options[name].Set(v)
}

// Set an option value. Generally do not call this yourself - use SetOption
// to safely set an option which will create the map value the first time.
func (o *Option) Set(v string) string {
	switch o.Type {
	case "bool":
		o.Default = v == "true"
	case "str":
		o.Default = v
	case "int":
		if val, err := strconv.Atoi(v); err != nil {
			log.Error("Doodad Option.Set: not an int: %v", val)
		} else {
			o.Default = val
		}
	default:
		log.Error("Doodad Option.Set: don't know how to set a %s type", o.Type)
	}
	return fmt.Sprintf("%v", o.Default)
}
