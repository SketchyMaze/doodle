package render

import (
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

// TextConfig are settings for rendered text.
type TextConfig struct {
	Text        string
	Size        int
	Color       sdl.Color
	StrokeColor sdl.Color
	X           int32
	Y           int32
	W           int32
	H           int32
}

// StrokedText draws text with a stroke color around it.
func StrokedText(t TextConfig) {
	stroke := func(copy TextConfig, x, y int32) {
		copy.Color = t.StrokeColor
		copy.X += x
		copy.Y += y
		Text(copy)
	}

	stroke(t, -1, -1)
	stroke(t, -1, 0)
	stroke(t, -1, 1)

	stroke(t, 1, -1)
	stroke(t, 1, 0)
	stroke(t, 1, 1)

	stroke(t, 0, -1)
	stroke(t, 0, 1)
	Text(t)
}

// Text draws text on the renderer.
func Text(t TextConfig) error {
	var (
		font    *ttf.Font
		surface *sdl.Surface
		tex     *sdl.Texture
		err     error
	)

	if font, err = LoadFont(t.Size); err != nil {
		return err
	}

	if surface, err = font.RenderUTF8Blended(t.Text, t.Color); err != nil {
		return err
	}
	defer surface.Free()

	if tex, err = Renderer.CreateTextureFromSurface(surface); err != nil {
		return err
	}
	defer tex.Destroy()

	Renderer.Copy(tex, nil, &sdl.Rect{
		X: int32(t.X),
		Y: int32(t.Y),
		W: int32(surface.W),
		H: int32(surface.H),
	})
	return nil
}
