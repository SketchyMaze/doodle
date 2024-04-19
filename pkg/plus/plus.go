// Package plus connects the open source Doodle engine to the Doodle++ feature.
package plus

import (
	"errors"

	"github.com/dgrijalva/jwt-go"
)

// Errors
var (
	ErrNotImplemented    = errors.New("not implemented")
	ErrRegisteredFeature = errors.New("feature not available")
)

// Registration object encoded into a license key file.
type Registration struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	jwt.StandardClaims
}
