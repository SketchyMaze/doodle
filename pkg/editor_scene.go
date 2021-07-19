package doodle

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"git.kirsle.net/apps/doodle/pkg/balance"
	"git.kirsle.net/apps/doodle/pkg/doodads"
	"git.kirsle.net/apps/doodle/pkg/drawtool"
	"git.kirsle.net/apps/doodle/pkg/enum"
	"git.kirsle.net/apps/doodle/pkg/keybind"
	"git.kirsle.net/apps/doodle/pkg/level"
	"git.kirsle.net/apps/doodle/pkg/license"
	"git.kirsle.net/apps/doodle/pkg/log"
	"git.kirsle.net/apps/doodle/pkg/modal"
	"git.kirsle.net/apps/doodle/pkg/modal/loadscreen"
	"git.kirsle.net/apps/doodle/pkg/usercfg"
	"git.kirsle.net/apps/doodle/pkg/userdir"
	"git.kirsle.net/go/render"
	"git.kirsle.net/go/render/event"
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
	Level       *level.Level
	Doodad      *doodads.Doodad
	ActiveLayer int // which layer (of a doodad) is being edited now?

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

	// Show the loading screen.
	loadscreen.ShowWithProgress()
	go func() {
		if err := s.setupAsync(d); err != nil {
			log.Error("EditorScene.setupAsync: %s", err)
		}
		loadscreen.Hide()
	}()

	return nil
}

