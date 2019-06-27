// +build js,wasm

package wasm

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/png"
)

// StoreBitmap stores a bitmap image to sessionStorage as a data URL for PNG
// base64 encoded image.
func StoreBitmap(filename string, img image.Image) error {
	var fh = bytes.NewBuffer([]byte{})

	if err := png.Encode(fh, img); err != nil {
		return err
	}

	var dataURI = "data:image/png;base64," + base64.StdEncoding.EncodeToString(fh.Bytes())

	SetSession(filename, dataURI)
	return nil
}
