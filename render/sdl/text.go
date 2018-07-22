package sdl

import (
	"git.kirsle.net/apps/doodle/render"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

var fonts map[int]*ttf.Font = map[int]*ttf.Font{}

// LoadFont loads and caches the font at a given size.
func LoadFont(size int) (*ttf.Font, error) {
	if font, ok := fonts[size]; ok {
		return font, nil
	}

	font, err := ttf.OpenFont("./fonts/DejaVuSansMono.ttf", size)
	if err != nil {
		return nil, err
	}
	fonts[size] = font

	return font, nil
}

// DrawText draws text on the canvas.
func (r *Renderer) DrawText(text render.Text, rect render.Rect) error {
	var (
		font    *ttf.Font
		surface *sdl.Surface
		tex     *sdl.Texture
		err     error
	)

	if font, err = LoadFont(text.Size); err != nil {
		return err
	}

	write := func(dx, dy int32, color sdl.Color) {
		if surface, err = font.RenderUTF8Blended(text.Text, color); err != nil {
			return
		}
		defer surface.Free()

		if tex, err = r.renderer.CreateTextureFromSurface(surface); err != nil {
			return
		}
		defer tex.Destroy()

		tmp := &sdl.Rect{
			X: rect.X + dx,
			Y: rect.Y + dy,
			W: surface.W,
			H: surface.H,
		}
		r.renderer.Copy(tex, nil, tmp)
	}

	// Does the text have a stroke around it?
	if text.Stroke != render.Invisible {
		color := ColorToSDL(text.Stroke)
		write(-1, -1, color)
		write(-1, 0, color)
		write(-1, 1, color)
		write(1, -1, color)
		write(1, 0, color)
		write(1, 1, color)
		write(0, -1, color)
		write(0, 1, color)
	}

	// Does it have a drop shadow?
	if text.Shadow != render.Invisible {
		write(1, 1, ColorToSDL(text.Shadow))
	}

	// Draw the text itself.
	write(0, 0, ColorToSDL(text.Color))

	return err
}
