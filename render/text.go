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
