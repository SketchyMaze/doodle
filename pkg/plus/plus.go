// Package plus connects the open source Doodle engine to the Doodle++ feature.
package plus

import (
	"errors"

	"git.kirsle.net/SketchyMaze/doodle/pkg/doodads"
	"git.kirsle.net/SketchyMaze/doodle/pkg/filesystem"
	"github.com/dgrijalva/jwt-go"
)

var ErrNotImplemented = errors.New("not implemented")

type Bridge interface {
	DoodadFromEmbeddable(filename string, fs filesystem.Embeddable, force bool) (*doodads.Doodad, error)
}

// Registration object encoded into a license key file.
type Registration struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	jwt.StandardClaims
}
