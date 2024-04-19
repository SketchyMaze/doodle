package doodle

// Menu Bar features for Edit Mode.
// In here is the SetupMenuBar() and menu item functions.
// The rest of it is controlled in editor_ui.go

import (
	"git.kirsle.net/SketchyMaze/doodle/pkg/balance"
	"git.kirsle.net/SketchyMaze/doodle/pkg/drawtool"
	"git.kirsle.net/SketchyMaze/doodle/pkg/enum"
	"git.kirsle.net/SketchyMaze/doodle/pkg/level/giant_screenshot"
	"git.kirsle.net/SketchyMaze/doodle/pkg/log"
	"git.kirsle.net/SketchyMaze/doodle/pkg/modal"
	"git.kirsle.net/SketchyMaze/doodle/pkg/native"
	"git.kirsle.net/SketchyMaze/doodle/pkg/plus/dpp"
	"git.kirsle.net/SketchyMaze/doodle/pkg/shmem"
	"git.kirsle.net/SketchyMaze/doodle/pkg/userdir"
	"git.kirsle.net/SketchyMaze/doodle/pkg/windows"
	"git.kirsle.net/go/render"
	"git.kirsle.net/go/ui"
)

// SetupMenuBar sets up the menu bar.
func (u *EditorUI) SetupMenuBar(d *Doodle) *ui.MenuBar {
	menu := ui.NewMenuBar("Main Menu")

	// Save and Save As common menu handler
	var (
		saveFunc func(filename string)
	)

	switch u.Scene.DrawingType {
	case enum.LevelDrawing:
		saveFunc = func(filename string) {
			if err := u.Scene.SaveLevel(filename); err != nil {
				d.FlashError("Error: %s", err)
			} else {
				d.Flash("Saved level: %s", filename)
			}
		}
	case enum.DoodadDrawing:
		saveFunc = func(filename string) {
			if err := u.Scene.SaveDoodad(filename); err != nil {
				d.FlashError("Error: %s", err)
			} else {
				d.Flash("Saved doodad: %s", filename)
			}
		}
	default:
		d.FlashError("Error: Scene.DrawingType is not a valid type")
	}

	////////
	// File menu
	fileMenu := menu.AddMenu("File")
	fileMenu.AddItemAccel("New level", "Ctrl-N", u.Scene.MenuNewLevel)
	fileMenu.AddItem("New doodad", u.Scene.MenuNewDoodad)
	fileMenu.AddItemAccel("Save", "Ctrl-S", u.Scene.MenuSave(false))
	fileMenu.AddItemAccel("Save as...", "Shift-Ctrl-S", func() {
		d.Prompt("Save as filename>", func(answer string) {
			if answer != "" {
				saveFunc(answer)
			}
		})
	})

	fileMenu.AddItemAccel("Open...", "Ctrl-O", u.Scene.MenuOpen)
	fileMenu.AddSeparator()
	fileMenu.AddItem("Exit to menu", func() {
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
		if balance.DPP {
			levelMenu.AddItem("Publish", func() {
				u.OpenPublishWindow()
			})
		}

		levelMenu.AddSeparator()
		levelMenu.AddItem("Screenshot", func() {
			// It takes a LONG TIME to render for medium+ maps.
			// Do so on a background thread.
			go func() {
				filename, err := giant_screenshot.SaveCroppedScreenshot(u.Scene.Level, u.Scene.GetDrawing().Viewport())
				if err != nil {
					d.FlashError("Error: %s", err.Error())
					return
				}

				d.FlashError("Screenshot saved as: %s", filename)
			}()
		})
		levelMenu.AddItem("Giant Screenshot", func() {
			// It takes a LONG TIME to render for medium+ maps.
			modal.Confirm(
				"Do you want to make a 'Giant Screenshot' of\n" +
					"your WHOLE level? Note: this may take several\n" +
					"seconds for very large maps!",
			).WithTitle("Giant Screenshot").Then(func() {
				// Show the wait modal and generate the screenshot on a background thread.
				m := modal.Wait("Generating a giant screenshot...").WithTitle("Please hold")
				go func() {
					defer m.Dismiss(true)

					filename, err := giant_screenshot.SaveGiantScreenshot(u.Scene.Level)
					if err != nil {
						d.FlashError("Error: %s", err.Error())
						return
					}

					d.FlashError("Giant screenshot saved as: %s", filename)
				}()
			})
		})
		levelMenu.AddItem("Open screenshot folder", func() {
			native.OpenLocalURL(userdir.ScreenshotDirectory)
		})

		if balance.Feature.ViewportWindow {
			levelMenu.AddSeparator()
			levelMenu.AddItemAccel("New viewport", "v", func() {
				pip := windows.MakePiPWindow(d.width, d.height, windows.PiP{
					Supervisor: u.Supervisor,
					Engine:     u.d.Engine,
					Level:      u.Scene.Level,
					Event:      u.d.event,

					Tool:      &u.Scene.UI.Canvas.Tool,
					BrushSize: &u.Scene.UI.Canvas.BrushSize,
				})

				pip.Show()
			})
		}
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

	viewMenu.AddSeparator()

	viewMenu.AddItemAccel("Close window", "←", func() {
		u.Supervisor.CloseActiveWindow()
	})
	viewMenu.AddItemAccel("Close all windows", "Shift-←", func() {
		u.Supervisor.CloseAllWindows()
	})

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
			u.OpenDoodadDropper()
		})
		toolMenu.AddItem("Link Tool", func() {
			u.Canvas.Tool = drawtool.LinkTool
			u.activeTool = u.Canvas.Tool.String()
			d.Flash("Link Tool selected. Click a doodad in your level to link it to another.")
		})
	}

	////////
	// Help menu
	var (
		helpMenu = u.d.MakeHelpMenu(menu, u.Supervisor)
	)

	// Registration item for Doodle++ builds.
	if balance.DPP {
		var registerText = "Register"
		if dpp.Driver.IsRegistered() {
			registerText = "Registration"
		}

		helpMenu.AddSeparator()
		helpMenu.AddItem(registerText, func() {
			u.licenseWindow.Show()
			u.Supervisor.FocusWindow(u.licenseWindow)
		})
	}

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

