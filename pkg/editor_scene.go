package doodle

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"git.kirsle.net/apps/doodle/lib/events"
	"git.kirsle.net/apps/doodle/lib/render"
	"git.kirsle.net/apps/doodle/pkg/balance"
	"git.kirsle.net/apps/doodle/pkg/doodads"
	"git.kirsle.net/apps/doodle/pkg/drawtool"
	"git.kirsle.net/apps/doodle/pkg/enum"
	"git.kirsle.net/apps/doodle/pkg/level"
	"git.kirsle.net/apps/doodle/pkg/log"
	"git.kirsle.net/apps/doodle/pkg/userdir"
)

// EditorScene manages the "Edit Level" game mode.
type EditorScene struct {
	// Configuration for the scene initializer.
	DrawingType enum.DrawingType
	OpenFile    bool
	Filename    string
	DoodadSize  int

	UI *EditorUI
	d  *Doodle

	// The current level or doodad object being edited, based on the
	// DrawingType.
	Level  *level.Level
	Doodad *doodads.Doodad

	// Custom debug overlay values.
	debTool       *string
	debSwatch     *string
	debWorldIndex *string

	// Last saved filename by the user.
	filename string
}

// Name of the scene.
func (s *EditorScene) Name() string {
	return "Edit"
}

// Setup the editor scene.
func (s *EditorScene) Setup(d *Doodle) error {
	// Debug overlay values.
	s.debTool = new(string)
	s.debSwatch = new(string)
	s.debWorldIndex = new(string)
	customDebugLabels = []debugLabel{
		{"Pixel:", s.debWorldIndex},
		{"Tool:", s.debTool},
		{"Swatch:", s.debSwatch},
	}

	// Initialize the user interface. It references the palette and such so it
	// must be initialized after those things.
	s.d = d
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
			s.UI.Canvas.LoadLevel(d.Engine, s.Level)
			s.UI.Canvas.InstallActors(s.Level.Actors)
		} else if s.filename != "" && s.OpenFile {
			log.Debug("EditorScene.Setup: Loading map from filename at %s", s.filename)
			if err := s.LoadLevel(s.filename); err != nil {
				d.Flash("LoadLevel error: %s", err)
			} else {
				s.UI.Canvas.InstallActors(s.Level.Actors)
			}
		}

		// Write locked level?
		if s.Level != nil && s.Level.Locked {
			if balance.WriteLockOverride {
				d.Flash("Note: write lock has been overridden")
			} else {
				d.Flash("That level is write-protected and cannot be viewed in the editor.")
				s.Level = nil
				s.UI.Canvas.ClearActors()
				s.filename = ""
			}
		}

		// No level?
		if s.Level == nil {
			log.Debug("EditorScene.Setup: initializing a new Level")
			s.Level = level.New()
			s.Level.Palette = level.DefaultPalette()
			s.UI.Canvas.LoadLevel(d.Engine, s.Level)
			s.UI.Canvas.ScrollTo(render.Origin)
			s.UI.Canvas.Scrollable = true
		}
	case enum.DoodadDrawing:
		// Getting a doodad from file?
		if s.filename != "" && s.OpenFile {
			log.Debug("EditorScene.Setup: Loading doodad from filename at %s", s.filename)
			if err := s.LoadDoodad(s.filename); err != nil {
				d.Flash("LoadDoodad error: %s", err)
			}
		}

		// Write locked doodad?
		if s.Doodad != nil && s.Doodad.Locked {
			if balance.WriteLockOverride {
				d.Flash("Note: write lock has been overridden")
			} else {
				d.Flash("That doodad is write-protected and cannot be viewed in the editor.")
				s.Doodad = nil
				s.filename = ""
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

	// Recompute the UI Palette window for the level's palette.
	s.UI.FinishSetup(d)

	d.Flash("Editor Mode. Press 'P' to play this map.")

	return nil
}

// Playtest switches the level into Play Mode.
func (s *EditorScene) Playtest() {
	log.Info("Play Mode, Go!")
	s.d.Goto(&PlayScene{
		Filename: s.filename,
		Level:    s.Level,
		CanEdit:  true,
	})
}

// Loop the editor scene.
func (s *EditorScene) Loop(d *Doodle, ev *events.State) error {
	// Update debug overlay values.
	*s.debTool = s.UI.Canvas.Tool.String()
	*s.debSwatch = s.UI.Canvas.Palette.ActiveSwatch.Name
	*s.debWorldIndex = s.UI.Canvas.WorldIndexAt(s.UI.cursor).String()

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

	// Undo/Redo key bindings.
	if ev.ControlActive.Now {
		key := ev.KeyName.Read()
		if key == "z" {
			s.UI.Canvas.UndoStroke()
		} else if key == "y" {
			s.UI.Canvas.RedoStroke()
		}
	}

	s.UI.Loop(ev)

	// Switching to Play Mode?
	switch ev.KeyName.Read() {
	case "p":
		s.Playtest()
	case "l":
		d.Flash("Line Tool selected.")
		s.UI.Canvas.Tool = drawtool.LineTool
		s.UI.activeTool = s.UI.Canvas.Tool.String()
	case "f":
		d.Flash("Pencil Tool selected.")
		s.UI.Canvas.Tool = drawtool.PencilTool
		s.UI.activeTool = s.UI.Canvas.Tool.String()
	case "r":
		d.Flash("Rectangle Tool selected.")
		s.UI.Canvas.Tool = drawtool.RectTool
		s.UI.activeTool = s.UI.Canvas.Tool.String()
	}

	return nil
}

// Draw the current frame.
func (s *EditorScene) Draw(d *Doodle) error {
	// Clear the canvas and fill it with magenta so it's clear if any spots are missed.
	d.Engine.Clear(render.RGBA(160, 120, 160, 255))

	s.UI.Present(d.Engine)

	return nil
}

// LoadLevel loads a level from disk.
func (s *EditorScene) LoadLevel(filename string) error {
	s.filename = filename

	level, err := level.LoadFile(filename)
	fmt.Printf("%+v\n", level)
	if err != nil {
		return fmt.Errorf("EditorScene.LoadLevel(%s): %s", filename, err)
	}

	s.DrawingType = enum.LevelDrawing
	s.Level = level
	s.UI.Canvas.LoadLevel(s.d.Engine, s.Level)

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

	return m.WriteFile(filename)
}

// LoadDoodad loads a doodad from disk.
func (s *EditorScene) LoadDoodad(filename string) error {
	s.filename = filename

	doodad, err := doodads.LoadFile(filename)
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
	return d.WriteFile(filename)
}

// Destroy the scene.
func (s *EditorScene) Destroy() error {
	return nil
}
