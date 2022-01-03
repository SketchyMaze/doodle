package doodle

import (
	"git.kirsle.net/apps/doodle/pkg/levelpack"
	"git.kirsle.net/apps/doodle/pkg/shmem"
	"git.kirsle.net/apps/doodle/pkg/usercfg"
	"git.kirsle.net/apps/doodle/pkg/windows"
	"git.kirsle.net/go/render"
	"git.kirsle.net/go/ui"
)

// Set up the menu bar for Play Scene.
func (u *PlayScene) setupMenuBar(d *Doodle) *ui.MenuBar {
	menu := ui.NewMenuBar("Main Menu")

	////////
	// Game menu
	gameMenu := menu.AddMenu("Game")
	gameMenu.AddItem("Story Mode", func() {
		// TODO: de-duplicate code from MainScene
		if u.winLevelPacks == nil {
			u.winLevelPacks = windows.NewLevelPackWindow(windows.LevelPack{
				Supervisor: u.supervisor,
				Engine:     d.Engine,

				OnPlayLevel: func(lp levelpack.LevelPack, which levelpack.Level) {
					if err := d.PlayFromLevelpack(lp, which); err != nil {
						shmem.FlashError(err.Error())
					}
				},
				OnCloseWindow: func() {
					u.winLevelPacks.Hide()
				},
			})
		}
		u.winLevelPacks.MoveTo(render.Point{
			X: (d.width / 2) - (u.winLevelPacks.Size().W / 2),
			Y: (d.height / 2) - (u.winLevelPacks.Size().H / 2),
		})
		u.winLevelPacks.Show()
	})
	gameMenu.AddItemAccel("New drawing", "Ctrl-N", d.GotoNewMenu)
	gameMenu.AddItemAccel("Open drawing", "Ctrl-O", d.GotoLoadMenu)

	gameMenu.AddSeparator()
	gameMenu.AddItem("Quit to menu", func() {
		d.Goto(&MainScene{})
	})
	gameMenu.AddItemAccel("Quit", "Escape", func() {
		d.ConfirmExit()
	})

	////////
	// Level menu
	levelMenu := menu.AddMenu("Level")
	levelMenu.AddItemAccel("Edit level", "E", u.EditLevel)

	// Hilariously broken, someday!
	if usercfg.Current.EnableFeatures {
		levelMenu.AddSeparator()
		levelMenu.AddItemAccel("New viewport", "v", func() {
			pip := windows.MakePiPWindow(d.width, d.height, windows.PiP{
				Supervisor: u.supervisor,
				Engine:     u.d.Engine,
				Level:      u.Level,
				Event:      u.d.event,
			})

			pip.Show()
		})
	}

	d.MakeHelpMenu(menu, u.supervisor)

	menu.Supervise(u.supervisor)
	menu.Compute(d.Engine)

	return menu
}
