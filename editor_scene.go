package doodle

import (
	"io/ioutil"
	"os"

	"git.kirsle.net/apps/doodle/events"
	"git.kirsle.net/apps/doodle/level"
	"git.kirsle.net/apps/doodle/render"
)

// EditorScene manages the "Edit Level" game mode.
type EditorScene struct {
	// Configuration for the scene initializer.
	OpenFile bool
	Filename string
	Canvas   *level.Grid

	UI *EditorUI

	// The canvas widget that contains the map we're working on.
	// XXX: in dev builds this is available at $ d.Scene.GetDrawing()
	drawing *level.Canvas

	// History of all the pixels placed by the user.
	filename string // Last saved filename.

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
	s.drawing = level.NewCanvas(true)
	s.drawing.Palette = level.DefaultPalette()
	if len(s.drawing.Palette.Swatches) > 0 {
		s.drawing.SetSwatch(s.drawing.Palette.Swatches[0])
	}

	// Were we given configuration data?
	if s.Filename != "" {
		log.Debug("EditorScene: Set filename to %s", s.Filename)
		s.filename = s.Filename
		s.Filename = ""
		if s.OpenFile {
			log.Debug("EditorScene: Loading map from filename at %s", s.filename)
			if err := s.LoadLevel(s.filename); err != nil {
				d.Flash("LoadLevel error: %s", err)
			}
		}
	}
	if s.Canvas != nil {
		log.Debug("EditorScene: Received Canvas from caller")
		s.drawing.Load(s.drawing.Palette, s.Canvas)
		s.Canvas = nil
	}

	// Initialize the user interface. It references the palette and such so it
	// must be initialized after those things.
	s.UI = NewEditorUI(d, s)
	d.Flash("Editor Mode. Press 'P' to play this map.")

	s.width = d.width // TODO: canvas width = copy the window size
	s.height = d.height
	return nil
}

// Loop the editor scene.
func (s *EditorScene) Loop(d *Doodle, ev *events.State) error {
	s.UI.Loop(ev)
	s.drawing.Loop(ev)

	// Switching to Play Mode?
	if ev.KeyName.Read() == "p" {
		log.Info("Play Mode, Go!")
		d.Goto(&PlayScene{
			Canvas: s.drawing.Grid(),
		})
		return nil
	}

	return nil
}

// Draw the current frame.
func (s *EditorScene) Draw(d *Doodle) error {
	// Clear the canvas and fill it with magenta so it's clear if any spots are missed.
	d.Engine.Clear(render.Magenta)

	s.UI.Present(d.Engine)

	// TODO: move inside the UI. Just an approximate position for now.
	s.drawing.MoveTo(render.NewPoint(0, 19))
	s.drawing.Resize(render.NewRect(d.width-150, d.height-44))
	s.drawing.Compute(d.Engine)
	s.drawing.Present(d.Engine, s.drawing.Point())

	return nil
}

// LoadLevel loads a level from disk.
func (s *EditorScene) LoadLevel(filename string) error {
	s.filename = filename
	return s.drawing.LoadFilename(filename)

}

// SaveLevel saves the level to disk.
// TODO: move this into the Canvas?
func (s *EditorScene) SaveLevel(filename string) {
	s.filename = filename

	m := level.New()
	m.Title = "Alpha"
	m.Author = os.Getenv("USER")
	m.Width = s.width
	m.Height = s.height
	m.Palette = s.drawing.Palette

	for pixel := range *s.drawing.Grid() {
		m.Pixels = append(m.Pixels, &level.Pixel{
			X:            pixel.X,
			Y:            pixel.Y,
			PaletteIndex: int32(pixel.Swatch.Index()),
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

// Destroy the scene.
func (s *EditorScene) Destroy() error {
	return nil
}
