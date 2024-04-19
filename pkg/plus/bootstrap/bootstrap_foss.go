//go:build !dpp
// +build !dpp

package bootstrap

import (
	"git.kirsle.net/SketchyMaze/doodle/pkg/plus/dpp"
)

var Driver dpp.Pluggable

func InitPlugins() {
	Driver = dpp.Plugin{}
}
