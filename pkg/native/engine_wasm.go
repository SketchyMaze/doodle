// +build js,wasm

package native

import (
	"image"

	"git.kirsle.net/go/render"
)

func TextToImage(e render.Engine, text render.Text) (image.Image, error) {
	return nil, errors.New("not supported on WASM")
}
