package wallpaper

import (
	"os"
	"io/ioutil"
	"encoding/base64"
)

/*
Functions related to the ingest of custom Wallpaper images for user levels.
*/

// FileToB64 loads an image file from disk and returns the Base64 encoded
// file data, if it is a valid image and so on.
func FileToB64(filename string) (string, error) {
	fh, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer fh.Close()

	bin, err := ioutil.ReadAll(fh)
	if err != nil {
		return "", err
	}

	b64 := base64.StdEncoding.EncodeToString(bin)
	if err != nil {
		return "", err
	}

	return b64, nil
}