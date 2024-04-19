package dpp

import (
	"git.kirsle.net/SketchyMaze/doodle/pkg/doodads"
	"git.kirsle.net/SketchyMaze/doodle/pkg/filesystem"
	"git.kirsle.net/SketchyMaze/doodle/pkg/level"
	"git.kirsle.net/SketchyMaze/doodle/pkg/levelpack"
	"git.kirsle.net/SketchyMaze/doodle/pkg/plus"
)

// Driver is the currently installed Doodle++ implementation (FOSS or DPP).
var Driver Pluggable

// Pluggable defines the interface for Doodle++ functions, so that their implementations
// can avoid cyclic dependency errors. Documentation for these functions is only spelled
// out in the SketchyMaze/dpp package.
type Pluggable interface {
	LoadFromEmbeddable(string, filesystem.Embeddable, bool) (*doodads.Doodad, error)
	IsRegistered() bool
	GetRegistration() (plus.Registration, error)
	UploadLicenseFile(string) (plus.Registration, error)
	IsLevelSigned(*level.Level) bool
	IsLevelPackSigned(*levelpack.LevelPack) bool
}
