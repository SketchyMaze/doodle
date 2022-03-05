package drawtool

import (
	"git.kirsle.net/apps/doodle/pkg/native"
	"git.kirsle.net/go/render"
	"git.kirsle.net/go/ui"
)

// TextSettings holds currently selected Text Tool settings.
type TextSettings struct {
	Font    string // like 'DejaVuSans.ttf'
	Size    int
	Message string
	Label   *ui.Label // cached label texture
}

// Currently active settings (global variable)
var TT TextSettings

// IsZero checks if the TextSettings are populated.
func (tt TextSettings) IsZero() bool {
	return tt.Font == "" && tt.Size == 0 && tt.Message == ""
}

// ToStroke converts a TextSettings configuration into a Freehand
// Stroke, coloring in all of the pixels.
func (tt TextSettings) ToStroke(e render.Engine, color render.Color, at render.Point) (*Stroke, error) {
	stroke := NewStroke(Freehand, color)

	// Render the text to a Go image so we can get the colors from
	// it uniformly.
	img, err := native.TextToImage(e, tt.Label.Font)
	if err != nil {
		return nil, err
	}

	// Pull all its pixels.
	var (
		max = img.Bounds().Max
		x   = 0
		y   = 0
	)
	for x = 0; x < max.X; x++ {
		for y = 0; y < max.Y; y++ {
			hue := img.At(x, y)
			r, g, b, _ := hue.RGBA()
			if r == 65535 && g == r && b == r {
				continue
			}

			stroke.Points = append(stroke.Points, render.NewPoint(x+at.X, y+at.Y))
		}
	}

	return stroke, nil
}
