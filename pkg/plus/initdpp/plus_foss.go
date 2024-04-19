//go:build !dpp
// +build !dpp

package plus

import (
	"git.kirsle.net/SketchyMaze/doodle/pkg/doodads"
	"git.kirsle.net/SketchyMaze/doodle/pkg/filesystem"
)

// DoodadFromEmbeddable may load a doodad from an embedding filesystem, such as a Level or LevelPack.
func DoodadFromEmbeddable(filename string, fs filesystem.Embeddable, force bool) (*doodads.Doodad, error) {
	return doodads.LoadFile(filename)
}

func IsRegistered() bool {
	return false
}

func GetRegistration() (*Registration, error) {
	return nil, ErrNotImplemented
}