// setupAsync initializes trhe editor scene in the background,
// underneath a loading screen.
func (s *EditorScene) setupAsync(d *Doodle) error {
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
			loadscreen.SetSubtitle(
				"Opening: "+s.Level.Title,
				"by "+s.Level.Author,
			)
			s.UI.Canvas.LoadLevel(d.Engine, s.Level)
			s.UI.Canvas.InstallActors(s.Level.Actors)
		} else if s.filename != "" && s.OpenFile {
			log.Debug("EditorScene.Setup: Loading map from filename at %s", s.filename)
			loadscreen.SetSubtitle(
				"Opening: " + s.filename,
			)
			if err := s.LoadLevel(s.filename); err != nil {
				d.Flash("LoadLevel error: %s", err)
			} else {
				s.UI.Canvas.InstallActors(s.Level.Actors)
			}
		}

		// Write locked level?
		if s.Level != nil && s.Level.Locked {
			if usercfg.Current.WriteLockOverride {
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

		// Update the loading screen with level info.
		loadscreen.SetSubtitle(
			"Opening: "+s.Level.Title,
			"by "+s.Level.Author,
		)
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
			if usercfg.Current.WriteLockOverride {
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

		// Update the loading screen with level info.
		loadscreen.SetSubtitle(
			s.Doodad.Title,
			"by "+s.Doodad.Author,
		)

		// TODO: move inside the UI. Just an approximate position for now.
		s.UI.Canvas.Resize(render.NewRect(s.DoodadSize, s.DoodadSize))
		s.UI.Canvas.ScrollTo(render.Origin)
		s.UI.Canvas.Scrollable = false
		s.UI.Workspace.Compute(d.Engine)
	}

	// Pre-cache all bitmap images from the level chunks.
	// Note: we are not running on the main thread, so SDL2 Textures
	// don't get created yet, but we do the full work of caching bitmap
	// images which later get fed directly into SDL2 saving speed at
	// runtime, + the bitmap generation is pretty wicked fast anyway.
	loadscreen.PreloadAllChunkBitmaps(s.UI.Canvas.Chunker())

	// Recompute the UI Palette window for the level's palette.
	s.UI.FinishSetup(d)

	d.Flash("Editor Mode.")
	if s.DrawingType == enum.LevelDrawing {
		d.Flash("Press 'P' to playtest this level.")
	}

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

// ConfirmUnload may pop up a confirmation modal to save the level before the
// user performs an action that may close the level, such as click File->New.
func (s *EditorScene) ConfirmUnload(fn func()) {
	if !s.UI.Canvas.Modified() {
		fn()
		return
	}

	modal.Confirm(
		"This drawing has unsaved changes. Are you sure you\nwant to continue and lose your changes?",
	).WithTitle("Confirm Closing Drawing").Then(fn)
}

// Loop the editor scene.
func (s *EditorScene) Loop(d *Doodle, ev *event.State) error {
	// Skip if still loading.
	if loadscreen.IsActive() {
		return nil
	}

	// Update debug overlay values.
	*s.debTool = s.UI.Canvas.Tool.String()
	*s.debSwatch = "???"
	*s.debWorldIndex = s.UI.Canvas.WorldIndexAt(s.UI.cursor).String()

	// Safely...
	if s.UI.Canvas.Palette != nil && s.UI.Canvas.Palette.ActiveSwatch != nil {
		*s.debSwatch = s.UI.Canvas.Palette.ActiveSwatch.Name
	}

	// Has the window been resized?
	if ev.WindowResized {
		w, h := d.Engine.WindowSize()
		if w != d.width || h != d.height {
			// Not a false alarm.
			d.width = w
			d.height = h
			s.UI.Resized(d)
			return nil
		}
	}

	// Menu key bindings.
	if keybind.NewLevel(ev) {
		// Ctrl-N, New Level
		s.MenuNewLevel()
	} else if keybind.SaveAs(ev) {
		// Shift-Ctrl-S, Save As
		s.MenuSave(true)()
	} else if keybind.Save(ev) {
		// Ctrl-S, Save
		s.MenuSave(false)()
	} else if keybind.Open(ev) {
		// Ctrl-O, Open
		s.MenuOpen()
	}

	// Undo/Redo key bindings.
	if keybind.Undo(ev) {
		s.UI.Canvas.UndoStroke()
		ev.ResetKeyDown()
	} else if keybind.Redo(ev) {
		s.UI.Canvas.RedoStroke()
		ev.ResetKeyDown()
	}

	// Zoom in/out.
	if balance.Feature.Zoom {
		if keybind.ZoomIn(ev) {
			d.Flash("Zoom in")
			s.UI.Canvas.Zoom++
			ev.ResetKeyDown()
		} else if keybind.ZoomOut(ev) {
			d.Flash("Zoom out")
			s.UI.Canvas.Zoom--
			ev.ResetKeyDown()
		} else if keybind.ZoomReset(ev) {
			d.Flash("Reset zoom")
			s.UI.Canvas.Zoom = 0
			ev.ResetKeyDown()
		}
	}

	// More keybinds
	if keybind.Origin(ev) {
		d.Flash("Scrolled back to level origin (0,0)")
		s.UI.Canvas.ScrollTo(render.Origin)
		ev.ResetKeyDown()
	}

	// s.UI.Loop(ev)

	// Switching to Play Mode?
	if s.DrawingType == enum.LevelDrawing && keybind.GotoPlay(ev) {
		s.Playtest()
	} else if keybind.LineTool(ev) {
		d.Flash("Line Tool selected.")
		s.UI.Canvas.Tool = drawtool.LineTool
		s.UI.activeTool = s.UI.Canvas.Tool.String()
	} else if keybind.PencilTool(ev) {
		d.Flash("Pencil Tool selected.")
		s.UI.Canvas.Tool = drawtool.PencilTool
		s.UI.activeTool = s.UI.Canvas.Tool.String()
	} else if keybind.RectTool(ev) {
		d.Flash("Rectangle Tool selected.")
		s.UI.Canvas.Tool = drawtool.RectTool
		s.UI.activeTool = s.UI.Canvas.Tool.String()
	} else if keybind.EllipseTool(ev) {
		d.Flash("Ellipse Tool selected.")
		s.UI.Canvas.Tool = drawtool.EllipseTool
		s.UI.activeTool = s.UI.Canvas.Tool.String()
	} else if keybind.EraserTool(ev) {
		d.Flash("Eraser Tool selected.")
		s.UI.Canvas.Tool = drawtool.EraserTool
		s.UI.activeTool = s.UI.Canvas.Tool.String()
	} else if keybind.DoodadDropper(ev) {
		s.UI.doodadWindow.Show()
	}

	s.UI.Loop(ev)

	return nil
}

// Draw the current frame.
func (s *EditorScene) Draw(d *Doodle) error {
	// Skip if still loading.
	if loadscreen.IsActive() {
		return nil
	}

	// Clear the canvas and fill it with magenta so it's clear if any spots are missed.
	d.Engine.Clear(render.RGBA(160, 120, 160, 255))

	s.UI.Present(d.Engine)

	return nil
}

// LoadLevel loads a level from disk.
func (s *EditorScene) LoadLevel(filename string) error {
	s.filename = filename

	level, err := level.LoadFile(filename)
	if err != nil {
		return fmt.Errorf("EditorScene.LoadLevel(%s): %s", filename, err)
	}

	s.DrawingType = enum.LevelDrawing
	s.Level = level
	s.UI.Canvas.LoadLevel(s.d.Engine, s.Level)

	log.Info("Installing %d actors into the drawing", len(level.Actors))
	if err := s.UI.Canvas.InstallActors(level.Actors); err != nil {
		summary := "This level references some doodads that were not found:"
		if strings.Contains(err.Error(), license.ErrRegisteredFeature.Error()) {
			summary = "This level contains embedded doodads, but this is not\n" +
				"available in the free version of the game. The following\n" +
				"doodads could not be loaded:"
		}
		modal.Alert("%s\n\n%s", summary, err).WithTitle("Level Errors")
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

	// Clear the modified flag on the level.
	s.UI.Canvas.SetModified(false)

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
	d.Layers[s.ActiveLayer].Chunker = s.UI.Canvas.Chunker()

	// Clear the modified flag on the level.
	s.UI.Canvas.SetModified(false)

	// Save it to their profile directory.
	filename = userdir.DoodadPath(filename)
	log.Info("Write Doodad: %s", filename)
	return d.WriteFile(filename)
}

// Destroy the scene.
func (s *EditorScene) Destroy() error {
	return nil
}
