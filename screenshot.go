package doodle

import (
	"fmt"
	"image"
	"image/png"
	"io/ioutil"
	"os"
	"time"

	"git.kirsle.net/apps/doodle/draw"
	"git.kirsle.net/apps/doodle/level"
)

// SaveLevel saves the level to disk.
func (d *Doodle) SaveLevel() {
	m := level.Level{
		Version: 1,
		Title:   "Alpha",
		Author:  os.Getenv("USER"),
		Width:   d.width,
		Height:  d.height,
		Palette: []level.Palette{
			level.Palette{
				Color: "#000000",
				Solid: true,
			},
		},
		Pixels: []level.Pixel{},
	}

	for pixel := range d.canvas {
		for point := range draw.Line(pixel.x, pixel.y, pixel.dx, pixel.dy) {
			m.Pixels = append(m.Pixels, level.Pixel{
				X:       point.X,
				Y:       point.Y,
				Palette: 0,
			})
		}
	}

	json, err := m.ToJSON()
	if err != nil {
		log.Error("SaveLevel error: %s", err)
		return
	}

	filename := fmt.Sprintf("./map-%s.json",
		time.Now().Format("2006-01-02T15-04-05"),
	)
	err = ioutil.WriteFile(filename, json, 0644)
	if err != nil {
		log.Error("Create map file error: %s", err)
		return
	}
}

// LoadLevel loads a map from JSON.
func (d *Doodle) LoadLevel(filename string) error {
	log.Info("Loading level from file: %s", filename)
	pixelHistory = []Pixel{}
	d.canvas = Grid{}

	m, err := level.LoadJSON(filename)
	if err != nil {
		return err
	}

	for _, point := range m.Pixels {
		pixel := Pixel{
			start: true,
			x:     point.X,
			y:     point.Y,
			dx:    point.X,
			dy:    point.Y,
		}
		pixelHistory = append(pixelHistory, pixel)
		d.canvas[pixel] = nil
	}

	return nil
}

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
			for point := range draw.Line(pixel.x, pixel.y, pixel.dx, pixel.dy) {
				screenshot.Set(int(point.X), int(point.Y), image.Black)
			}
		}
	}

	// Create the screenshot directory.
	if _, err := os.Stat("./screenshots"); os.IsNotExist(err) {
		log.Info("Creating directory: ./screenshots")
		err = os.Mkdir("./screenshots", 0755)
		if err != nil {
			log.Error("Can't create ./screenshots: %s", err)
			return
		}
	}

	filename := fmt.Sprintf("./screenshots/screenshot-%s.png",
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
