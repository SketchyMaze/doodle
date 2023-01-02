package windows

import (
	"git.kirsle.net/SketchyMaze/doodle/pkg/balance"
	"git.kirsle.net/SketchyMaze/doodle/pkg/shmem"
	magicform "git.kirsle.net/SketchyMaze/doodle/pkg/uix/magic-form"
	"git.kirsle.net/SketchyMaze/doodle/pkg/usercfg"
	"git.kirsle.net/go/render"
	"git.kirsle.net/go/ui"
)

// CheatsMenu window.
type CheatsMenu struct {
	// Settings passed in by doodle
	Supervisor *ui.Supervisor
	Engine     render.Engine

	// SceneName: the caller will provide a fresh SceneName since
	// the cheats window could span multiple scenes.
	SceneName func() string

	// Window wants to run a developer shell command (e.g. cheat codes).
	RunCommand           func(string)
	OnSetPlayerCharacter func(string)
}

// MakeCheatsMenu initializes a settings window for any scene.
// The window width/height are the actual SDL2 window dimensions.
func MakeCheatsMenu(cfg CheatsMenu) *ui.Window {
	var (
	// Application window width/height to center our window
	// _, h = shmem.CurrentRenderEngine.WindowSize()
	)

	win := NewCheatsWindow(cfg)
	win.Compute(cfg.Engine)
	win.Supervise(cfg.Supervisor)

	// Center the window.
	// size := win.Size()
	win.MoveTo(render.Point{
		X: 20,
		Y: 40,
	})

	return win
}

// NewCheatsWindow initializes the window.
func NewCheatsWindow(cfg CheatsMenu) *ui.Window {
	var (
		Width  = 200
		Height = 300
	)

	window := ui.NewWindow("Cheats Menu")
	window.SetButtons(ui.CloseButton)
	window.Configure(ui.Config{
		Width:      Width,
		Height:     Height,
		Background: balance.CheatsMenuBackground,
	})

	///////////
	// Tab Bar
	tabFrame := ui.NewTabFrame("Tab Frame")
	tabFrame.SetBackground(balance.CheatsMenuBackground)
	window.Pack(tabFrame, ui.Pack{
		Side:  ui.N,
		FillX: true,
	})

	// Make the tabs
	cfg.makePlayModeTab(tabFrame, Width, Height)
	cfg.makeMiscTab(tabFrame, Width, Height)

	tabFrame.Supervise(cfg.Supervisor)

	return window
}

// Cheats Menu "Play Mode" Tab
func (c CheatsMenu) makePlayModeTab(tabFrame *ui.TabFrame, Width, Height int) *ui.Frame {
	tab := tabFrame.AddTab("Gameplay", ui.NewLabel(ui.Label{
		Text: "Gameplay",
		Font: balance.TabFont,
	}))
	tab.Resize(render.NewRect(Width-4, Height-tab.Size().H-46))

	// Run a command on the developer shell.
	run := func(command string) {
		if c.RunCommand != nil {
			c.RunCommand(command)
		} else {
			shmem.FlashError("CheatsMenu: RunCommand() handler not available")
		}
	}

	// Dummy variable for the "play as" dropdown.
	var playAs string

	form := magicform.Form{
		Supervisor: c.Supervisor,
		Engine:     c.Engine,
		Vertical:   true,
		LabelWidth: 90,
		PadY:       0,
		PadX:       0,
	}
	form.Create(tab, []magicform.Field{
		{
			Label: "These cheats are available\n" +
				"only during level gameplay.",
			Font: balance.UIFont,
		},
		{
			Label:        "Play as:",
			TextVariable: &playAs,
			Options:      balance.CheatMenuActors,
			Font:         balance.UIFont,
			OnSelect: func(v interface{}) {
				doodad := v.(string)
				if c.OnSetPlayerCharacter != nil {
					c.OnSetPlayerCharacter(doodad)
				} else {
					shmem.FlashError("OnSetPlayerCharacter(%s): handler not ready", doodad)
				}
			},
		},
		{
			Buttons: []magicform.Field{
				{
					Label:       "God Mode",
					Font:        balance.SmallFont,
					ButtonStyle: &balance.ButtonDanger,
					Tooltip: ui.Tooltip{
						Text: "Makes you invulnerable to damage and fire.",
					},
					OnClick: func() {
						run(balance.CheatGodMode)
					},
				},
				{
					Label: "Show hidden actors",
					Font:  balance.SmallFont,
					OnClick: func() {
						run(balance.CheatShowAllActors)
					},
				},
			},
		},
		{
			Label: "Inventory",
			Font:  balance.LabelFont,
		},
		{
			Buttons: []magicform.Field{
				{
					Label: "Give Keys",
					Font:  balance.SmallFont,
					Tooltip: ui.Tooltip{
						Text: "Get all four colored keys and\n99x small keys",
					},
					OnClick: func() {
						run(balance.CheatGiveKeys)
					},
				},
				{
					Label: "Give Gems",
					Font:  balance.SmallFont,
					Tooltip: ui.Tooltip{
						Text: "Get 1x of each of the four Gemstones.",
					},
					OnClick: func() {
						run(balance.CheatGiveGems)
					},
				},
				{
					Label:       "Drop All",
					Font:        balance.SmallFont,
					ButtonStyle: &balance.ButtonDanger,
					Tooltip: ui.Tooltip{
						Text: "Remove ALL items from your inventory.",
					},
					OnClick: func() {
						run(balance.CheatDropItems)
					},
				},
			},
		},
		{
			Label: "Physics",
			Font:  balance.LabelFont,
		},
		{
			Buttons: []magicform.Field{
				{
					Label: "Antigravity",
					Font:  balance.SmallFont,
					Tooltip: ui.Tooltip{
						Text: "Allows free movement in four directions",
					},
					OnClick: func() {
						run(balance.CheatAntigravity)
					},
				},
				{
					Label: "NoClip",
					Font:  balance.SmallFont,
					Tooltip: ui.Tooltip{
						Text: "Toggle physical collision\n" +
							"checks with level and actors.",
					},
					OnClick: func() {
						run(balance.CheatNoclip)
					},
				},
			},
		},
		{
			Buttons: []magicform.Field{
				{
					Label:       "Skip this level",
					Font:        balance.SmallFont,
					ButtonStyle: &balance.ButtonDanger,
					Tooltip: ui.Tooltip{
						Text: "Consider the current level a win.",
					},
					OnClick: func() {
						run(balance.CheatSkipLevel)
					},
				},
			},
		},
	})

	return tab
}

