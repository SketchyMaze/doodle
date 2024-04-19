//go:build !dpp
// +build !dpp

package dpp

import (
	"git.kirsle.net/SketchyMaze/doodle/pkg/doodads"
	"git.kirsle.net/SketchyMaze/doodle/pkg/filesystem"
	"git.kirsle.net/SketchyMaze/doodle/pkg/level"
	"git.kirsle.net/SketchyMaze/doodle/pkg/levelpack"
	"git.kirsle.net/SketchyMaze/doodle/pkg/plus"
)

type Plugin struct{}

func (Plugin) LoadFromEmbeddable(filename string, fs filesystem.Embeddable, force bool) (*doodads.Doodad, error) {
	if result, err := doodads.LoadFile(filename); err != nil {
		return nil, plus.ErrRegisteredFeature
	} else {
		return result, nil
	}
}

func (Plugin) IsRegistered() bool {
	return false
}

func (Plugin) GetRegistration() (plus.Registration, error) {
	return plus.Registration{}, plus.ErrNotImplemented
}

func (Plugin) UploadLicenseFile(string) (plus.Registration, error) {
	return plus.Registration{}, plus.ErrNotImplemented
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
