package sdl

import (
	"fmt"
	"strings"

	"git.kirsle.net/apps/doodle/lib/events"
	"git.kirsle.net/apps/doodle/lib/render"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

// TODO: font filenames
var defaultFontFilename = "./fonts/DejaVuSans.ttf"

var fonts = map[string]*ttf.Font{}

// LoadFont loads and caches the font at a given size.
func LoadFont(filename string, size int) (*ttf.Font, error) {
	if filename == "" {
		filename = defaultFontFilename
	}

	// Cached font available?
	keyName := fmt.Sprintf("%s@%d", filename, size)
	if font, ok := fonts[keyName]; ok {
		return font, nil
	}

	font, err := ttf.OpenFont(filename, size)
	if err != nil {
		return nil, err
	}
	fonts[keyName] = font

	return font, nil
}

// Keysym returns the current key pressed, taking into account the Shift
// key modifier.
func (r *Renderer) Keysym(ev *events.State) string {
	if key := ev.KeyName.Read(); key != "" {
		if ev.ShiftActive.Pressed() {
			if symbol, ok := shiftMap[key]; ok {
				return symbol
			}
			return strings.ToUpper(key)
		}
	}
	return ""
}

// ComputeTextRect computes and returns a Rect for how large the text would
// appear if rendered.
func (r *Renderer) ComputeTextRect(text render.Text) (render.Rect, error) {
	var (
		rect    render.Rect
		font    *ttf.Font
		surface *sdl.Surface
		color   = ColorToSDL(text.Color)
		err     error
	)

	if font, err = LoadFont(text.FontFilename, text.Size); err != nil {
		return rect, err
	}

	if surface, err = font.RenderUTF8Blended(text.Text, color); err != nil {
		return rect, err
	}
	defer surface.Free()

	rect.W = surface.W
	rect.H = surface.H
	return rect, err
}

// DrawText draws text on the canvas.
func (r *Renderer) DrawText(text render.Text, point render.Point) error {
	var (
		font    *ttf.Font
		surface *sdl.Surface
		tex     *sdl.Texture
		err     error
	)

	if font, err = LoadFont(text.FontFilename, text.Size); err != nil {
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
			X: point.X + dx,
			Y: point.Y + dy,
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

// shiftMap maps keys to their Shift versions.
var shiftMap = map[string]string{
	"`": "~",
	"1": "!",
	"2": "@",
	"3": "#",
	"4": "$",
	"5": "%",
	"6": "^",
	"7": "&",
	"8": "*",
	"9": "(",
	"0": ")",
	"-": "_",
	"=": "+",
	"[": "{",
	"]": "}",
	`\`: "|",
	";": ":",
	`'`: `"`,
	",": "<",
	".": ">",
	"/": "?",
}
