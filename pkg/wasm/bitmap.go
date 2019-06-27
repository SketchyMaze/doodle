// +build !js

package wasm

import (
	"image"
	"os"

	"golang.org/x/image/bmp"
)

// StoreBitmap stores a bitmap image to disk.
func StoreBitmap(filename string, img image.Image) error {
	fh, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer fh.Close()
	return bmp.Encode(fh, img)
}
