package doodle

import (
	"fmt"
	"image"
	"image/png"
	"io/ioutil"
	"os"
	"time"

	"git.kirsle.net/apps/doodle/events"
	"git.kirsle.net/apps/doodle/level"
	"git.kirsle.net/apps/doodle/render"
)

// EditorScene manages the "Edit Level" game mode.
type EditorScene struct {
	// Configuration for the scene initializer.
	OpenFile bool
	Filename string
	Canvas   render.Grid

	// History of all the pixels placed by the user.
	pixelHistory []level.Pixel
	lastPixel    *level.Pixel // last pixel placed while mouse down and dragging
	canvas       render.Grid
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
	// Were we given configuration data?
	if s.Filename != "" {
		log.Debug("EditorScene: Set filename to %s", s.Filename)
		s.filename = s.Filename
		s.Filename = ""
		if s.OpenFile {
			log.Debug("EditorScene: Loading map from filename at %s", s.filename)
			if err := s.LoadLevel(s.filename); err != nil {
				return err
			}
		}
	}
	if s.Canvas != nil {
		log.Debug("EditorScene: Received Canvas from caller")
		s.canvas = s.Canvas
		s.Canvas = nil
	}

	d.Flash("Editor Mode. Press 'P' to play this map.")

	if s.pixelHistory == nil {
		s.pixelHistory = []level.Pixel{}
	}
	if s.canvas == nil {
		log.Debug("EditorScene: Setting default canvas to an empty grid")
		s.canvas = render.Grid{}
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

	// Switching to Play Mode?
	if ev.KeyName.Read() == "p" {
		log.Info("Play Mode, Go!")
		d.Goto(&PlayScene{
			Canvas: s.canvas,
		})
		return nil
	}

	// Clear the canvas and fill it with white.
	d.Engine.Clear(render.White)

	// Clicking? Log all the pixels while doing so.
	if ev.Button1.Now {
		// log.Warn("Button1: %+v", ev.Button1)
		lastPixel := s.lastPixel
		pixel := level.Pixel{
			X: ev.CursorX.Now,
			Y: ev.CursorY.Now,
		}

		// Append unique new pixels.
		if len(s.pixelHistory) == 0 || s.pixelHistory[len(s.pixelHistory)-1] != pixel {
			if lastPixel != nil {
				// Draw the pixels in between.
				if *lastPixel != pixel {
					for point := range render.IterLine(lastPixel.X, lastPixel.Y, pixel.X, pixel.Y) {
						dot := level.Pixel{
							X:       point.X,
							Y:       point.Y,
							Palette: lastPixel.Palette,
						}
						s.canvas[dot] = nil
					}
				}
			}

			s.lastPixel = &pixel
			s.pixelHistory = append(s.pixelHistory, pixel)

			// Save in the pixel canvas map.
			s.canvas[pixel] = nil
		}
	} else {
		s.lastPixel = nil
	}

	return nil
}

// Draw the current frame.
func (s *EditorScene) Draw(d *Doodle) error {
	s.canvas.Draw(d.Engine)

	return nil
}

// LoadLevel loads a level from disk.
func (s *EditorScene) LoadLevel(filename string) error {
	s.filename = filename
	s.pixelHistory = []level.Pixel{}
	s.canvas = render.Grid{}

	m, err := level.LoadJSON(filename)
	if err != nil {
		return err
	}

	for _, point := range m.Pixels {
		pixel := level.Pixel{
			X: point.X,
			Y: point.Y,
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
		m.Pixels = append(m.Pixels, level.Pixel{
			X:       pixel.X,
			Y:       pixel.Y,
			Palette: 0,
		})
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
		screenshot.Set(int(pixel.X), int(pixel.Y), image.Black)
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

// Destroy the scene.
func (s *EditorScene) Destroy() error {
	return nil
}
