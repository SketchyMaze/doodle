package wallpaper

// The methods that deal in cached Textures for Doodle.

import (
	"errors"
	"fmt"
	"image"
	"os"

	"git.kirsle.net/apps/doodle/lib/render"
	"git.kirsle.net/apps/doodle/pkg/userdir"
	"golang.org/x/image/bmp"
)

// CornerTexture returns the Texture.
func (wp *Wallpaper) CornerTexture(e render.Engine) (render.Texturer, error) {
	if !wp.ready {
		return nil, errors.New("wallpaper not ready")
	}

	if wp.tex.corner == nil {
		tex, err := texture(e, wp.corner, wp.Name+"c")
		wp.tex.corner = tex
		return tex, err
	}
	return wp.tex.corner, nil
}

// TopTexture returns the Texture.
func (wp *Wallpaper) TopTexture(e render.Engine) (render.Texturer, error) {
	if wp.tex.top == nil {
		tex, err := texture(e, wp.top, wp.Name+"t")
		wp.tex.top = tex
		return tex, err
	}
	return wp.tex.top, nil
}

// LeftTexture returns the Texture.
func (wp *Wallpaper) LeftTexture(e render.Engine) (render.Texturer, error) {
	if wp.tex.left == nil {
		tex, err := texture(e, wp.left, wp.Name+"l")
		wp.tex.left = tex
		return tex, err
	}
	return wp.tex.left, nil
}

// RepeatTexture returns the Texture.
func (wp *Wallpaper) RepeatTexture(e render.Engine) (render.Texturer, error) {
	if wp.tex.repeat == nil {
		tex, err := texture(e, wp.repeat, wp.Name+"x")
		wp.tex.repeat = tex
		return tex, err
	}
	return wp.tex.repeat, nil
}

func texture(e render.Engine, img *image.RGBA, name string) (render.Texturer, error) {
	filename := userdir.CacheFilename("wallpaper", name+".bmp")
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		fh, err := os.Create(filename)
		if err != nil {
			return nil, fmt.Errorf("texture(%s): %s", name, err.Error())
		}

		err = bmp.Encode(fh, img)
		if err != nil {
			return nil, fmt.Errorf("texture(%s): bmp.Encode: %s", name, err.Error())
		}

		err = fh.Close()
		if err != nil {
			return nil, fmt.Errorf("texture(%s): fh.Close: %s", name, err.Error())
		}
	}

	texture, err := e.NewBitmap(filename)
	if err != nil {
		return nil, fmt.Errorf("CornerTexture: NewBitmap(%s): %s", filename, err.Error())
	}
	return texture, nil
}
