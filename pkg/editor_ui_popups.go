package doodle

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"git.kirsle.net/apps/doodle/pkg/doodads"
	"git.kirsle.net/apps/doodle/pkg/level"
	"git.kirsle.net/apps/doodle/pkg/level/publishing"
	"git.kirsle.net/apps/doodle/pkg/log"
	"git.kirsle.net/apps/doodle/pkg/modal"
	"git.kirsle.net/apps/doodle/pkg/windows"
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
	u.layersWindow.Hide()
	u.layersWindow = nil
	u.SetupPopups(u.d)
	u.layersWindow.Show()
}

// OpenPaletteWindow opens the Palette Editor window.
func (u *EditorUI) OpenPaletteWindow() {
	// TODO: recompute the window so the actual loaded level palette gets in
	u.paletteEditor.Hide()
	u.paletteEditor = nil
	u.SetupPopups(u.d)
	u.paletteEditor.Show()
}

// OpenDoodadDropper opens the Doodad Dropper window.
func (u *EditorUI) OpenDoodadDropper() {
	// NOTE: most places in the code call this directly, nice
	// and simple window :) but OpenDoodadDropper() added for consistency.
	u.doodadWindow.Show()
}

// OpenPublishWindow opens the Publisher window.
func (u *EditorUI) OpenPublishWindow() {
	u.publishWindow.Hide()
	u.publishWindow = nil
	u.SetupPopups(u.d)
	u.publishWindow.Show()
}

// SetupPopups preloads popup windows like the DoodadDropper.
func (u *EditorUI) SetupPopups(d *Doodle) {
	// Common window configure function.
	var configure = func(window *ui.Window) {
		var size = window.Size()
		window.Compute(d.Engine)
		window.Supervise(u.Supervisor)

		// Center the window.
		window.MoveTo(render.Point{
			X: (d.width / 2) - (size.W / 2),
			Y: (d.height / 2) - (size.H / 2),
		})

		window.Hide()
	}

	// Doodad Dropper.
	if u.doodadWindow == nil {
		u.doodadWindow = windows.NewDoodadDropper(windows.DoodadDropper{
			Supervisor: u.Supervisor,
			Engine:     d.Engine,

			OnStartDragActor: u.startDragActor,
			OnCancel: func() {
				u.doodadWindow.Hide()
			},
		})
		configure(u.doodadWindow)
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
				u.Canvas.LoadLevel(d.Engine, scene.Level)
			},
			OnCancel: func() {
				u.levelSettingsWindow.Hide()
			},
		})
		configure(u.levelSettingsWindow)
	}

	// Publish Level (embed doodads)
	if u.publishWindow == nil {
		scene, _ := d.Scene.(*EditorScene)

		u.publishWindow = windows.NewPublishWindow(windows.Publish{
			Supervisor: u.Supervisor,
			Engine:     d.Engine,
			Level:      scene.Level,

			OnPublish: func(includeBuiltins bool) {
				log.Debug("OnPublish: include builtins=%+v", includeBuiltins)
				cwd, _ := os.Getwd()
				d.Prompt(fmt.Sprintf("File name (relative to %s)> ", cwd), func(answer string) {
					if answer == "" {
						d.Flash("A file name is required to publish this level.")
						return
					}

					if !strings.HasSuffix(answer, ".level") {
						answer += ".level"
					}

					answer = filepath.Join(cwd, answer)
					log.Debug("call with includeBuiltins=%+v", includeBuiltins)
					if _, err := publishing.Publish(scene.Level, answer, includeBuiltins); err != nil {
						modal.Alert("Error when publishing the level: %s", err)
						return
					}
					d.Flash("Exported published level to: %s", answer)
				})
			},
			OnCancel: func() {
				u.publishWindow.Hide()
			},
		})
		configure(u.publishWindow)
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
					u.Canvas.LoadLevel(d.Engine, scene.Level)
					scene.Level.Chunker.Redraw()
				} else if scene.Doodad != nil {
					log.Warn("RELOAD DOODAD")
					u.Canvas.LoadDoodadToLayer(u.Scene.Doodad, u.Scene.ActiveLayer)
					u.Scene.Doodad.Layers[u.Scene.ActiveLayer].Chunker.Redraw()
				}

				// Reload the palette frame to reflect the changed data.
				u.Palette.Hide()
				u.Palette = u.SetupPalette(d)
				u.Resized(d)
			},
			OnAddColor: func() {
				// Adding a new color to the palette.
				sw := pal.AddSwatch()
				log.Info("Added new palette color: %+v", sw)

				// Awkward but... reload this very same window.
				u.paletteEditor.Hide()
				u.paletteEditor = nil
				u.SetupPopups(d)
				u.paletteEditor.Show()
			},
			OnCancel: func() {
				u.paletteEditor.Hide()
			},
		})
		configure(u.paletteEditor)
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
				u.layersWindow.Hide()
				u.layersWindow = nil
				u.SetupPopups(d)
				u.layersWindow.Show()
			},
			OnAddLayer: func() {
				layer := doodads.Layer{
					Name:    fmt.Sprintf("layer %d", len(scene.Doodad.Layers)),
					Chunker: level.NewChunker(scene.DoodadSize),
				}
				scene.Doodad.Layers = append(scene.Doodad.Layers, layer)
				log.Info("Added new layer: %d %s",
					len(scene.Doodad.Layers), layer.Name)

				// Awkward but... reload this very same window.
				// Otherwise, the window doesn't update to show the new
				// layer having been added.
				u.layersWindow.Hide()
				u.layersWindow = nil
				u.SetupPopups(d)
				u.layersWindow.Show()
			},
			OnChangeLayer: func(index int) {
				if index < 0 || index >= len(scene.Doodad.Layers) {
					d.Flash("OnChangeLayer: layer %d out of range", index)
					return
				}

				log.Info("CHANGE DOODAD LAYER TO %d", index)
				u.Canvas.LoadDoodadToLayer(u.Scene.Doodad, index)
				u.Scene.ActiveLayer = index
			},
			OnCancel: func() {
				u.layersWindow.Hide()
			},
		})
		configure(u.layersWindow)
	}
}
