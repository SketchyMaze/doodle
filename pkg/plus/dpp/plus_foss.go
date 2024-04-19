//go:build !dpp
// +build !dpp

package dpp

import (
	"errors"

	"git.kirsle.net/SketchyMaze/doodle/pkg/doodads"
	"git.kirsle.net/SketchyMaze/doodle/pkg/filesystem"
	"git.kirsle.net/SketchyMaze/doodle/pkg/level"
	"git.kirsle.net/SketchyMaze/doodle/pkg/levelpack"
	"git.kirsle.net/SketchyMaze/doodle/pkg/plus"
)

var ErrNotImplemented = errors.New("not implemented")

type Plugin struct{}

func (Plugin) LoadFromEmbeddable(filename string, fs filesystem.Embeddable, force bool) (*doodads.Doodad, error) {
	return doodads.LoadFile(filename)
}

func (Plugin) IsRegistered() bool {
	return false
}

func (Plugin) GetRegistration() (plus.Registration, error) {
	return plus.Registration{}, ErrNotImplemented
}

func (Plugin) UploadLicenseFile(string) (plus.Registration, error) {
	return plus.Registration{}, ErrNotImplemented
}

func (Plugin) IsLevelPackSigned(*levelpack.LevelPack) bool {
	return false
}

func (Plugin) IsLevelSigned(*level.Level) bool {
	return false
}

func init() {
	Driver = Plugin{}
}
