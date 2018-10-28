package wallpaper

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"testing"
)

func TestWallpaper(t *testing.T) {
	var testFunc = func(width, height int) {
		var (
			qWidth  = width / 2
			qHeight = height / 2
			red     = color.RGBA{255, 0, 0, 255}
			green   = color.RGBA{0, 255, 0, 255}
			blue    = color.RGBA{0, 0, 255, 255}
			pink    = color.RGBA{255, 0, 255, 255}
		)

		// Create a dummy image that is width*height and has the four
		// quadrants laid out as solid colors:
		// Red  | Green
		// Blue | Pink
		img := image.NewRGBA(image.Rect(0, 0, width, height))
		draw.Draw(
			// Corner: red
			img, // dst Image
			image.Rect(0, 0, qWidth, qHeight), // r Rectangle
			image.NewUniform(red),             // src Image
			image.Point{0, 0},                 // sp Point
			draw.Over,                         // op Op
		)
		draw.Draw(
			// Top: green
			img,
			image.Rect(qWidth, 0, width, qHeight),
			image.NewUniform(green),
			image.Point{qWidth, 0},
			draw.Over,
		)
		draw.Draw(
			// Left: blue
			img,
			image.Rect(0, qHeight, qWidth, height),
			image.NewUniform(blue),
			image.Point{0, qHeight},
			draw.Over,
		)
		draw.Draw(
			// Repeat: pink
			img,
			image.Rect(qWidth, qHeight, width, height),
			image.NewUniform(pink),
			image.Point{qWidth, qHeight},
			draw.Over,
		)

		// Output as png to disk if you wanna see what's in it.
		if os.Getenv("T_WALLPAPER_PNG") != "" {
			fn := fmt.Sprintf("test-%dx%d.png", width, height)
			if fh, err := os.Create(fn); err == nil {
				defer fh.Close()
				if err := png.Encode(fh, img); err != nil {
					t.Errorf("err: %s", err)
				}
			}
		}

		wp, err := FromImage(nil, img, "dummy")
		if err != nil {
			t.Errorf("Couldn't create FromImage: %s", err)
			t.FailNow()
		}

		// Check the quarter size is what we expected.
		w, h := wp.QuarterSize()
		if w != qWidth || h != qHeight {
			t.Errorf(
				"Got wrong quarter size: expected %dx%d but got %dx%d",
				qWidth, qHeight,
				w, h,
			)
		}

		// Test the colors.
		testColor := func(name string, img *image.RGBA, expect color.RGBA) {
			if actual := img.At(5, 5); actual != expect {
				t.Errorf(
					"%s: expected color %v but got %v",
					name,
					expect,
					actual,
				)
			}
		}
		testColor("Corner", wp.Corner(), red)
		testColor("Top", wp.Top(), green)
		testColor("Left", wp.Left(), blue)
		testColor("Repeat", wp.Repeat(), pink)
	}

	testFunc(128, 128)
	testFunc(128, 64)
	testFunc(64, 128)
	testFunc(12, 12)
	testFunc(57, 39)
}
