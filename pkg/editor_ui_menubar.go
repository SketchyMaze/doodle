package doodle

// Menu Bar features for Edit Mode.
// In here is the SetupMenuBar() and menu item functions.
// The rest of it is controlled in editor_ui.go

import (
	"strconv"

	"git.kirsle.net/apps/doodle/pkg/balance"
	"git.kirsle.net/apps/doodle/pkg/drawtool"
	"git.kirsle.net/apps/doodle/pkg/enum"
	"git.kirsle.net/apps/doodle/pkg/log"
	"git.kirsle.net/apps/doodle/pkg/native"
	"git.kirsle.net/apps/doodle/pkg/windows"
	"git.kirsle.net/go/render"
	"git.kirsle.net/go/ui"
)

// SetupMenuBar sets up the menu bar.
func (u *EditorUI) SetupMenuBar(d *Doodle) *ui.MenuBar {
	menu := ui.NewMenuBar("Main Menu")

	// Save and Save As common menu handler
	var (
		drawingType string
		saveFunc    func(filename string)
	)

	switch u.Scene.DrawingType {
	case enum.LevelDrawing:
		drawingType = "level"
		saveFunc = func(filename string) {
			if err := u.Scene.SaveLevel(filename); err != nil {
				d.Flash("Error: %s", err)
			} else {
				d.Flash("Saved level: %s", filename)
			}
		}
	case enum.DoodadDrawing:
		drawingType = "doodad"
		saveFunc = func(filename string) {
			if err := u.Scene.SaveDoodad(filename); err != nil {
				d.Flash("Error: %s", err)
			} else {
				d.Flash("Saved doodad: %s", filename)
			}
		}
	default:
		d.Flash("Error: Scene.DrawingType is not a valid type")
	}

	////////
	// File menu
	fileMenu := menu.AddMenu("File")
	fileMenu.AddItemAccel("New level", "Ctrl-N", u.Scene.MenuNewLevel)
	fileMenu.AddItem("New doodad", func() {
		u.Scene.ConfirmUnload(func() {
			d.Prompt("Doodad size [100]>", func(answer string) {
				size := balance.DoodadSize
				if answer != "" {
					i, err := strconv.Atoi(answer)
					if err != nil {
						d.Flash("Error: Doodad size must be a number.")
						return
					}
					size = i
				}
				d.NewDoodad(size)
			})
		})
	})
	fileMenu.AddItemAccel("Save", "Ctrl-S", u.Scene.MenuSave(false))
	fileMenu.AddItemAccel("Save as...", "Shift-Ctrl-S", func() {
		d.Prompt("Save as filename>", func(answer string) {
			if answer != "" {
				saveFunc(answer)
			}
		})
	})

	if balance.Feature.EmbeddableDoodads && drawingType == "level" {
		fileMenu.AddItem("Publish level", func() {
			u.OpenPublishWindow()
		})
	}

	fileMenu.AddItemAccel("Open...", "Ctrl-O", u.Scene.MenuOpen)
	fileMenu.AddSeparator()
	fileMenu.AddItem("Close "+drawingType, func() {
		u.Scene.ConfirmUnload(func() {
			d.Goto(&MainScene{})
		})
	})
	fileMenu.AddItemAccel("Quit", "Escape", func() {
		d.ConfirmExit()
	})

	////////
	// Edit menu
	editMenu := menu.AddMenu("Edit")
	editMenu.AddItemAccel("Undo", "Ctrl-Z", func() {
		u.Canvas.UndoStroke()
	})
	editMenu.AddItemAccel("Redo", "Ctrl-Y", func() {
		u.Canvas.RedoStroke()
	})
	editMenu.AddSeparator()
	editMenu.AddItem("Settings", func() {
		if u.settingsWindow == nil {
			u.settingsWindow = d.MakeSettingsWindow(u.Supervisor)
		}
		u.settingsWindow.Show()
	})

	////////
	// Level menu
	if u.Scene.DrawingType == enum.LevelDrawing {
		levelMenu := menu.AddMenu("Level")
		levelMenu.AddItem("Level Properties", func() {
			log.Info("Opening the window")

			// Open the New Level window in edit-settings mode.
			u.levelSettingsWindow.Hide()
			u.levelSettingsWindow = nil
			u.SetupPopups(u.d)
			u.levelSettingsWindow.Show()
		})
		levelMenu.AddItem("Attached files", func() {
			log.Info("Opening the FileSystem window")
			u.OpenFileSystemWindow()
		})
		levelMenu.AddItemAccel("Playtest", "P", func() {
			u.Scene.Playtest()
		})
	}

	////////
	// Doodad Menu
	if u.Scene.DrawingType == enum.DoodadDrawing {
		levelMenu := menu.AddMenu("Doodad")
		levelMenu.AddItem("Doodad Properties", func() {
			log.Info("Opening the window")

			// Open the New Level window in edit-settings mode.
			u.doodadPropertiesWindow.Hide()
			u.doodadPropertiesWindow = nil
			u.SetupPopups(u.d)
			u.doodadPropertiesWindow.Show()
		})

		levelMenu.AddItem("Layers", func() {
			u.OpenLayersWindow()
		})
	}

	////////
	// View menu
	if balance.Feature.Zoom {
		viewMenu := menu.AddMenu("View")
		viewMenu.AddItemAccel("Zoom in", "+", func() {
			u.Canvas.Zoom++
		})
		viewMenu.AddItemAccel("Zoom out", "-", func() {
			u.Canvas.Zoom--
		})
		viewMenu.AddItemAccel("Reset zoom", "1", func() {
			u.Canvas.Zoom = 0
		})
		viewMenu.AddItemAccel("Scroll drawing to origin", "0", func() {
			u.Canvas.ScrollTo(render.Origin)
		})
	}

	////////
	// Tools menu
	toolMenu := menu.AddMenu("Tools")
	toolMenu.AddItemAccel("Debug overlay", "F3", func() {
		DebugOverlay = !DebugOverlay
		if DebugOverlay {
			d.Flash("Debug overlay enabled. Press F3 to turn it off.")
		}
	})
	toolMenu.AddItemAccel("Command shell", "`", func() {
		d.shell.Open = true
	})
	toolMenu.AddSeparator()
	toolMenu.AddItem("Edit Palette", func() {
		u.OpenPaletteWindow()
	})

	// Draw Tools
	toolMenu.AddItemAccel("Pencil Tool", "F", func() {
		u.Canvas.Tool = drawtool.PencilTool
		u.activeTool = u.Canvas.Tool.String()
		d.Flash("Pencil Tool selected.")
	})
	toolMenu.AddItemAccel("Line Tool", "L", func() {
		u.Canvas.Tool = drawtool.LineTool
		u.activeTool = u.Canvas.Tool.String()
		d.Flash("Line Tool selected.")
	})
	toolMenu.AddItemAccel("Rectangle Tool", "R", func() {
		u.Canvas.Tool = drawtool.RectTool
		u.activeTool = u.Canvas.Tool.String()
		d.Flash("Rectangle Tool selected.")
	})
	toolMenu.AddItemAccel("Ellipse Tool", "C", func() {
		u.Canvas.Tool = drawtool.EllipseTool
		u.activeTool = u.Canvas.Tool.String()
		d.Flash("Ellipse Tool selected.")
	})
	toolMenu.AddItemAccel("Eraser Tool", "x", func() {
		u.Canvas.Tool = drawtool.EraserTool
		u.activeTool = u.Canvas.Tool.String()
		d.Flash("Eraser Tool selected.")
	})

	if u.Scene.DrawingType == enum.LevelDrawing {
		toolMenu.AddItemAccel("Doodads", "q", func() {
			log.Info("Open the DoodadDropper")
			u.doodadWindow.Show()
		})
		toolMenu.AddItem("Link Tool", func() {
			u.Canvas.Tool = drawtool.LinkTool
			u.activeTool = u.Canvas.Tool.String()
			d.Flash("Link Tool selected. Click a doodad in your level to link it to another.")
		})
	}

	////////
	// Help menu
	helpMenu := menu.AddMenu("Help")
	helpMenu.AddItemAccel("User Manual", "F1", func() {
		native.OpenLocalURL(balance.GuidebookPath)
	})
	helpMenu.AddItem("Register", func() {
		u.licenseWindow.Show()
	})
	helpMenu.AddItem("About", func() {
		if u.aboutWindow == nil {
			u.aboutWindow = windows.NewAboutWindow(windows.About{
				Supervisor: u.Supervisor,
				Engine:     d.Engine,
			})
			u.aboutWindow.Compute(d.Engine)
			u.aboutWindow.Supervise(u.Supervisor)

			// Center the window.
			u.aboutWindow.MoveTo(render.Point{
				X: (d.width / 2) - (u.aboutWindow.Size().W / 2),
				Y: 60,
			})
		}
		u.aboutWindow.Show()
	})

	menu.Supervise(u.Supervisor)
	menu.Compute(d.Engine)

	return menu
}

