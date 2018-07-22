package doodle

import (
	"fmt"
	"image"
	"image/png"
	"io/ioutil"
	"os"
	"time"

	"git.kirsle.net/apps/doodle/draw"
	"git.kirsle.net/apps/doodle/events"
	"git.kirsle.net/apps/doodle/level"
	"git.kirsle.net/apps/doodle/render"
)

// EditorScene manages the "Edit Level" game mode.
type EditorScene struct {
	// History of all the pixels placed by the user.
	pixelHistory []Pixel
	canvas       Grid
	filename     string // Last saved filename.

	// Canvas size
	width  int32
	height int32
}

// Name of the scene.
func (s *EditorScene) Name() string {
	return "Edit"
}

// Setup the editor scene.
func (s *EditorScene) Setup(d *Doodle) error {
	if s.pixelHistory == nil {
		s.pixelHistory = []Pixel{}
	}
	if s.canvas == nil {
		s.canvas = Grid{}
	}
	s.width = d.width // TODO: canvas width = copy the window size
	s.height = d.height
	return nil
}

// Loop the editor scene.
func (s *EditorScene) Loop(d *Doodle, ev *events.State) error {
	// Taking a screenshot?
	if ev.ScreenshotKey.Pressed() {
		log.Info("Taking a screenshot")
		s.Screenshot()
	}

	// Clear the canvas and fill it with white.
	d.Engine.Clear(render.White)

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
		if len(s.pixelHistory) == 0 || s.pixelHistory[len(s.pixelHistory)-1] != pixel {
			// If not a start pixel, make the delta coord the previous one.
			if !pixel.start && len(s.pixelHistory) > 0 {
				prev := s.pixelHistory[len(s.pixelHistory)-1]
				pixel.dx = prev.x
				pixel.dy = prev.y
			}

			s.pixelHistory = append(s.pixelHistory, pixel)

			// Save in the pixel canvas map.
			s.canvas[pixel] = nil
		}
	}

	return nil
}

// Draw the current frame.
func (s *EditorScene) Draw(d *Doodle) error {
	for i, pixel := range s.pixelHistory {
		if !pixel.start && i > 0 {
			prev := s.pixelHistory[i-1]
			if prev.x == pixel.x && prev.y == pixel.y {
				d.Engine.DrawPoint(
					render.Black,
					render.Point{pixel.x, pixel.y},
				)
			} else {
				d.Engine.DrawLine(
					render.Black,
					render.Point{pixel.x, pixel.y},
					render.Point{prev.x, prev.y},
				)
			}
		}
		d.Engine.DrawPoint(render.Black, render.Point{pixel.x, pixel.y})
	}

	return nil
}

// LoadLevel loads a level from disk.
func (s *EditorScene) LoadLevel(filename string) error {
	s.filename = filename
	s.pixelHistory = []Pixel{}
	s.canvas = Grid{}

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
		s.pixelHistory = append(s.pixelHistory, pixel)
		s.canvas[pixel] = nil
	}

	return nil
}

// SaveLevel saves the level to disk.
func (s *EditorScene) SaveLevel(filename string) {
	s.filename = filename
	m := level.Level{
		Version: 1,
		Title:   "Alpha",
		Author:  os.Getenv("USER"),
		Width:   s.width,
		Height:  s.height,
		Palette: []level.Palette{
			level.Palette{
				Color: "#000000",
				Solid: true,
			},
		},
		Pixels: []level.Pixel{},
	}

	for pixel := range s.canvas {
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

	err = ioutil.WriteFile(filename, json, 0644)
	if err != nil {
		log.Error("Create map file error: %s", err)
		return
	}
}

// Screenshot saves the level canvas to disk as a PNG image.
func (s *EditorScene) Screenshot() {
	screenshot := image.NewRGBA(image.Rect(0, 0, int(s.width), int(s.height)))

	// White-out the image.
	for x := 0; x < int(s.width); x++ {
		for y := 0; y < int(s.height); y++ {
			screenshot.Set(x, y, image.White)
		}
	}

	// Fill in the dots we drew.
	for pixel := range s.canvas {
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
