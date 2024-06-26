package doodle

import (
	"fmt"
	"path/filepath"
	"strings"

	"git.kirsle.net/SketchyMaze/doodle/pkg/balance"
	"git.kirsle.net/SketchyMaze/doodle/pkg/doodads"
	"git.kirsle.net/SketchyMaze/doodle/pkg/drawtool"
	"git.kirsle.net/SketchyMaze/doodle/pkg/level"
	"git.kirsle.net/SketchyMaze/doodle/pkg/log"
	"git.kirsle.net/SketchyMaze/doodle/pkg/modal"
	"git.kirsle.net/SketchyMaze/doodle/pkg/plus/dpp"
	"git.kirsle.net/SketchyMaze/doodle/pkg/windows"
	"git.kirsle.net/go/render"
	"git.kirsle.net/go/ui"
)

/*
Functions to manage popup windows in the Editor Mode, such as:

* The Palette Editor
* The Layers Window
* etc.
*/

// Opens the "Layers" window (for editing doodads)
func (u *EditorUI) OpenLayersWindow() {
	u.layersWindow.Close()
	u.layersWindow = nil
	u.SetupPopups(u.d)
	u.layersWindow.Show()
}

// OpenPaletteWindow opens the Palette Editor window.
func (u *EditorUI) OpenPaletteWindow() {
	// TODO: recompute the window so the actual loaded level palette gets in
	u.paletteEditor.Close()
	u.paletteEditor = nil
	u.SetupPopups(u.d)
	u.paletteEditor.Show()
}

// OpenDoodadDropper opens the Doodad Dropper window.
func (u *EditorUI) OpenDoodadDropper() {
	// NOTE: most places in the code call this directly, nice
	// and simple window :) but OpenDoodadDropper() added for consistency.
	u.doodadWindow.Show()
	u.Supervisor.FocusWindow(u.doodadWindow)
}

// OpenTextTool opens the Text Tool window.
func (u *EditorUI) OpenTextTool() {
	u.textToolWindow.Show()
	u.Supervisor.FocusWindow(u.textToolWindow)
}

// OpenPublishWindow opens the Publisher window.
func (u *EditorUI) OpenPublishWindow() {
	scene, _ := u.d.Scene.(*EditorScene)

	u.publishWindow = windows.NewPublishWindow(windows.Publish{
		Supervisor: u.Supervisor,
		Engine:     u.d.Engine,
		Level:      scene.Level,

		OnPublish: func(includeBuiltins bool) {
			u.d.FlashError("OnPublish Called")
			// XXX: Paid Version Only.
			if !dpp.Driver.IsRegistered() {
				if u.licenseWindow != nil {
					u.licenseWindow.Show()
					u.Supervisor.FocusWindow(u.licenseWindow)
				}
				u.d.FlashError("Level Publishing is only available in the full version of the game.")
				return
			}

			// NOTE: this function just saves the level. SaveDoodads and SaveBuiltins
			// are toggled in the publish window and the save handler does publishing.
			u.Scene.SaveLevel(u.Scene.filename)
			u.d.Flash("Saved level: %s", u.Scene.filename)
		},
		OnCancel: func() {
			u.publishWindow.Close()
		},
	})
	u.ConfigureWindow(u.d, u.publishWindow)

	u.publishWindow.Close()
	// u.publishWindow = nil
	u.SetupPopups(u.d)
	u.publishWindow.Show()
}

// OpenPublishWindow opens the FileSystem window.
func (u *EditorUI) OpenFileSystemWindow() {
	u.filesystemWindow.Close()
	u.filesystemWindow = nil
	u.SetupPopups(u.d)
	u.filesystemWindow.Show()
}

// ConfigureWindow sets default window config functions, like
// centering them on screen.
func (u *EditorUI) ConfigureWindow(d *Doodle, window *ui.Window) {
	var size = window.Size()
	window.Compute(d.Engine)
	window.Supervise(u.Supervisor)

	// Center the window.
	window.MoveTo(render.Point{
		X: (d.width / 2) - (size.W / 2),
		Y: (d.height / 2) - (size.H / 2),
	})

	window.Close()
}