// Menu functions that have keybind callbacks below.

// File->New level, or Ctrl-N
func (s *EditorScene) MenuNewLevel() {
	s.ConfirmUnload(func() {
		s.d.GotoNewMenu()
	})
}

// File->Open, or Ctrl-O
func (s *EditorScene) MenuOpen() {
	s.ConfirmUnload(func() {
		s.d.GotoLoadMenu()
	})
}

// File->Save, or Ctrl-S
// File->Save As, or Shift-Ctrl-S
// NOTICE: this one returns a func() so you need to call that one!
func (s *EditorScene) MenuSave(as bool) func() {
	return func() {
		var (
			// drawingType string
			saveFunc func(filename string)
		)

		switch s.DrawingType {
		case enum.LevelDrawing:
			// drawingType = "level"
			saveFunc = func(filename string) {
				if err := s.SaveLevel(filename); err != nil {
					s.d.Flash("Error: %s", err)
				} else {
					s.d.Flash("Saved level: %s", filename)
				}
			}
		case enum.DoodadDrawing:
			// drawingType = "doodad"
			saveFunc = func(filename string) {
				if err := s.SaveDoodad(filename); err != nil {
					s.d.Flash("Error: %s", err)
				} else {
					s.d.Flash("Saved doodad: %s", filename)
				}
			}
		default:
			s.d.Flash("Error: Scene.DrawingType is not a valid type")
		}

		// "Save As"?
		if as {
			s.d.Prompt("Save as filename>", func(answer string) {
				if answer != "" {
					saveFunc(answer)
				}
			})
			return
		}

		// "Save", write to existing filename or prompt for it.
		if s.filename != "" {
			saveFunc(s.filename)
		} else {
			s.d.Prompt("Save filename>", func(answer string) {
				if answer != "" {
					saveFunc(answer)
				}
			})
		}
	}
}
