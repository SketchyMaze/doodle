//go:build dpp
// +build dpp

package dpp

import (
	"encoding/json"

	"git.kirsle.net/SketchyMaze/doodle/pkg/doodads"
	"git.kirsle.net/SketchyMaze/doodle/pkg/filesystem"
	"git.kirsle.net/SketchyMaze/doodle/pkg/level"
	"git.kirsle.net/SketchyMaze/doodle/pkg/levelpack"
	"git.kirsle.net/SketchyMaze/doodle/pkg/native"
	"git.kirsle.net/SketchyMaze/doodle/pkg/plus"
	"git.kirsle.net/SketchyMaze/dpp/embedding"
	"git.kirsle.net/SketchyMaze/dpp/license"
	"git.kirsle.net/SketchyMaze/dpp/license/levelsigning"
)

type Plugin struct{}

func (Plugin) LoadFromEmbeddable(filename string, fs filesystem.Embeddable, force bool) (*doodads.Doodad, error) {
	return embedding.LoadFromEmbeddable(filename, fs, force)
}

func (Plugin) IsRegistered() bool {
	return license.IsRegistered()
}

func (Plugin) GetRegistration() (plus.Registration, error) {
	reg, err := license.GetRegistration()
	if err != nil {
		return plus.Registration{}, err
	}

	return translateLicenseStruct(reg)
}

func (Plugin) UploadLicenseFile(filename string) (plus.Registration, error) {
	reg, err := license.UploadLicenseFile(filename)
	if err != nil {
		return plus.Registration{}, err
	}

	return translateLicenseStruct(reg)
}

// Hack: to translate JWT token types, easiest is to just encode/decode them (inner jwt.StandardClaims complexity).
func translateLicenseStruct(reg license.Registration) (plus.Registration, error) {
	// Set the DefaultAuthor to the registered user's name.
	if reg.Name != "" {
		native.DefaultAuthor = reg.Name
	}

	// Marshal to JSON and back to cast the type.
	var (
		result       plus.Registration
		jsonStr, err = json.Marshal(reg)
	)
	if err != nil {
		return plus.Registration{}, err
	}
	err = json.Unmarshal(jsonStr, &result)
	return result, err
}

func (Plugin) IsLevelPackSigned(lp *levelpack.LevelPack) bool {
	return levelsigning.IsLevelPackSigned(lp)
}

func (Plugin) IsLevelSigned(lvl *level.Level) bool {
	return levelsigning.IsLevelSigned(lvl)
}

func init() {
	Driver = Plugin{}
}
