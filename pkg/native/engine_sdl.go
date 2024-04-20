//go:build !js
// +build !js

package native

import (
	"errors"
	"fmt"
	"image"

	"git.kirsle.net/SketchyMaze/doodle/pkg/log"
	"git.kirsle.net/SketchyMaze/doodle/pkg/shmem"
	"git.kirsle.net/go/render"
	"git.kirsle.net/go/render/sdl"
	sdl2 "github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

// Native render engine functions (SDL2 edition),
// not for JavaScript/WASM yet.

// HasTouchscreen checks if the device has at least one SDL_GetNumTouchDevices.
//
// Note: SDL2 GetNumTouchDevices will sometimes return 0 until a touch device is actually touched. On Macbooks,
// the trackpad counts as a touch device and on first touch, HasTouchscreen may begin returning true.
func HasTouchscreen(e render.Engine) bool {
	if _, ok := e.(*sdl.Renderer); ok {
		return sdl2.GetNumTouchDevices() > 0
	}
	return false
}

// CopyToClipboard puts some text on your clipboard.
func CopyToClipboard(text string) error {
	if _, ok := shmem.CurrentRenderEngine.(*sdl.Renderer); ok {
		return sdl2.SetClipboardText(text)
	}
	return errors.New("not supported")
}

// CountTextures returns the count of loaded SDL2 textures, for the F3 debug overlay, or "n/a"
func CountTextures(e render.Engine) string {
	var texCount = "n/a"
	if sdl, ok := e.(*sdl.Renderer); ok {
		texCount = fmt.Sprintf("%d", sdl.CountTextures())
	}
	return texCount
}

// FreeTextures will free all SDL2 textures currently in memory.
func FreeTextures(e render.Engine) {
	if sdl, ok := e.(*sdl.Renderer); ok {
		texCount := sdl.FreeTextures()
		if texCount > 0 {
			log.Info("FreeTextures: %d SDL2 textures freed", texCount)
		}
	}
}

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

// Set the window to maximized.
func MaximizeWindow(e render.Engine) {
	if sdl, ok := e.(*sdl.Renderer); ok {
		sdl.Maximize()
	}
}
