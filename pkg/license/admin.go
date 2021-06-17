package license

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"io/ioutil"
	"time"

	"github.com/dgrijalva/jwt-go"
)

// AdminGenerateKeys generates the ECDSA public and private key pair for the admin
// side of creating signed license files.
func AdminGenerateKeys() (*ecdsa.PrivateKey, error) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	return privateKey, err
}

// AdminWriteKeys writes the admin signing key to .pem files on disk.
func AdminWriteKeys(key *ecdsa.PrivateKey, privateFile, publicFile string) error {
	// Encode the private key to PEM format.
	x509Encoded, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		return err
	}
	pemEncoded := pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: x509Encoded,
	})

	// Encode the public key to PEM format.
	x509EncodedPub, err := x509.MarshalPKIXPublicKey(key.Public())
	if err != nil {
		return err
	}
	pemEncodedPub := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: x509EncodedPub,
	})

	// Write the files.
	if err := ioutil.WriteFile(privateFile, pemEncoded, 0600); err != nil {
		return err
	}
	if err := ioutil.WriteFile(publicFile, pemEncodedPub, 0644); err != nil {
		return err
	}

	return nil
}

// AdminLoadPrivateKey loads the private key from disk.
func AdminLoadPrivateKey(privateFile string) (*ecdsa.PrivateKey, error) {
	// Read the private key file.
	pemEncoded, err := ioutil.ReadFile(privateFile)
	if err != nil {
		return nil, err
	}

	// Decode the private key.
	block, _ := pem.Decode([]byte(pemEncoded))
	x509Encoded := block.Bytes
	privateKey, _ := x509.ParseECPrivateKey(x509Encoded)
	return privateKey, nil
}

// AdminLoadPublicKey loads the private key from disk.
func AdminLoadPublicKey(publicFile string) (*ecdsa.PublicKey, error) {
	pemEncodedPub, err := ioutil.ReadFile(publicFile)
	if err != nil {
		return nil, err
	}

	// Decode the public key.
	blockPub, _ := pem.Decode([]byte(pemEncodedPub))
	x509EncodedPub := blockPub.Bytes
	genericPublicKey, _ := x509.ParsePKIXPublicKey(x509EncodedPub)
	publicKey := genericPublicKey.(*ecdsa.PublicKey)

	return publicKey, nil
}

// AdminSignRegistration signs the registration object.
func AdminSignRegistration(key *ecdsa.PrivateKey, reg Registration) (string, error) {
	reg.StandardClaims = jwt.StandardClaims{
		Issuer:    "Maze Admin",
		IssuedAt:  time.Now().Unix(),
		NotBefore: time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodES384, reg)
	signed, err := token.SignedString(key)
	if err != nil {
		return "", err
	}
	return signed, nil
}
