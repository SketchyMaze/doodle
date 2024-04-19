//go:build dpp
// +build dpp

package plus

import (
	"git.kirsle.net/SketchyMaze/doodle/pkg/doodads"
	"git.kirsle.net/SketchyMaze/doodle/pkg/filesystem"
	"git.kirsle.net/SketchyMaze/dpp/embedding"
	"git.kirsle.net/SketchyMaze/dpp/license"
)

func DoodadFromEmbeddable(filename string, fs filesystem.Embeddable, force bool) (*doodads.Doodad, error) {
	return embedding.LoadFromEmbeddable(filename, fs, force)
}

func IsRegistered() bool {
	return license.IsRegistered()
}

func GetRegistration() (*Registration, error) {
	reg, err := license.GetRegistration()
	return reg.(*Registration), err
}
