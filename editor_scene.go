package doodle

import (
	"fmt"
	"io/ioutil"
	"os"

	"git.kirsle.net/apps/doodle/balance"
	"git.kirsle.net/apps/doodle/enum"
	"git.kirsle.net/apps/doodle/events"
	"git.kirsle.net/apps/doodle/level"
	"git.kirsle.net/apps/doodle/render"
)

// EditorScene manages the "Edit Level" game mode.
type EditorScene struct {
	// Configuration for the scene initializer.
	OpenFile bool
	Filename string

	UI *EditorUI

	// The current level being edited.
	DrawingType enum.DrawingType
	Level       *level.Level

	// The canvas widget that contains the map we're working on.
	// XXX: in dev builds this is available at $ d.Scene.GetDrawing()
	drawing *level.Canvas

	// Last saved filename by the user.
	filename string
}

// Name of the scene.
func (s *EditorScene) Name() string {
	return "Edit"
}

// Setup the editor scene.
func (s *EditorScene) Setup(d *Doodle) error {
	s.drawing = level.NewCanvas(balance.ChunkSize, true)
	if len(s.drawing.Palette.Swatches) > 0 {
		s.drawing.SetSwatch(s.drawing.Palette.Swatches[0])
	}

	// TODO: move inside the UI. Just an approximate position for now.
	s.drawing.MoveTo(render.NewPoint(0, 19))
	s.drawing.Resize(render.NewRect(d.width-150, d.height-44))
	s.drawing.Compute(d.Engine)

	// // Were we given configuration data?
	if s.Filename != "" {
		log.Debug("EditorScene.Setup: Set filename to %s", s.Filename)
		s.filename = s.Filename
		s.Filename = ""
	}
	if s.Level != nil {
		log.Debug("EditorScene.Setup: received level from scene caller")
		s.drawing.LoadLevel(s.Level)
	} else if s.filename != "" && s.OpenFile {
		log.Debug("EditorScene.Setup: Loading map from filename at %s", s.filename)
		if err := s.LoadLevel(s.filename); err != nil {
			d.Flash("LoadLevel error: %s", err)
		}
	}

	// No level?
	if s.Level == nil {
		log.Debug("EditorScene.Setup: initializing a new Level")
		s.Level = level.New()
		s.Level.Palette = level.DefaultPalette()
		s.drawing.LoadLevel(s.Level)
	}

	// Initialize the user interface. It references the palette and such so it
	// must be initialized after those things.
	s.UI = NewEditorUI(d, s)
	d.Flash("Editor Mode. Press 'P' to play this map.")

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
			Filename: s.filename,
			Level:    s.Level,
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
	s.drawing.Present(d.Engine, s.drawing.Point())

	return nil
}

// LoadLevel loads a level from disk.
func (s *EditorScene) LoadLevel(filename string) error {
	s.filename = filename

	level, err := level.LoadJSON(filename)
	if err != nil {
		return fmt.Errorf("EditorScene.LoadLevel(%s): %s", filename, err)
	}

	s.DrawingType = enum.LevelDrawing
	s.Level = level
	s.drawing.LoadLevel(s.Level)
	return nil
}

// SaveLevel saves the level to disk.
// TODO: move this into the Canvas?
func (s *EditorScene) SaveLevel(filename string) {
	s.filename = filename

	m := s.Level
	if m.Title == "" {
		m.Title = "Alpha"
	}
	if m.Author == "" {
		m.Author = os.Getenv("USER")
	}

	m.Palette = s.drawing.Palette
	m.Chunker = s.drawing.Chunker()

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