// Cheats Menu "Misc" Tab
func (c CheatsMenu) makeMiscTab(tabFrame *ui.TabFrame, Width, Height int) *ui.Frame {
	tab := tabFrame.AddTab("Misc", ui.NewLabel(ui.Label{
		Text: "Misc",
		Font: balance.TabFont,
	}))
	tab.Resize(render.NewRect(Width-4, Height-tab.Size().H-46))

	// Run a command on the developer shell.
	run := func(command string) {
		if c.RunCommand != nil {
			c.RunCommand(command)
		} else {
			shmem.FlashError("CheatsMenu: RunCommand() handler not available")
		}
	}

	form := magicform.Form{
		Supervisor: c.Supervisor,
		Engine:     c.Engine,
		Vertical:   true,
		LabelWidth: 90,
		PadY:       0,
		PadX:       0,
	}
	form.Create(tab, []magicform.Field{
		{
			Label:        "Enable cheats menu",
			BoolVariable: &usercfg.Current.EnableCheatsMenu,
			Tooltip: ui.Tooltip{
				Text: "Enables a Help->Cheats Menu during gameplay.",
			},
			OnClick: func() {
				saveGameSettings()
			},
		},
		{
			Label: "Level Editor",
			Font:  balance.LabelFont,
		},
		{
			Buttons: []magicform.Field{
				{
					Label: "Show hidden doodads",
					Font:  balance.SmallFont,
					Tooltip: ui.Tooltip{
						Text: "Enable hidden built-in doodads (such as Boy)\n" +
							"to be used in the Level Editor.",
						Edge: ui.Bottom,
					},
					OnClick: func() {
						// Like `boolProp show-hidden-doodads true`
						var bp = "show-hidden-doodads"
						if v, err := balance.GetBoolProp(bp); err == nil {
							v = !v
							balance.BoolProp(bp, v)
							if v {
								shmem.Flash("Hidden doodads will appear when you next reload the level editor.")
							} else {
								shmem.Flash("Hidden doodads are again hidden from the level editor.")
							}
						}
					},
				},
			},
		},
		{
			Label: "Testing",
			Font:  balance.LabelFont,
		},
		{
			Buttons: []magicform.Field{
				{
					Label: "Load Screen",
					Font:  balance.SmallFont,
					OnClick: func() {
						run(balance.CheatDebugLoadScreen)
					},
				},
				{
					Label: "Wait Screen",
					Font:  balance.SmallFont,
					OnClick: func() {
						run(balance.CheatDebugWaitScreen)
					},
				},
			},
		},
		{
			Label: "Level Progression",
			Font:  balance.LabelFont,
		},
		{
			Buttons: []magicform.Field{
				{
					Label: "Unlock all levels",
					Font:  balance.SmallFont,
					Tooltip: ui.Tooltip{
						Text: "For this play session, any level may be opened\n" +
							"from Story Mode regardless of the padlock icon.",
					},
					OnClick: func() {
						run(balance.CheatUnlockLevels)
					},
				},
			},
		},
		{
			Label: "Debugging",
			Font:  balance.LabelFont,
		},
		{
			Buttons: []magicform.Field{
				{
					Label: "Debug overlay",
					Font:  balance.SmallFont,
					OnClick: func() {
						run("boolprop DO flip")
					},
				},
				{
					Label: "Show hitboxes",
					Font:  balance.SmallFont,
					OnClick: func() {
						run("boolprop DC flip")
					},
				},
			},
		},
	})

	return tab
}
