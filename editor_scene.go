package doodle

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"git.kirsle.net/apps/doodle/balance"
	"git.kirsle.net/apps/doodle/doodads"
	"git.kirsle.net/apps/doodle/enum"
	"git.kirsle.net/apps/doodle/events"
	"git.kirsle.net/apps/doodle/level"
	"git.kirsle.net/apps/doodle/render"
	"git.kirsle.net/apps/doodle/uix"
)

// EditorScene manages the "Edit Level" game mode.
type EditorScene struct {
	// Configuration for the scene initializer.
	DrawingType enum.DrawingType
	OpenFile    bool
	Filename    string
	DoodadSize  int

	UI *EditorUI

	// The current level or doodad object being edited, based on the
	// DrawingType.
	Level  *level.Level
	Doodad *doodads.Doodad

	// The canvas widget that contains the map we're working on.
	// XXX: in dev builds this is available at $ d.Scene.GetDrawing()
	drawing *uix.Canvas

	// Last saved filename by the user.
	filename string
}

// Name of the scene.
func (s *EditorScene) Name() string {
	return "Edit"
}

// Setup the editor scene.
func (s *EditorScene) Setup(d *Doodle) error {
	s.drawing = uix.NewCanvas(balance.ChunkSize, true)
	if len(s.drawing.Palette.Swatches) > 0 {
		s.drawing.SetSwatch(s.drawing.Palette.Swatches[0])
	}

	// TODO: move inside the UI. Just an approximate position for now.
	s.drawing.MoveTo(render.NewPoint(0, 19))
	s.drawing.Resize(render.NewRect(d.width-150, d.height-44))
	s.drawing.Compute(d.Engine)

	// Were we given configuration data?
	if s.Filename != "" {
		log.Debug("EditorScene.Setup: Set filename to %s", s.Filename)
		s.filename = s.Filename
		s.Filename = ""
	}

	// Loading a Level or a Doodad?
	switch s.DrawingType {
	case enum.LevelDrawing:
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
			s.drawing.ScrollTo(render.Origin)
			s.drawing.Scrollable = true
		}
	case enum.DoodadDrawing:
		// No Doodad?
		if s.filename != "" && s.OpenFile {
			log.Debug("EditorScene.Setup: Loading doodad from filename at %s", s.filename)
			if err := s.LoadDoodad(s.filename); err != nil {
				d.Flash("LoadDoodad error: %s", err)
			}
		}

		// No Doodad?
		if s.Doodad == nil {
			log.Debug("EditorScene.Setup: initializing a new Doodad")
			s.Doodad = doodads.New(s.DoodadSize)
			s.drawing.LoadDoodad(s.Doodad)
		}

		// TODO: move inside the UI. Just an approximate position for now.
		s.drawing.MoveTo(render.NewPoint(200, 200))
		s.drawing.Resize(render.NewRect(int32(s.DoodadSize), int32(s.DoodadSize)))
		s.drawing.ScrollTo(render.Origin)
		s.drawing.Scrollable = false
		s.drawing.Compute(d.Engine)
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
func (s *EditorScene) SaveLevel(filename string) error {
	if s.DrawingType != enum.LevelDrawing {
		return errors.New("SaveLevel: current drawing is not a Level type")
	}

	if !strings.HasSuffix(filename, extLevel) {
		filename += extLevel
	}

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
		return fmt.Errorf("SaveLevel error: %s", err)
	}

	// Save it to their profile directory.
	filename = LevelPath(filename)
	log.Info("Write Level: %s", filename)
	err = ioutil.WriteFile(filename, json, 0644)
	if err != nil {
		return fmt.Errorf("Create map file error: %s", err)
	}

	return nil
}

// LoadDoodad loads a doodad from disk.
func (s *EditorScene) LoadDoodad(filename string) error {
	s.filename = filename

	doodad, err := doodads.LoadJSON(filename)
	if err != nil {
		return fmt.Errorf("EditorScene.LoadDoodad(%s): %s", filename, err)
	}

	s.DrawingType = enum.DoodadDrawing
	s.Doodad = doodad
	s.DoodadSize = doodad.Layers[0].Chunker.Size
	s.drawing.LoadDoodad(s.Doodad)
	return nil
}

// SaveDoodad saves the doodad to disk.
func (s *EditorScene) SaveDoodad(filename string) error {
	if s.DrawingType != enum.DoodadDrawing {
		return errors.New("SaveDoodad: current drawing is not a Doodad type")
	}

	if !strings.HasSuffix(filename, extDoodad) {
		filename += extDoodad
	}

	s.filename = filename
	d := s.Doodad
	if d.Title == "" {
		d.Title = "Untitled Doodad"
	}
	if d.Author == "" {
		d.Author = os.Getenv("USER")
	}

	// TODO: is this copying necessary?
	d.Palette = s.drawing.Palette
	d.Layers[0].Chunker = s.drawing.Chunker()

	// Save it to their profile directory.
	filename = DoodadPath(filename)
	log.Info("Write Doodad: %s", filename)
	err := d.WriteJSON(filename)
	return err
}

// Destroy the scene.
func (s *EditorScene) Destroy() error {
	return nil
}
