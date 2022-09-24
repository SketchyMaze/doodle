package wallpaper

// The methods that deal in cached Textures for Doodle.

import (
	"errors"
	"image"

	"git.kirsle.net/SketchyMaze/doodle/pkg/shmem"
	"git.kirsle.net/SketchyMaze/doodle/pkg/userdir"
	"git.kirsle.net/go/render"
)

/*
Free all SDL2 textures from memory.

The Canvas widget will free wallpaper textures in its Destroy method. Note that
if the wallpaper was still somehow in use, the textures will be regenerated the
next time a method like CornerTexture() asks for one.

Returns the number of textures freed (up to 4) or -1 if wallpaper was not ready.
*/
func (wp *Wallpaper) Free() int {
	if !wp.ready {
		return -1
	}

	var freed int

	if wp.tex.corner != nil {
		wp.tex.corner.Free()
		wp.tex.corner = nil
		freed++
	}

	if wp.tex.top != nil {
		wp.tex.top.Free()
		wp.tex.top = nil
		freed++
	}

	if wp.tex.left != nil {
		wp.tex.left.Free()
		wp.tex.left = nil
		freed++
	}

	if wp.tex.repeat != nil {
		wp.tex.repeat.Free()
		wp.tex.repeat = nil
		freed++
	}

	return freed
}

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

// texture creates or returns a cached texture for a wallpaper.
func texture(e render.Engine, img *image.RGBA, name string) (render.Texturer, error) {
	filename := userdir.CacheFilename("wallpaper", name+".bmp")
	texture, err := shmem.CurrentRenderEngine.StoreTexture(filename, img)
	return texture, err
}