func (s *EditorScene) MenuNewDoodad() {
	s.ConfirmUnload(func() {
		// New doodad size with prompt.
		s.d.GotoNewDoodadMenu()
	})
}

// File->Open, or Ctrl-O
func (s *EditorScene) MenuOpen() {
	s.ConfirmUnload(func() {
		if s.winOpenLevel == nil {
			s.winOpenLevel = windows.NewOpenDrawingWindow(windows.OpenDrawing{
				Supervisor: s.UI.Supervisor,
				Engine:     shmem.CurrentRenderEngine,
				OnOpenDrawing: func(filename string) {
					s.d.EditFile(filename)
				},
				OnCloseWindow: func() {
					s.winOpenLevel.Destroy()
					s.winOpenLevel = nil
				},
			})
		}
		s.winOpenLevel.MoveTo(render.Point{
			X: (s.d.width / 2) - (s.winOpenLevel.Size().W / 2),
			Y: (s.d.height / 2) - (s.winOpenLevel.Size().H / 2),
		})
		s.winOpenLevel.Show()
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
					s.d.FlashError("Error: %s", err)
				} else {
					s.d.Flash("Saved level: %s", filename)
				}
			}
		case enum.DoodadDrawing:
			// drawingType = "doodad"
			saveFunc = func(filename string) {
				if err := s.SaveDoodad(filename); err != nil {
					s.d.FlashError("Error: %s", err)
				} else {
					s.d.Flash("Saved doodad: %s", filename)
				}
			}
		default:
			s.d.FlashError("Error: Scene.DrawingType is not a valid type")
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
