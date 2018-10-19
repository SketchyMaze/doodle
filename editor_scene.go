package doodle

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"git.kirsle.net/apps/doodle/doodads"
	"git.kirsle.net/apps/doodle/enum"
	"git.kirsle.net/apps/doodle/events"
	"git.kirsle.net/apps/doodle/level"
	"git.kirsle.net/apps/doodle/pkg/userdir"
	"git.kirsle.net/apps/doodle/render"
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

	// Last saved filename by the user.
	filename string
}

// Name of the scene.
func (s *EditorScene) Name() string {
	return "Edit"
}

// Setup the editor scene.
func (s *EditorScene) Setup(d *Doodle) error {
	// Initialize the user interface. It references the palette and such so it
	// must be initialized after those things.
	s.UI = NewEditorUI(d, s)

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
			s.UI.Canvas.LoadLevel(s.Level)
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
			s.UI.Canvas.LoadLevel(s.Level)
			s.UI.Canvas.ScrollTo(render.Origin)
			s.UI.Canvas.Scrollable = true
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
			s.UI.Canvas.LoadDoodad(s.Doodad)
		}

		// TODO: move inside the UI. Just an approximate position for now.
		s.UI.Canvas.Resize(render.NewRect(int32(s.DoodadSize), int32(s.DoodadSize)))
		s.UI.Canvas.ScrollTo(render.Origin)
		s.UI.Canvas.Scrollable = false
		s.UI.Workspace.Compute(d.Engine)
	}

	d.Flash("Editor Mode. Press 'P' to play this map.")

	return nil
}

// Loop the editor scene.
func (s *EditorScene) Loop(d *Doodle, ev *events.State) error {
	// Has the window been resized?
	if resized := ev.Resized.Read(); resized {
		w, h := d.Engine.WindowSize()
		if w != d.width || h != d.height {
			// Not a false alarm.
			d.width = w
			d.height = h
			s.UI.Resized(d)
			return nil
		}
	}

	s.UI.Loop(ev)

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

	return nil
}

// LoadLevel loads a level from disk.
func (s *EditorScene) LoadLevel(filename string) error {
	s.filename = filename

	level, err := level.LoadJSON(filename)
	fmt.Printf("%+v\n", level)
	if err != nil {
		return fmt.Errorf("EditorScene.LoadLevel(%s): %s", filename, err)
	}

	s.DrawingType = enum.LevelDrawing
	s.Level = level
	s.UI.Canvas.LoadLevel(s.Level)

	// TODO: debug
	for i, actor := range level.Actors {
		log.Info("Actor %s is a %s", i, actor.ID())
	}
	for name, file := range level.Files {
		log.Info("File %s has: %s", name, file.Data)
	}

	log.Info("Installing %d actors into the drawing", len(level.Actors))
	if err := s.UI.Canvas.InstallActors(level.Actors); err != nil {
		return fmt.Errorf("EditorScene.LoadLevel: InstallActors: %s", err)
	}

	return nil
}

// SaveLevel saves the level to disk.
// TODO: move this into the Canvas?
func (s *EditorScene) SaveLevel(filename string) error {
	if s.DrawingType != enum.LevelDrawing {
		return errors.New("SaveLevel: current drawing is not a Level type")
	}

	if !strings.HasSuffix(filename, enum.LevelExt) {
		filename += enum.LevelExt
	}

	s.filename = filename

	m := s.Level
	if m.Title == "" {
		m.Title = "Alpha"
	}
	if m.Author == "" {
		m.Author = os.Getenv("USER")
	}

	m.Palette = s.UI.Canvas.Palette
	m.Chunker = s.UI.Canvas.Chunker()

	json, err := m.ToJSON()
	if err != nil {
		return fmt.Errorf("SaveLevel error: %s", err)
	}

	// Save it to their profile directory.
	filename = userdir.LevelPath(filename)
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
	s.UI.Canvas.LoadDoodad(s.Doodad)
	return nil
}

// SaveDoodad saves the doodad to disk.
func (s *EditorScene) SaveDoodad(filename string) error {
	if s.DrawingType != enum.DoodadDrawing {
		return errors.New("SaveDoodad: current drawing is not a Doodad type")
	}

	if !strings.HasSuffix(filename, enum.DoodadExt) {
		filename += enum.DoodadExt
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
	d.Palette = s.UI.Canvas.Palette
	d.Layers[0].Chunker = s.UI.Canvas.Chunker()

	// Save it to their profile directory.
	filename = userdir.DoodadPath(filename)
	log.Info("Write Doodad: %s", filename)
	err := d.WriteJSON(filename)
	return err
}

// Destroy the scene.
func (s *EditorScene) Destroy() error {
	debugWorldIndex = render.Origin
	return nil
}
