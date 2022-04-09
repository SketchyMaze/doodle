/*
Package sprites manages miscellaneous in-game sprites.

The sprites are relatively few for UI purposes. Their textures are
loaded ONE time and cached in this package for performance.
*/
package sprites

import (
	"bytes"
	"errors"
	"image/png"
	"io/ioutil"
	"os"
	"runtime"

	"git.kirsle.net/apps/doodle/assets"
	"git.kirsle.net/apps/doodle/pkg/log"
	"git.kirsle.net/apps/doodle/pkg/wasm"
	"git.kirsle.net/go/render"
	"git.kirsle.net/go/ui"
)

// Cache of loaded sprites.
var cache = map[string]*ui.Image{}

// FlushCache clears the sprites cache.
func FlushCache() {
	panic("TODO: free textures")
}

// LoadImage loads a sprite as a ui.Image object. It checks Doodle's embedded
// bindata, then the filesystem before erroring out.
//
// NOTE: only .png images supported as of now. TODO
func LoadImage(e render.Engine, filename string) (*ui.Image, error) {
	if cached, ok := cache[filename]; ok {
		return cached, nil
	}

	// Try the bindata first.
	if data, err := assets.Asset(filename); err == nil {
		log.Debug("sprites.LoadImage: %s from bindata", filename)

		img, err := png.Decode(bytes.NewBuffer(data))
		if err != nil {
			return nil, err
		}

		if image, err := ui.ImageFromImage(img); err == nil {
			cache[filename] = image
			return image, nil
		} else {
			return nil, err
		}
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

		if image, err := ui.ImageFromImage(img); err == nil {
			cache[filename] = image
			return image, nil
		} else {
			return nil, err
		}
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

		if image, err := ui.ImageFromImage(img); err == nil {
			cache[filename] = image
			return image, nil
		} else {
			return nil, err
		}
	}

	return nil, errors.New("no such sprite found")
}