// SetupPopups preloads popup windows like the DoodadDropper.
func (u *EditorUI) SetupPopups(d *Doodle) {
	// License Registration Window.
	if u.licenseWindow == nil {
		cfg := windows.License{
			Supervisor: u.Supervisor,
			Engine:     d.Engine,
			OnCancel: func() {
				u.licenseWindow.Close()
			},
		}
		cfg.OnLicensed = func() {
			// License status has changed, reload the window!
			if u.licenseWindow != nil {
				u.licenseWindow.Close()
			}
			u.licenseWindow = windows.MakeLicenseWindow(d.width, d.height, cfg)
		}

		cfg.OnLicensed()
		u.licenseWindow.Close()
	}

	// Doodad Dropper.
	if u.doodadWindow == nil {
		u.doodadWindow = windows.NewDoodadDropper(windows.DoodadDropper{
			Supervisor: u.Supervisor,
			Engine:     d.Engine,

			OnStartDragActor: u.startDragActor,
			OnCancel: func() {
				u.doodadWindow.Close()
			},
		})
		u.ConfigureWindow(d, u.doodadWindow)
	}

	// Text Tool window.
	if u.textToolWindow == nil {
		u.textToolWindow = windows.NewTextToolWindow(windows.TextTool{
			Supervisor: u.Supervisor,
			Engine:     d.Engine,
			OnChangeSettings: func(font string, size int, message string) {
				log.Info("Updated Text Tool settings: %s (%d): %s", font, size, message)
				drawtool.TT = drawtool.TextSettings{
					Font:    font,
					Size:    size,
					Message: message,
				}
			},
		})
		u.ConfigureWindow(d, u.textToolWindow)
	}

	// Page Settings
	if u.levelSettingsWindow == nil {
		scene, _ := d.Scene.(*EditorScene)

		u.levelSettingsWindow = windows.NewAddEditLevel(windows.AddEditLevel{
			Supervisor: u.Supervisor,
			Engine:     d.Engine,
			EditLevel:  scene.Level,

			OnChangePageTypeAndWallpaper: func(pageType level.PageType, wallpaper string) {
				log.Info("OnChangePageTypeAndWallpaper called: %+v, %+v", pageType, wallpaper)
				scene.Level.PageType = pageType
				scene.Level.Wallpaper = wallpaper
				u.Canvas.Destroy() // clean up old textures
				u.Canvas.LoadLevel(scene.Level)
			},
			OnUpdateScreenshot: func() error {
				return scene.UpdateLevelScreenshot(scene.Level)
			},
			OnReload: func() {
				log.Warn("RELOAD LEVEL")
				scene.Reset()
			},
			OnCancel: func() {
				u.levelSettingsWindow.Close()
			},
		})
		u.ConfigureWindow(d, u.levelSettingsWindow)
	}

	// Doodad Properties
	if u.doodadPropertiesWindow == nil {
		scene, _ := d.Scene.(*EditorScene)

		cfg := &windows.DoodadProperties{
			Supervisor: u.Supervisor,
			Engine:     d.Engine,
			EditDoodad: scene.Doodad,
		}

		// Rebuild the window. TODO: hacky af.
		cfg.OnRefresh = func() {
			u.doodadPropertiesWindow.Close()
			u.doodadPropertiesWindow = nil
			u.SetupPopups(u.d)
			u.doodadPropertiesWindow.Show()
		}

		u.doodadPropertiesWindow = windows.NewDoodadPropertiesWindow(cfg)
		u.ConfigureWindow(d, u.doodadPropertiesWindow)
	}

	// Level FileSystem Viewer.
	if u.filesystemWindow == nil {
		scene, _ := d.Scene.(*EditorScene)

		u.filesystemWindow = windows.NewFileSystemWindow(windows.FileSystem{
			Supervisor: u.Supervisor,
			Engine:     d.Engine,
			Level:      scene.Level,

			OnDelete: func(filename string) bool {
				// Check if it is an embedded doodad.
				if strings.HasPrefix(filename, balance.EmbeddedDoodadsBasePath) {
					// Check if we have the doodad installed locally.
					if _, err := doodads.LoadFile(filepath.Base(filename)); err != nil {
						modal.Alert(
							"Cannot remove %s:\n\n"+
								"This doodad is still in use by the level and does not\n"+
								"exist on your local device, so can not be deleted.",
							filepath.Base(filename),
						).WithTitle("Cannot Remove Custom Doodad")
						return false
					}
				}

				// Can't delete the current wallpaper.
				if filepath.Base(filename) == scene.Level.Wallpaper {
					modal.Alert(
						"This wallpaper is still in use as the level background, so can\n" +
							"not be deleted. Change the wallpaper in the Page Settings window\n" +
							"to one of the defaults and then you may remove this file from the level.",
					).WithTitle("Cannot Remove Current Wallpaper")
					return false
				}

				if ok := scene.Level.DeleteFile(filename); !ok {
					modal.Alert("Failed to remove file from level data!")
					return false
				}

				return true
			},
			OnCancel: func() {
				u.filesystemWindow.Close()
			},
		})
		u.ConfigureWindow(d, u.filesystemWindow)
	}

	// Palette Editor.
	if u.paletteEditor == nil {
		scene, _ := d.Scene.(*EditorScene)

		// Which palette?
		var pal *level.Palette
		if scene.Level != nil {
			pal = scene.Level.Palette
		} else if scene.Doodad != nil {
			pal = scene.Doodad.Palette
		}

		u.paletteEditor = windows.NewPaletteEditor(windows.PaletteEditor{
			Supervisor:  u.Supervisor,
			Engine:      d.Engine,
			IsDoodad:    scene.Doodad != nil,
			EditPalette: pal,

			OnChange: func() {
				// Reload the level.
				if scene.Level != nil {
					log.Warn("RELOAD LEVEL")
					u.Canvas.LoadLevel(scene.Level)
					scene.Level.Chunker.Redraw()
				} else if scene.Doodad != nil {
					log.Warn("RELOAD DOODAD")
					u.Canvas.LoadDoodadToLayer(u.Scene.Doodad, u.Scene.ActiveLayer)
					u.Scene.Doodad.Layers[u.Scene.ActiveLayer].Chunker.Redraw()
				}

				// Flush the palette cache in case swatches got renamed,
				// so it rebuilds the "color by name" map from scratch.
				pal.FlushCaches()

				// Reload the palette frame to reflect the changed data.
				u.Palette.Hide()
				u.Palette = u.SetupPalette(d)
				u.Resized(d)
			},
			OnAddColor: func() {
				// Adding a new color to the palette.
				sw, err := pal.NewSwatch()
				if err != nil {
					modal.Alert("Couldn't add this swatch: %s", err).WithTitle("Limit Reached")
					return
				}

				log.Info("Added new palette color: %+v", sw)

				// Awkward but... reload this very same window.
				u.paletteEditor.Close()
				u.paletteEditor = nil
				u.SetupPopups(d)
				u.paletteEditor.Show()
			},
			OnCancel: func() {
				u.paletteEditor.Close()
			},
		})
		u.ConfigureWindow(d, u.paletteEditor)
	}

	// Layers window (doodad editor)
	if u.layersWindow == nil {
		scene, _ := d.Scene.(*EditorScene)

		u.layersWindow = windows.NewLayerWindow(windows.Layers{
			Supervisor:  u.Supervisor,
			Engine:      d.Engine,
			EditDoodad:  scene.Doodad,
			ActiveLayer: scene.ActiveLayer,

			OnChange: func(self *doodads.Doodad) {
				// Reload the level.
				log.Warn("RELOAD LEVEL")
				u.Canvas.LoadDoodad(u.Scene.Doodad)

				for i := range self.Layers {
					scene.Doodad.Layers[i] = self.Layers[i]
				}

				// Awkward but... reload this very same window.
				// Otherwise, the window doesn't update to show the new
				// layer having been added.
				u.layersWindow.Close()
				u.layersWindow = nil
				u.SetupPopups(d)
				u.layersWindow.Show()
			},
			OnAddLayer: func() {
				layer := scene.Doodad.AddLayer(
					fmt.Sprintf("layer %d", len(scene.Doodad.Layers)),
					nil,
				)
				log.Info("Added new layer: %d %s",
					len(scene.Doodad.Layers), layer.Name)

				// Awkward but... reload this very same window.
				// Otherwise, the window doesn't update to show the new
				// layer having been added.
				u.layersWindow.Close()
				u.layersWindow = nil
				u.SetupPopups(d)
				u.layersWindow.Show()
			},
			OnChangeLayer: func(index int) {
				if index < 0 || index >= len(scene.Doodad.Layers) {
					d.FlashError("OnChangeLayer: layer %d out of range", index)
					return
				}

				log.Info("CHANGE DOODAD LAYER TO %d", index)
				u.Canvas.LoadDoodadToLayer(u.Scene.Doodad, index)
				u.Scene.ActiveLayer = index
			},
			OnCancel: func() {
				u.layersWindow.Close()
			},
		})
		u.ConfigureWindow(d, u.layersWindow)
	}
}
