// Package pattern applies a kind of brush texture to a palette swatch.
package pattern

import (
	"errors"
	"fmt"

	"git.kirsle.net/apps/doodle/pkg/log"
	"git.kirsle.net/apps/doodle/pkg/sprites"
	"git.kirsle.net/go/render"
	"git.kirsle.net/go/ui"
)

// Pattern applies a texture to a color in level drawings.
type Pattern struct {
	Name     string
	Filename string
	Hidden   bool // boolProp showHiddenDoodads true
}

// Builtins are the list of the game's built-in patterns.
var Builtins = []Pattern{
	{
		Name:     "No pattern",
		Filename: "",
	},
	{
		Name:     "Noise",
		Filename: "noise.png",
	},
	{
		Name:     "Marker",
		Filename: "marker.png",
	},
	{
		Name:     "Ink",
		Filename: "ink.png",
	},
	{
		Name:     "Perlin Noise",
		Filename: "perlin-noise.png",
	},
	{
		Name:     "Bubbles",
		Filename: "bubbles.png",
	},
	{
		Name:     "Circles",
		Filename: "circles.png",
	},
	{
		Name:     "Grid",
		Filename: "grid.png",
	},
	{
		Name:     "Bars (debug)",
		Filename: "bars.png",
		Hidden:   true,
	},
}

// Images is a map of file names to ui.Image widgets,
// after LoadBuiltins had been called.
var images map[string]*ui.Image

// LoadBuiltins loads all of the PNG textures of built-in patterns
// into ui.Image widgets.
func LoadBuiltins(e render.Engine) {
	images = map[string]*ui.Image{}

	for _, pat := range Builtins {
		if pat.Filename == "" {
			continue
		}

		img, err := sprites.LoadImage(e, "assets/pattern/"+pat.Filename)
		if err != nil {
			log.Error("Load pattern %s: %s", pat.Filename, err)
		}
		images[pat.Filename] = img
	}
}

// GetImage returns the ui.Image for a builtin pattern.
func GetImage(filename string) (*ui.Image, error) {
	if images == nil {
		return nil, errors.New("pattern.GetImage: LoadBuiltins() was not called")
	}

	if im, ok := images[filename]; ok {
		return im, nil
	}
	return nil, fmt.Errorf("pattern.GetImage: filename %s not found", filename)
}

// SampleColor samples a color with the pattern for a given coordinate in infinite space.
func SampleColor(filename string, color render.Color, point render.Point) render.Color {
	if filename == "" {
		return color
	}

	// Not loaded in memory?
	if _, ok := images[filename]; !ok {
		return color
	}

	// Translate the world coord (point) into the bounds of the texture image.
	var (
		image  = images[filename].Image // the Go image.Image
		bounds = image.Bounds()
		coord  = render.Point{
			// The world coordinate bounded to the pattern image size.
			X: render.AbsInt(point.X % bounds.Max.X),
			Y: render.AbsInt(point.Y % bounds.Max.Y),
		}

		// Sample the color from the pattern texture.
		colorAt = render.FromColor(image.At(coord.X, coord.Y))

		// Average the RGBA color out to a grayscale brightness.
		// sourceAvgGray  = (int(color.Red) + int(color.Blue) + int(color.Green)/3) % 255
		// patternAvgGray = (int(colorAt.Red) + int(colorAt.Blue) + int(colorAt.Green)/3) % 255
	)

	// See if the gray average is brighter or lower than the color.
	// if sourceAvgGray < patternAvgGray {
	// 	delta := patternAvgGray - sourceAvgGray
	// 	color = color.Lighten(delta)
	// } else if sourceAvgGray > patternAvgGray {
	// 	color = color.Darken(sourceAvgGray - patternAvgGray)
	// }

	// return OverlayFilter(color, colorAt)
	// return ScreenFilter(color, colorAt)
	return GrayToColor(color, colorAt)

	// log.Info("color: %s at point: %s image point: %s", color, point, coord)

	// return color
}

// GrayToColor samples a colorful swatch with the grayscale pattern img.
func GrayToColor(color, grayscale render.Color) render.Color {
	// The grayscale image ranges from 0 to 255.
	// The color might be #FF0000 (red)
	// 127 in grayscale should be FF0000 (perfectly red)
	// 0 (black) in grayscale should be black in output
	// 255 (white) in grayscale should be white in output
	var (
		AR = float64(color.Red)
		AG = float64(color.Green)
		AB = float64(color.Blue)
		BR = float64(grayscale.Red)
		BG = float64(grayscale.Green)
		BB = float64(grayscale.Blue)
	)

	// If the pattern has a fully transparent pixel here, return transparent.
	if grayscale.Alpha == 0 {
		return render.RGBA(1, 0, 0, 1)
	}

	convert := func(cc, gs float64) uint8 {
		var delta float64
		if gs < 127 {
			// return uint8(cc + cc/gs)
			delta = cc * (gs / 255)
		} else {
			delta = cc * (gs / 255)
		}
		return uint8(delta)
	}

	return render.RGBA(
		convert(AR, BR),
		convert(AG, BG),
		convert(AB, BB),
		255,
	)
}

// ScreenFilter applies a "screen" blend mode between the two colors (a > b).
func ScreenFilter(a, b render.Color) render.Color {
	// The algorithm we're going for is:
	// 1 - (1 - a) * (1 - b)
	var (
		AR = a.Red
		AG = a.Green
		AB = a.Blue
		BR = b.Red
		BG = b.Green
		BB = b.Blue

		deltaR = 255 - (255-AR)*(255-BR)
		deltaG = 255 - (255-AG)*(255-BG)
		deltaB = 255 - (255-AB)*(255-BB)
	)

	// If the pattern has a fully transparent pixel here, return transparent.
	// if b.Alpha == 0 {
	// 	return render.RGBA(1, 0, 0, 1)
	// }

	return render.RGBA(deltaR, deltaG, deltaB, a.Alpha)
}

// OverlayFilter applies an "overlay" blend mode between the two colors.
func OverlayFilter(a, b render.Color) render.Color {
	// The algorithm we're going for is:
	// If a < 0.5: 2ab
	//  Otherwise: 1 - 2(1 - a)(1 - b)
	munch := func(a, b uint8) uint8 {
		if a < 127 {
			return 2 * a * b
		}
		return 255 - (2 * (255 - a) * (255 - b))
	}

	// If the pattern has a fully transparent pixel here, return transparent.
	if b.Alpha == 0 {
		return render.RGBA(1, 0, 0, 0)
	}

	var (
		AR = a.Red
		AG = a.Green
		AB = a.Blue
		BR = b.Red
		BG = b.Green
		BB = b.Blue

		deltaR = munch(AR, BR)
		deltaG = munch(AG, BG)
		deltaB = munch(AB, BB)
	)

	return render.RGBA(deltaR, deltaG, deltaB, a.Alpha)
}
