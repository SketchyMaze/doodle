//go:build dpp
// +build dpp

/*
Package bootstrap is a common import between the Doodle and Doodad programs.

Its chief job is to work around circular dependency issues when dealing with
pluggable parts of the codebase, such as Doodle++ which adds features for the
official release which are missing from the FOSS version.
*/
package bootstrap

import (
	"git.kirsle.net/SketchyMaze/doodle/pkg/plus/dpp"
)

var Driver dpp.Pluggable

func InitPlugins() {
	Driver = dpp.Plugin{}
}
