package doodle

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"git.kirsle.net/SketchyMaze/doodle/pkg/balance"
	"git.kirsle.net/SketchyMaze/doodle/pkg/cursor"
	"git.kirsle.net/SketchyMaze/doodle/pkg/doodads"
	"git.kirsle.net/SketchyMaze/doodle/pkg/drawtool"
	"git.kirsle.net/SketchyMaze/doodle/pkg/enum"
	"git.kirsle.net/SketchyMaze/doodle/pkg/keybind"
	"git.kirsle.net/SketchyMaze/doodle/pkg/level"
	"git.kirsle.net/SketchyMaze/doodle/pkg/level/giant_screenshot"
	"git.kirsle.net/SketchyMaze/doodle/pkg/level/publishing"
	"git.kirsle.net/SketchyMaze/doodle/pkg/license"
	"git.kirsle.net/SketchyMaze/doodle/pkg/log"
	"git.kirsle.net/SketchyMaze/doodle/pkg/modal"
	"git.kirsle.net/SketchyMaze/doodle/pkg/modal/loadscreen"
	"git.kirsle.net/SketchyMaze/doodle/pkg/native"
	"git.kirsle.net/SketchyMaze/doodle/pkg/usercfg"
	"git.kirsle.net/SketchyMaze/doodle/pkg/userdir"
	"git.kirsle.net/SketchyMaze/doodle/pkg/windows"
	"git.kirsle.net/go/render"
	"git.kirsle.net/go/render/event"
)

// EditorScene manages the "Edit Level" game mode.
type EditorScene struct {
	// Configuration for the scene initializer.
	DrawingType            enum.DrawingType
	OpenFile               bool
	Filename               string
	DoodadSize             render.Rect
	RememberScrollPosition render.Point // Play mode remembers it for us

	UI *EditorUI
	d  *Doodle

	// The current level or doodad object being edited, based on the
	// DrawingType.
	Level       *level.Level
	Doodad      *doodads.Doodad
	ActiveLayer int // which layer (of a doodad) is being edited now?

	// Custom debug overlay values.
	debTool            *string
	debSwatch          *string
	debWorldIndex      *string
	debLoadingViewport *string

	// Last saved filename by the user.
	filename string

	lastAutosaveAt time.Time
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
	s.debLoadingViewport = new(string)
	customDebugLabels = []debugLabel{
		{"Pixel:", s.debWorldIndex},
		{"Tool:", s.debTool},
		{"Swatch:", s.debSwatch},
		{"Chunks:", s.debLoadingViewport},
	}

	// Initialize autosave time.
	s.lastAutosaveAt = time.Now()

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

// Reset the editor scene from scratch. Good nuclear option when you change the level's
// palette on-the-fly or some other sticky situation and want to reload the editor.
func (s *EditorScene) Reset() {
	if s.Level != nil {
		s.Level.Chunker.Redraw()
	}
	if s.Doodad != nil {
		s.Doodad.Layers[s.ActiveLayer].Chunker.Redraw()
	}

	s.d.Goto(&EditorScene{
		Filename: s.Filename,
		Level:    s.Level,
		Doodad:   s.Doodad,
	})
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
			s.UI.Canvas.LoadLevel(s.Level)
			if err := s.installActors(); err != nil {
				log.Error("InstallActors: %s", err)
			}
		} else if s.filename != "" && s.OpenFile {
			log.Debug("EditorScene.Setup: Loading map from filename at %s", s.filename)
			loadscreen.SetSubtitle(
				"Opening: " + s.filename,
			)
			if err := s.LoadLevel(s.filename); err != nil {
				d.FlashError("LoadLevel error: %s", err)
			} else {
				if err := s.installActors(); err != nil {
					log.Error("InstallActors: %s", err)
				}
			}
		}

		// Write locked level?
		if s.Level != nil && s.Level.Locked {
			if usercfg.Current.WriteLockOverride {
				d.Flash("Note: write lock has been overridden")
			} else {
				d.FlashError("That level is write-protected and cannot be viewed in the editor.")
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
			s.UI.Canvas.LoadLevel(s.Level)
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
				d.FlashError("LoadDoodad error: %s", err)
			}
		}

		// Write locked doodad?
		if s.Doodad != nil && s.Doodad.Locked {
			if usercfg.Current.WriteLockOverride {
				d.Flash("Note: write lock has been overridden")
			} else {
				d.FlashError("That doodad is write-protected and cannot be viewed in the editor.")
				s.Doodad = nil
				s.filename = ""
			}
		}

		// No Doodad?
		if s.Doodad == nil {
			log.Debug("EditorScene.Setup: initializing a new Doodad")
			s.Doodad = doodads.New(s.DoodadSize.W, s.DoodadSize.H)
			s.UI.Canvas.LoadDoodad(s.Doodad)
		}

		// Update the loading screen with level info.
		loadscreen.SetSubtitle(
			s.Doodad.Title,
			"by "+s.Doodad.Author,
		)

		// TODO: move inside the UI. Just an approximate position for now.
		s.UI.Canvas.Resize(s.DoodadSize)
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

	// Scroll the level to the remembered position from when we went
	// to Play Mode and back. If no remembered position, this is zero
	// anyway.
	if s.RememberScrollPosition.IsZero() && s.Level != nil {
		s.UI.Canvas.ScrollTo(s.Level.ScrollPosition)
	} else {
		s.UI.Canvas.ScrollTo(s.RememberScrollPosition)
	}

	d.Flash("Editor Mode.")
	if s.DrawingType == enum.LevelDrawing {
		d.Flash("Press 'P' to playtest this level.")
	}

	return nil
}

