package sdl

import (
	"fmt"

	"git.kirsle.net/apps/doodle/render"
	"github.com/veandco/go-sdl2/sdl"
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

// Size returns the dimensions of the texture.
func (t *Texture) Size() render.Rect {
	return render.NewRect(t.width, t.height)
}

// NewBitmap initializes a texture from a bitmap image.
func (r *Renderer) NewBitmap(filename string) (render.Texturer, error) {
	surface, err := sdl.LoadBMP(filename)
	if err != nil {
		return nil, fmt.Errorf("NewBitmap: LoadBMP: %s", err)
	}
	defer surface.Free()

	tex, err := r.renderer.CreateTextureFromSurface(surface)
	if err != nil {
		return nil, fmt.Errorf("NewBitmap: create texture: %s", err)
	}

	return &Texture{
		width:  surface.W,
		height: surface.H,
		tex:    tex,
	}, nil
}
