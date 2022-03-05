// +build !js

package native

import (
	"image"

	"git.kirsle.net/apps/doodle/pkg/log"
	"git.kirsle.net/go/render"
	"git.kirsle.net/go/render/sdl"
	sdl2 "github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

// Native render engine functions (SDL2 edition),
// not for JavaScript/WASM yet.

/*
TextToImage takes an SDL2_TTF texture and makes it into a Go image.

Notes:
- The text is made Black & White with a white background on the image.
- Drop shadow, stroke, etc. probably not supported.
- Returns a non-antialiased image.
*/
func TextToImage(e render.Engine, text render.Text) (image.Image, error) {
	// engine, _ := e.(*sdl.Renderer)

	// Make the text black & white for ease of identifying pixels.
	text.Color = render.Black

	var (
		// renderer = engine.GetSDL2Renderer()
		font     *ttf.Font
		surface  *sdl2.Surface
		pixFmt   *sdl2.PixelFormat
		surface2 *sdl2.Surface
		err      error
	)

	if font, err = sdl.LoadFont(text.FontFilename, text.Size); err != nil {
		return nil, err
	}

	if surface, err = font.RenderUTF8Solid(text.Text, sdl.ColorToSDL(text.Color)); err != nil {
		return nil, err
	}
	defer surface.Free()
	log.Error("surf fmt: %+v", surface.Format)

	// Convert the Surface into a pixelformat that supports the .At(x,y)
	// function properly, as the one we got above is "Not implemented"
	if pixFmt, err = sdl2.AllocFormat(sdl2.PIXELFORMAT_RGB888); err != nil {
		return nil, err
	}
	if surface2, err = surface.Convert(pixFmt, 0); err != nil {
		return nil, err
	}
	defer surface2.Free()

	// Read back the pixels.
	var (
		x   int
		y   int
		w   = int(surface2.W)
		h   = int(surface2.H)
		img = image.NewRGBA(image.Rect(x, y, w, h))
	)
	for x = 0; x < w; x++ {
		for y = 0; y < h; y++ {
			hue := surface2.At(x, y)
			img.Set(x, y, hue)
			// log.Warn("hue: %s", hue)
			// r, g, b, _ := hue.RGBA()
			// if r == 0 && g == 0 && b == 0 {
			// 	img.Set(x, y, hue)
			// } else {
			// 	img.Set(x, y, color.Transparent)
			// }
		}
	}

	return img, nil
}
