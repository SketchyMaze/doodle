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

// EditorScene manages the "Edit Level" game mode.
type EditorScene struct {
	// History of all the pixels placed by the user.
	pixelHistory []Pixel
	canvas       Grid

	// Canvas size
	width  int32
	height int32
}

// Setup the editor scene.
func (e *EditorScene) Setup(d *Doodle) error {
	if e.pixelHistory == nil {
		e.pixelHistory = []Pixel{}
	}
	if e.canvas == nil {
		e.canvas = Grid{}
	}
	e.width = d.width // TODO: canvas width = copy the window size
	e.height = d.height
	return nil
}

// Loop the editor scene.
func (e *EditorScene) Loop(d *Doodle) error {
	ev := d.events

	// Taking a screenshot?
	if ev.ScreenshotKey.Pressed() {
		log.Info("Taking a screenshot")
		e.Screenshot()
		e.SaveLevel()
	}

	// Clear the canvas and fill it with white.
	d.renderer.SetDrawColor(255, 255, 255, 255)
	d.renderer.Clear()

	// Clicking? Log all the pixels while doing so.
	if ev.Button1.Now {
		pixel := Pixel{
			start: ev.Button1.Pressed(),
			x:     ev.CursorX.Now,
			y:     ev.CursorY.Now,
			dx:    ev.CursorX.Now,
			dy:    ev.CursorY.Now,
		}

		// Append unique new pixels.
		if len(e.pixelHistory) == 0 || e.pixelHistory[len(e.pixelHistory)-1] != pixel {
			// If not a start pixel, make the delta coord the previous one.
			if !pixel.start && len(e.pixelHistory) > 0 {
				prev := e.pixelHistory[len(e.pixelHistory)-1]
				pixel.dx = prev.x
				pixel.dy = prev.y
			}

			e.pixelHistory = append(e.pixelHistory, pixel)

			// Save in the pixel canvas map.
			e.canvas[pixel] = nil
		}
	}

	d.renderer.SetDrawColor(0, 0, 0, 255)
	for i, pixel := range e.pixelHistory {
		if !pixel.start && i > 0 {
			prev := e.pixelHistory[i-1]
			if prev.x == pixel.x && prev.y == pixel.y {
				d.renderer.DrawPoint(pixel.x, pixel.y)
			} else {
				d.renderer.DrawLine(
					pixel.x,
					pixel.y,
					prev.x,
					prev.y,
				)
			}
		}
		d.renderer.DrawPoint(pixel.x, pixel.y)
	}

	// Draw the FPS.
	d.DrawDebugOverlay()

	d.renderer.Present()

	return nil
}

// LoadLevel loads a level from disk.
func (e *EditorScene) LoadLevel(filename string) error {
	e.pixelHistory = []Pixel{}
	e.canvas = Grid{}

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
		e.pixelHistory = append(e.pixelHistory, pixel)
		e.canvas[pixel] = nil
	}

	return nil
}

// SaveLevel saves the level to disk.
func (e *EditorScene) SaveLevel() {
	m := level.Level{
		Version: 1,
		Title:   "Alpha",
		Author:  os.Getenv("USER"),
		Width:   e.width,
		Height:  e.height,
		Palette: []level.Palette{
			level.Palette{
				Color: "#000000",
				Solid: true,
			},
		},
		Pixels: []level.Pixel{},
	}

	for pixel := range e.canvas {
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

// Screenshot saves the level canvas to disk as a PNG image.
func (e *EditorScene) Screenshot() {
	screenshot := image.NewRGBA(image.Rect(0, 0, int(e.width), int(e.height)))

	// White-out the image.
	for x := 0; x < int(e.width); x++ {
		for y := 0; y < int(e.height); y++ {
			screenshot.Set(x, y, image.White)
		}
	}

	// Fill in the dots we drew.
	for pixel := range e.canvas {
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
