package sdl

import (
	"bytes"
	"fmt"
	"image"

	"git.kirsle.net/apps/doodle/lib/render"
	"github.com/veandco/go-sdl2/sdl"
	"golang.org/x/image/bmp"
)

// Copy a texture into the renderer.
func (r *Renderer) Copy(t render.Texturer, src, dst render.Rect) {
	if tex, ok := t.(*Texture); ok {
		var (
			a = RectToSDL(src)
			b = RectToSDL(dst)
		)
		r.renderer.Copy(tex.tex, &a, &b)
	}
}

// Texture can hold on to SDL textures for caching and optimization.
type Texture struct {
	tex    *sdl.Texture
	width  int32
	height int32
}

// NewTexture caches an SDL texture from a bitmap.
func (r *Renderer) NewTexture(filename string, img image.Image) (render.Texturer, error) {
	var (
		fh = bytes.NewBuffer([]byte{})
	)

	err := bmp.Encode(fh, img)
	if err != nil {
		return nil, fmt.Errorf("NewTexture: bmp.Encode: %s", err)
	}

	// Create an SDL RWOps from the bitmap data in memory.
	sdlRW, err := sdl.RWFromMem(fh.Bytes())
	if err != nil {
		return nil, fmt.Errorf("NewTexture: sdl.RWFromMem: %s", err)
	}

	surface, err := sdl.LoadBMPRW(sdlRW, true)
	if err != nil {
		return nil, fmt.Errorf("NewTexture: sdl.LoadBMPRW: %s", err)
	}
	defer surface.Free()

	// TODO: chroma key color hardcoded to white here
	key := sdl.MapRGB(surface.Format, 255, 255, 255)
	surface.SetColorKey(true, key)

	texture, err := r.renderer.CreateTextureFromSurface(surface)
	if err != nil {
		return nil, fmt.Errorf("NewBitmap: create texture: %s", err)
	}

	tex := &Texture{
		width:  surface.W,
		height: surface.H,
		tex:    texture,
	}
	r.textures[filename] = tex

	return tex, nil
}

// Size returns the dimensions of the texture.
func (t *Texture) Size() render.Rect {
	return render.NewRect(t.width, t.height)
}

// NewBitmap initializes a texture from a bitmap image.
func (r *Renderer) NewBitmap(filename string) (render.Texturer, error) {
	if tex, ok := r.textures[filename]; ok {
		return tex, nil
	}
	return nil, fmt.Errorf("NewBitmap(%s): not found in texture cache", filename)
}
