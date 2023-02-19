package levelsigning

import (
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io/ioutil"

	"git.kirsle.net/SketchyMaze/doodle/pkg/level"
	"git.kirsle.net/SketchyMaze/doodle/pkg/levelpack"
	"git.kirsle.net/SketchyMaze/doodle/pkg/license"
	"git.kirsle.net/SketchyMaze/doodle/pkg/log"
)

// IsLevelSigned returns a quick answer.
func IsLevelSigned(lvl *level.Level) bool {
	return VerifyLevel(license.Signer, lvl)
}

// IsLevelPackSigned returns a quick answer.
func IsLevelPackSigned(lp *levelpack.LevelPack) bool {
	return VerifyLevelPack(license.Signer, lp)
}

/*
SignLevel creates a signature on a level file which allows it to load its
embedded doodads even for free versions of the game.

Free versions will verify a level's signature before bailing out with the
"can't play levels w/ embedded doodads" response.

NOTE: this only supported Zipfile levels and will assume the level you
pass has a Zipfile to access embedded assets.
*/
func SignLevel(key *ecdsa.PrivateKey, lvl *level.Level) ([]byte, error) {
	// Encode the attached files data to deterministic JSON.
	certificate, err := StringifyAssets(lvl)
	if err != nil {
		return nil, err
	}

	log.Info("Sign file tree: %s", certificate)
	digest := shasum(certificate)

	signature, err := ecdsa.SignASN1(rand.Reader, key, digest)
	if err != nil {
		return nil, err
	}
	log.Info("Digest: %x  Signature: %x", digest, signature)

	return signature, nil
}

// VerifyLevel verifies a level's signature and returns if it is OK.
func VerifyLevel(publicKey *ecdsa.PublicKey, lvl *level.Level) bool {
	// No signature = not verified.
	if lvl.Signature == nil || len(lvl.Signature) == 0 {
		return false
	}

	// Encode the attached files data to deterministic JSON.
	certificate, err := StringifyAssets(lvl)
	if err != nil {
		log.Error("VerifyLevel: couldn't stringify assets: %s", err)
		return false
	}

	digest := shasum(certificate)

	// Verify the signature against our public key.
	return ecdsa.VerifyASN1(publicKey, digest, lvl.Signature)
}

/*
SignLevelpack applies a signature to a levelpack as a whole, to allow its
shared custom doodads to be loaded by its levels in free games.
*/
func SignLevelPack(key *ecdsa.PrivateKey, lp *levelpack.LevelPack) ([]byte, error) {
	// Encode the attached files data to deterministic JSON.
	certificate, err := StringifyLevelpackAssets(lp)
	if err != nil {
		return nil, err
	}

	log.Info("Sign file tree: %s", certificate)
	digest := shasum(certificate)

	signature, err := ecdsa.SignASN1(rand.Reader, key, digest)
	if err != nil {
		return nil, err
	}
	log.Info("Digest: %x  Signature: %x", digest, signature)

	return signature, nil
}

// VerifyLevelPack verifies a levelpack's signature and returns if it is OK.
func VerifyLevelPack(publicKey *ecdsa.PublicKey, lp *levelpack.LevelPack) bool {
	// No signature = not verified.
	if lp.Signature == nil || len(lp.Signature) == 0 {
		return false
	}

	// Encode the attached files data to deterministic JSON.
	certificate, err := StringifyLevelpackAssets(lp)
	if err != nil {
		log.Error("VerifyLevelPack: couldn't stringify assets: %s", err)
		return false
	}

	digest := shasum(certificate)

	// Verify the signature against our public key.
	return ecdsa.VerifyASN1(publicKey, digest, lp.Signature)
}

// StringifyAssets creates the signing checksum of a level's attached assets.
func StringifyAssets(lvl *level.Level) ([]byte, error) {
	// Get a listing of all embedded files. Note: gives us a conveniently
	// sorted array of files too.
	files := lvl.Files.List()

	// Pair each filename with its SHA256 sum.
	var checksum = map[string]string{}
	for _, filename := range files {
		if sum, err := lvl.Files.Checksum(filename); err != nil {
			return nil, fmt.Errorf("when checksum %s got error: %s", filename, err)
		} else {
			checksum[filename] = sum
		}
	}

	// Encode the payload to deterministic JSON.
	certificate, err := json.Marshal(checksum)
	if err != nil {
		return nil, err
	}

	return certificate, nil
}

// StringifyLevelpackAssets creates the signing checksum of a level's attached assets.
func StringifyLevelpackAssets(lp *levelpack.LevelPack) ([]byte, error) {
	var (
		files = []string{}
		seen  = map[string]struct{}{}
	)

	// Enumerate the files in the zipfile assets/ folder.
	for _, file := range lp.Zipfile.File {
		if file.Name == "index.json" {
			continue
		}

		if _, ok := seen[file.Name]; !ok {
			files = append(files, file.Name)
			seen[file.Name] = struct{}{}
		}
	}

	// Pair each filename with its SHA256 sum.
	var checksum = map[string]string{}
	for _, filename := range files {
		file, err := lp.Zipfile.Open(filename)
		if err != nil {
			return nil, err
		}

		bin, err := ioutil.ReadAll(file)
		if err != nil {
			return nil, err
		}

		checksum[filename] = fmt.Sprintf("%x", shasum(bin))
	}

	// Encode the payload to deterministic JSON.
	certificate, err := json.Marshal(checksum)
	if err != nil {
		return nil, err
	}

	return certificate, nil
}

// Common function to SHA-256 checksum a thing.
func shasum(data []byte) []byte {
	h := sha256.New()
	h.Write(data)
	return h.Sum(nil)
}
