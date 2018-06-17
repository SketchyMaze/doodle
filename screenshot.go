package doodle

import (
	"fmt"
	"image"
	"image/png"
	"math"
	"os"
	"time"
)

// Screenshot saves the level canvas to disk as a PNG image.
func (d *Doodle) Screenshot() {
	screenshot := image.NewRGBA(image.Rect(0, 0, int(d.width), int(d.height)))

	// White-out the image.
	for x := 0; x < int(d.width); x++ {
		for y := 0; y < int(d.height); y++ {
			screenshot.Set(x, y, image.White)
		}
	}

	// Fill in the dots we drew.
	for pixel := range d.canvas {
		// A line or a dot?
		if pixel.x == pixel.dx && pixel.y == pixel.dy {
			screenshot.Set(int(pixel.x), int(pixel.y), image.Black)
		} else {
			// Draw a line. TODO: get this into its own function!
			// https://en.wikipedia.org/wiki/Digital_differential_analyzer_(graphics_algorithm)
			var (
				x1 = pixel.x
				x2 = pixel.dx
				y1 = pixel.y
				y2 = pixel.dy
			)
			var (
				dx = float64(x2 - x1)
				dy = float64(y2 - y1)
			)
			var step float64
			if math.Abs(dx) >= math.Abs(dy) {
				step = math.Abs(dx)
			} else {
				step = math.Abs(dy)
			}

			dx = dx / step
			dy = dy / step
			x := float64(x1)
			y := float64(y1)
			for i := 0; i <= int(step); i++ {
				screenshot.Set(int(x), int(y), image.Black)
				x += dx
				y += dy
			}
		}

	}

	filename := fmt.Sprintf("screenshot-%s.png",
		time.Now().Format("2006-01-02T15-04-05"),
	)
	fh, err := os.Create(filename)
	if err != nil {
		log.Error(err.Error())
		return
	}
	defer fh.Close()

	if err := png.Encode(fh, screenshot); err != nil {
		log.Error(err.Error())
		return
	}
}