// Common function to install the actors into the level.
//
// InstallActors may return an error if doodads were not found - because the
// player is on the free version and can't load attached doodads from nonsigned
// files.
func (s *EditorScene) installActors() error {
	if err := s.UI.Canvas.InstallActors(s.Level.Actors); err != nil {
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

// Playtest switches the level into Play Mode.
func (s *EditorScene) Playtest() {
	log.Info("Play Mode, Go!")
	s.d.Goto(&PlayScene{
		Filename:               s.filename,
		Level:                  s.Level,
		CanEdit:                true,
		RememberScrollPosition: s.UI.Canvas.Scroll,
	})
}

// PlaytestFrom enters play mode starting at a custom spawn point.
func (s *EditorScene) PlaytestFrom(p render.Point) {
	log.Info("Play Mode, Go!")
	s.d.Goto(&PlayScene{
		Filename:               s.filename,
		Level:                  s.Level,
		CanEdit:                true,
		RememberScrollPosition: s.UI.Canvas.Scroll,
		SpawnPoint:             p,
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
	*s.debLoadingViewport = "???"

	// Safely...
	if s.UI.Canvas.Palette != nil && s.UI.Canvas.Palette.ActiveSwatch != nil {
		*s.debSwatch = s.UI.Canvas.Palette.ActiveSwatch.Name
	}
	if s.UI.Canvas != nil {
		inside, outside := s.UI.Canvas.LoadUnloadMetrics()
		*s.debLoadingViewport = fmt.Sprintf("%d in %d out %d cached %d gc", inside, outside, s.UI.Canvas.Chunker().CacheSize(), s.UI.Canvas.Chunker().GCSize())
	}

	// Has the window been resized?
	if ev.WindowResized {
		s.UI.Resized(d)
		return nil
	}

	// Run all of the keybinds.
	binders := []struct {
		v bool
		f func()
	}{
		{
			keybind.NewLevel(ev), func() {
				// Ctrl-N, New Level
				s.MenuNewLevel()
			},
		},
		{
			keybind.SaveAs(ev), func() {
				// Shift-Ctrl-S, Save As
				s.MenuSave(true)()
			},
		},
		{
			keybind.Save(ev), func() {
				// Ctrl-S, Save
				s.MenuSave(false)()
			},
		},
		{
			keybind.Open(ev), func() {
				// Ctrl-O, Open
				s.MenuOpen()
			},
		},
		{
			keybind.Undo(ev), func() {
				// Ctrl-Z, Undo
				s.UI.Canvas.UndoStroke()
				ev.ResetKeyDown()
			},
		},
		{
			keybind.Redo(ev), func() {
				// Ctrl-Y, Undo
				s.UI.Canvas.RedoStroke()
				ev.ResetKeyDown()
			},
		},
		{

			keybind.ZoomIn(ev), func() {
				s.UI.Canvas.Zoom++
				ev.ResetKeyDown()
			},
		},
		{
			keybind.ZoomOut(ev), func() {
				s.UI.Canvas.Zoom--
				ev.ResetKeyDown()
			},
		},
		{
			keybind.ZoomReset(ev), func() {
				s.UI.Canvas.Zoom = 0
				ev.ResetKeyDown()
			},
		},
		{
			keybind.Origin(ev), func() {
				d.Flash("Scrolled back to level origin (0,0)")
				s.UI.Canvas.ScrollTo(render.Origin)
				ev.ResetKeyDown()
			},
		},
		{
			keybind.CloseAllWindows(ev), func() {
				s.UI.Supervisor.CloseAllWindows()
			},
		},
		{
			keybind.CloseTopmostWindow(ev), func() {
				s.UI.Supervisor.CloseActiveWindow()
			},
		},
		{
			keybind.NewViewport(ev), func() {
				if s.DrawingType != enum.LevelDrawing {
					return
				}

				pip := windows.MakePiPWindow(d.width, d.height, windows.PiP{
					Supervisor: s.UI.Supervisor,
					Engine:     s.d.Engine,
					Level:      s.Level,
					Event:      s.d.event,

					Tool:      &s.UI.Canvas.Tool,
					BrushSize: &s.UI.Canvas.BrushSize,
				})

				pip.Show()
			},
		},
	}
	for _, bind := range binders {
		if bind.v {
			bind.f()
		}
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
		s.UI.OpenDoodadDropper()
	}

	s.UI.Loop(ev)

	// Trigger auto-save of the level in case of crash or accidental closure.
	if time.Since(s.lastAutosaveAt) > balance.AutoSaveInterval {
		s.lastAutosaveAt = time.Now()
		if !usercfg.Current.DisableAutosave {
			if err := s.AutoSave(); err != nil {
				d.FlashError("Autosave error: %s", err)
			}
		}
	}

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
	s.UI.Canvas.LoadLevel(s.Level)

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
		m.Author = native.DefaultAuthor()
	}

	m.Palette = s.UI.Canvas.Palette
	m.Chunker = s.UI.Canvas.Chunker()

	// Store the scroll position.
	m.ScrollPosition = s.UI.Canvas.Scroll

	// Clear the modified flag on the level.
	s.UI.Canvas.SetModified(false)

	// Attach doodads to the level on save.
	if err := publishing.Publish(m); err != nil {
		log.Error("Error publishing level: %s", err.Error())
	}

	s.lastAutosaveAt = time.Now()

	// Save screenshots into the level file.
	if !m.HasScreenshot() {
		s.UpdateLevelScreenshot(m)
	}

	return m.WriteFile(filename)
}

// UpdateLevelScreenshot updates a level screenshot in its zipfile.
func (s *EditorScene) UpdateLevelScreenshot(lvl *level.Level) error {
	// The level must have been saved and have a filename to update in.
	if s.filename == "" {
		return errors.New("Save your level to disk before updating its screenshot.")
	}

	if err := giant_screenshot.UpdateLevelScreenshots(lvl, s.UI.Canvas.Viewport().Point()); err != nil {
		return fmt.Errorf("Error saving level screenshots: %s", err)
	}
	return nil
}

// AutoSave takes an autosave snapshot of the level or drawing.
func (s *EditorScene) AutoSave() error {
	var (
		filename = "_autosave.level"
		err      error
	)

	s.d.FlashError("Beginning AutoSave() in a background thread")

	// Trigger the auto-save in the background to not block the main thread.
	go func() {
		var err error
		switch s.DrawingType {
		case enum.LevelDrawing:
			err = s.Level.WriteFile(filename)
			s.d.Flash("Automatically saved level to %s", filename)
		case enum.DoodadDrawing:
			filename = "_autosave.doodad"
			err = s.Doodad.WriteFile(filename)
			s.d.Flash("Automatically saved doodad to %s", filename)
		}

		if err != nil {
			s.d.FlashError("Error saving %s: %s", filename, err)
		}
	}()

	return err
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
	s.DoodadSize = doodad.Size
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
		d.Author = native.DefaultAuthor()
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
	// Free SDL2 textures. Note: if they are switching to the Editor, the chunks still have
	// their bitmaps cached and will regen the textures as needed.
	s.UI.Teardown()

	// Reset the cursor to default.
	cursor.Current = cursor.NewPointer(s.d.Engine)

	return nil
}
