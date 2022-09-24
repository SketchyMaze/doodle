// Package license holds functions related to paid product activation.
package license

import (
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"io/ioutil"
	"path/filepath"

	"git.kirsle.net/SketchyMaze/doodle/pkg/userdir"
	"github.com/dgrijalva/jwt-go"
)

// Errors
var (
	ErrRegisteredFeature = errors.New("feature not available")
)

// Registration object encoded into a license key file.
type Registration struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	jwt.StandardClaims
}

// IsRegistered returns a boolean answer: is the product registered?
func IsRegistered() bool {
	if _, err := GetRegistration(); err == nil {
		return true
	}
	return false
}

// GetRegistration returns the currently registered user, by checking
// for the license.key file in the profile folder.
func GetRegistration() (Registration, error) {
	if Signer == nil {
		return Registration{}, errors.New("signer not ready")
	}

	filename := filepath.Join(userdir.ProfileDirectory, "license.key")
	jwt, err := ioutil.ReadFile(filename)
	if err != nil {
		return Registration{}, err
	}

	// Check if the JWT is valid.
	reg, err := Validate(Signer, string(jwt))
	if err != nil {
		return Registration{}, err
	}

	return reg, err
}

// UploadLicenseFile handles the user selecting the license key file, and it is
// validated and ingested.
func UploadLicenseFile(filename string) (Registration, error) {
	if Signer == nil {
		return Registration{}, errors.New("signer not ready")
	}

	jwt, err := ioutil.ReadFile(filename)
	if err != nil {
		return Registration{}, err
	}

	// Check if the JWT is valid.
	reg, err := Validate(Signer, string(jwt))
	if err != nil {
		return Registration{}, err
	}

	// Upload the license to Doodle's profile directory.
	outfile := filepath.Join(userdir.ProfileDirectory, "license.key")
	if err := ioutil.WriteFile(outfile, jwt, 0644); err != nil {
		return Registration{}, err
	}

	return reg, nil
}

// Validate the registration is signed by the appropriate public key.
func Validate(publicKey *ecdsa.PublicKey, tokenString string) (Registration, error) {
	var reg Registration
	token, err := jwt.ParseWithClaims(tokenString, &reg, func(token *jwt.Token) (interface{}, error) {
		return publicKey, nil
	})
	if err != nil {
		return reg, err
	}

	if !token.Valid {
		return reg, errors.New("token not valid")
	}
	return reg, nil
}

// ParsePublicKeyPEM loads a public key from PEM format.
func ParsePublicKeyPEM(keytext string) (*ecdsa.PublicKey, error) {
	blockPub, _ := pem.Decode([]byte(keytext))
	x509EncodedPub := blockPub.Bytes
	genericPublicKey, _ := x509.ParsePKIXPublicKey(x509EncodedPub)
	publicKey := genericPublicKey.(*ecdsa.PublicKey)
	return publicKey, nil
}
