package doodle

import (
	"git.kirsle.net/SketchyMaze/doodle/pkg/levelpack"
	"git.kirsle.net/SketchyMaze/doodle/pkg/shmem"
	"git.kirsle.net/SketchyMaze/doodle/pkg/usercfg"
	"git.kirsle.net/SketchyMaze/doodle/pkg/windows"
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
				Supervisor: u.Supervisor,
				Engine:     d.Engine,

				OnPlayLevel: func(lp *levelpack.LevelPack, which levelpack.Level) {
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
	gameMenu.AddItemAccel("Open drawing", "Ctrl-O", func() {
		if u.winOpenLevel == nil {
			u.winOpenLevel = windows.NewOpenDrawingWindow(windows.OpenDrawing{
				Supervisor: u.Supervisor,
				Engine:     shmem.CurrentRenderEngine,
				OnOpenDrawing: func(filename string) {
					d.EditFile(filename)
				},
				OnCloseWindow: func() {
					u.winOpenLevel.Destroy()
					u.winOpenLevel = nil
				},
			})
		}
		u.winOpenLevel.MoveTo(render.Point{
			X: (d.width / 2) - (u.winOpenLevel.Size().W / 2),
			Y: (d.height / 2) - (u.winOpenLevel.Size().H / 2),
		})
		u.winOpenLevel.Show()
	})

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
	levelMenu.AddItem("Restart level", u.RestartLevel)
	levelMenu.AddItem("Retry from checkpoint", func() {
		u.SetImperfect()
		u.RetryCheckpoint()
	})
	levelMenu.AddSeparator()
	levelMenu.AddItemAccel("Edit level", "E", u.EditLevel)

	// Hilariously broken, someday!
	if usercfg.Current.EnableFeatures {
		levelMenu.AddSeparator()
		levelMenu.AddItemAccel("New viewport", "v", func() {
			pip := windows.MakePiPWindow(d.width, d.height, windows.PiP{
				Supervisor: u.Supervisor,
				Engine:     u.d.Engine,
				Level:      u.Level,
				Event:      u.d.event,
			})

			pip.Show()
		})
	}

	helpMenu := d.MakeHelpMenu(menu, u.Supervisor)
	if usercfg.Current.EnableCheatsMenu {
		helpMenu.AddSeparator()
		helpMenu.AddItem("Cheats Menu", func() {
			if u.cheatsWindow != nil {
				u.cheatsWindow.Hide()
				u.cheatsWindow.Destroy()
				u.cheatsWindow = nil
			}

			u.cheatsWindow = u.d.MakeCheatsWindow(u.Supervisor)
			u.cheatsWindow.Show()
		})
	}

	menu.Supervise(u.Supervisor)
	menu.Compute(d.Engine)

	return menu
}
