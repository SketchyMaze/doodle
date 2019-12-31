package sprites

import (
	"bytes"
	"errors"
	"image/png"
	"io/ioutil"
	"os"
	"runtime"

	"git.kirsle.net/apps/doodle/pkg/bindata"
	"git.kirsle.net/apps/doodle/pkg/log"
	"git.kirsle.net/apps/doodle/pkg/wasm"
	"git.kirsle.net/go/render"
	"git.kirsle.net/go/ui"
)

// LoadImage loads a sprite as a ui.Image object. It checks Doodle's embedded
// bindata, then the filesystem before erroring out.
//
// NOTE: only .png images supported as of now. TODO
func LoadImage(e render.Engine, filename string) (*ui.Image, error) {
	// Try the bindata first.
	if data, err := bindata.Asset(filename); err == nil {
		log.Debug("sprites.LoadImage: %s from bindata", filename)

		img, err := png.Decode(bytes.NewBuffer(data))
		if err != nil {
			return nil, err
		}

		tex, err := e.StoreTexture(filename, img)
		if err != nil {
			return nil, err
		}

		return ui.ImageFromTexture(tex), nil
	}

	// WASM: try the file over HTTP ajax request.
	if runtime.GOOS == "js" {
		data, err := wasm.HTTPGet(filename)
		if err != nil {
			return nil, err
		}

		img, err := png.Decode(bytes.NewBuffer(data))
		if err != nil {
			return nil, err
		}

		tex, err := e.StoreTexture(filename, img)
		if err != nil {
			return nil, err
		}

		return ui.ImageFromTexture(tex), nil
	}

	// Then try the file system.
	if _, err := os.Stat(filename); !os.IsNotExist(err) {
		log.Debug("sprites.LoadImage: %s from filesystem", filename)

		data, err := ioutil.ReadFile(filename)
		if err != nil {
			return nil, err
		}

		img, err := png.Decode(bytes.NewBuffer(data))
		if err != nil {
			return nil, err
		}

		tex, err := e.StoreTexture(filename, img)
		if err != nil {
			return nil, err
		}

		return ui.ImageFromTexture(tex), nil
	}

	return nil, errors.New("no such sprite found")
}
