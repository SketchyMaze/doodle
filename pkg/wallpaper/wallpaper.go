package wallpaper

import (
	"image"
	"image/draw"
	"os"
	"path/filepath"
	"strings"

	"git.kirsle.net/apps/doodle/render"
)

// Wallpaper is a repeatable background image to go behind levels.
type Wallpaper struct {
	Name   string
	Format string // image file format
	Image  *image.RGBA

	// Parsed values.
	quarterWidth  int
	quarterHeight int

	// The four parsed images.
	corner *image.RGBA // Top Left corner
	top    *image.RGBA // Top repeating
	left   *image.RGBA // Left repeating
	repeat *image.RGBA // Main repeating

	// Cached textures.
	tex struct {
		corner render.Texturer
		top    render.Texturer
		left   render.Texturer
		repeat render.Texturer
	}
}

// FromImage creates a Wallpaper from an image.Image.
// If the renger.Engine is nil it will compute images but not pre-cache any
// textures yet.
func FromImage(e render.Engine, img *image.RGBA, name string) (*Wallpaper, error) {
	wp := &Wallpaper{
		Name:  name,
		Image: img,
	}
	wp.cache(e)
	return wp, nil
}

// FromFile creates a Wallpaper from a file on disk.
// If the renger.Engine is nil it will compute images but not pre-cache any
// textures yet.
func FromFile(e render.Engine, filename string) (*Wallpaper, error) {
	fh, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	img, format, err := image.Decode(fh)
	if err != nil {
		return nil, err
	}

	// Ugly hack: make it an image.RGBA because the thing we get tends to be
	// an image.Paletted, UGH!
	var b = img.Bounds()
	rgba := image.NewRGBA(b)
	for x := b.Min.X; x < b.Max.X; x++ {
		for y := b.Min.Y; y < b.Max.Y; y++ {
			rgba.Set(x, y, img.At(x, y))
		}
	}

	wp := &Wallpaper{
		Name:   strings.Split(filepath.Base(filename), ".")[0],
		Format: format,
		Image:  rgba,
	}
	wp.cache(e)
	return wp, nil
}

// cache the bitmap images.
func (wp *Wallpaper) cache(e render.Engine) {
	// Zero-bound the rect cuz an image.Rect doesn't necessarily contain 0,0
	var rect = wp.Image.Bounds()
	if rect.Min.X < 0 {
		rect.Max.X += rect.Min.X
		rect.Min.X = 0
	}
	if rect.Min.Y < 0 {
		rect.Max.Y += rect.Min.Y
		rect.Min.Y = 0
	}

	// Our quarter rect size.
	wp.quarterWidth = int(float64((rect.Max.X - rect.Min.X) / 2))
	wp.quarterHeight = int(float64((rect.Max.Y - rect.Min.Y) / 2))
	quarter := image.Rect(0, 0, wp.quarterWidth, wp.quarterHeight)

	// Slice the image into the four corners.
	slice := func(dx, dy int) *image.RGBA {
		slice := image.NewRGBA(quarter)
		draw.Draw(
			slice,
			image.Rect(0, 0, wp.quarterWidth, wp.quarterHeight),
			wp.Image,
			image.Point{dx, dy},
			draw.Over,
		)
		return slice
	}
	wp.corner = slice(0, 0)
	wp.top = slice(wp.quarterWidth, 0)
	wp.left = slice(0, wp.quarterHeight)
	wp.repeat = slice(wp.quarterWidth, wp.quarterHeight)

}

// QuarterSize returns the width and height of the quarter images.
func (wp *Wallpaper) QuarterSize() (int, int) {
	return wp.quarterWidth, wp.quarterHeight
}

// Corner returns the top left corner image.
func (wp *Wallpaper) Corner() *image.RGBA {
	return wp.corner
}

// Top returns the top repeating image.
func (wp *Wallpaper) Top() *image.RGBA {
	return wp.top
}

// Left returns the left repeating image.
func (wp *Wallpaper) Left() *image.RGBA {
	return wp.left
}

// Repeat returns the main repeating image.
func (wp *Wallpaper) Repeat() *image.RGBA {
	return wp.repeat
}
