package doodle

import (
	"git.kirsle.net/SketchyMaze/doodle/pkg/balance"
	"git.kirsle.net/SketchyMaze/doodle/pkg/branding"
	"git.kirsle.net/SketchyMaze/doodle/pkg/native"
	"git.kirsle.net/SketchyMaze/doodle/pkg/windows"
	"git.kirsle.net/go/render"
	"git.kirsle.net/go/ui"
)

// Common menubars between Play and Edit.

// MakeHelpMenu creates the "Help" menu with its common items
// across any scene.
func (d *Doodle) MakeHelpMenu(menu *ui.MenuBar, supervisor *ui.Supervisor) *ui.MenuButton {
	helpMenu := menu.AddMenu("Help")
	helpMenu.AddItemAccel("User Manual", "F1", func() {
		native.OpenLocalURL(balance.GuidebookPath)
	})
	helpMenu.AddItem("About", func() {
		aboutWindow := windows.NewAboutWindow(windows.About{
			Supervisor: supervisor,
			Engine:     d.Engine,
		})
		aboutWindow.Compute(d.Engine)
		aboutWindow.Supervise(supervisor)

		// Center the window.
		aboutWindow.MoveTo(render.Point{
			X: (d.width / 2) - (aboutWindow.Size().W / 2),
			Y: 60,
		})
		aboutWindow.Show()
	})
	helpMenu.AddSeparator()
	helpMenu.AddItem("Go to Website", func() {
		native.OpenURL(branding.Website)
	})
	helpMenu.AddItem("Guidebook Online", func() {
		native.OpenURL(branding.GuidebookURL)
	})
	return helpMenu
}
