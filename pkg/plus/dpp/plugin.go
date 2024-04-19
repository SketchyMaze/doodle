package dpp

import (
	"git.kirsle.net/SketchyMaze/doodle/pkg/doodads"
	"git.kirsle.net/SketchyMaze/doodle/pkg/filesystem"
	"git.kirsle.net/SketchyMaze/doodle/pkg/level"
	"git.kirsle.net/SketchyMaze/doodle/pkg/levelpack"
	"git.kirsle.net/SketchyMaze/doodle/pkg/plus"
)

var Driver Pluggable

// Plugin
type Pluggable interface {
	LoadFromEmbeddable(string, filesystem.Embeddable, bool) (*doodads.Doodad, error)
	IsRegistered() bool
	GetRegistration() (plus.Registration, error)
	UploadLicenseFile(string) (plus.Registration, error)
	IsLevelSigned(*level.Level) bool
	IsLevelPackSigned(*levelpack.LevelPack) bool
}
