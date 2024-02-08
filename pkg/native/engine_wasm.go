//go:build js && wasm
// +build js,wasm

package native

import (
	"errors"
	"image"

	"git.kirsle.net/go/render"
)

func HasTouchscreen(e render.Engine) bool {
	return false
}

func TextToImage(e render.Engine, text render.Text) (image.Image, error) {
	return nil, errors.New("not supported on WASM")
}

func CopyToClipboard(text string) error {
	return errors.New("not supported on WASM")
}

func CountTextures(e render.Engine) string {
	return "n/a"
}

func FreeTextures() {}

func MaximizeWindow(e render.Engine) {}
