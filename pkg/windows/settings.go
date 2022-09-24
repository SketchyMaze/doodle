package windows

import (
	"strings"

	"git.kirsle.net/SketchyMaze/doodle/pkg/balance"
	"git.kirsle.net/SketchyMaze/doodle/pkg/gamepad"
	"git.kirsle.net/SketchyMaze/doodle/pkg/log"
	"git.kirsle.net/SketchyMaze/doodle/pkg/native"
	magicform "git.kirsle.net/SketchyMaze/doodle/pkg/uix/magic-form"
	"git.kirsle.net/SketchyMaze/doodle/pkg/usercfg"
	"git.kirsle.net/SketchyMaze/doodle/pkg/userdir"
	"git.kirsle.net/go/render"
	"git.kirsle.net/go/ui"
)

// Settings window.
type Settings struct {
	// Settings passed in by doodle
	Supervisor *ui.Supervisor
	Engine     render.Engine

	// Boolean bindings.
	DebugOverlay       *bool
	DebugCollision     *bool
	HorizontalToolbars *bool
	EnableFeatures     *bool
	CrosshairSize      *int
	CrosshairColor     *render.Color
	HideTouchHints     *bool
	DisableAutosave    *bool
	ControllerStyle    *int

	// Configuration options.
	SceneName string // name of scene which called this window
	ActiveTab string // specify the tab to open
	OnApply   func()
}

// MakeSettingsWindow initializes a settings window for any scene.
// The window width/height are the actual SDL2 window dimensions.
func MakeSettingsWindow(windowWidth, windowHeight int, cfg Settings) *ui.Window {
	win := NewSettingsWindow(cfg)
	win.Compute(cfg.Engine)
	win.Supervise(cfg.Supervisor)

	// Center the window.
	size := win.Size()
	win.MoveTo(render.Point{
		X: (windowWidth / 2) - (size.W / 2),
		Y: (windowHeight / 2) - (size.H / 2),
	})

	return win
}

// NewSettingsWindow initializes the window.
func NewSettingsWindow(cfg Settings) *ui.Window {
	var (
		Width  = 400
		Height = 360
	)

	window := ui.NewWindow("Settings")
	window.SetButtons(ui.CloseButton)
	window.Configure(ui.Config{
		Width:      Width,
		Height:     Height,
		Background: render.Grey,
	})

	///////////
	// Tab Bar
	tabFrame := ui.NewTabFrame("Tab Frame")
	tabFrame.SetBackground(render.DarkGrey)
	window.Pack(tabFrame, ui.Pack{
		Side:  ui.N,
		FillX: true,
	})

	// Make the tabs
	cfg.makeOptionsTab(tabFrame, Width, Height)
	cfg.makeControlsTab(tabFrame, Width, Height)
	cfg.makeControllerTab(tabFrame, Width, Height)
	cfg.makeExperimentalTab(tabFrame, Width, Height)

	tabFrame.Supervise(cfg.Supervisor)

	return window
}

// saveGameSettings controls pkg/usercfg to write the user settings
// to disk, based on the settings toggle-able from the UI window.
func saveGameSettings() {
	log.Info("Saving game settings")
	if err := usercfg.Save(); err != nil {
		log.Error("Couldn't save game settings: %s", err)
	}
}

// Settings Window "Options" Tab
func (c Settings) makeOptionsTab(tabFrame *ui.TabFrame, Width, Height int) *ui.Frame {
	tab := tabFrame.AddTab("Options", ui.NewLabel(ui.Label{
		Text: "Options",
		Font: balance.TabFont,
	}))
	tab.Resize(render.NewRect(Width-4, Height-tab.Size().H-46))

	// Common click handler for all settings,
	// so we can write the updated info to disk.
	onClick := func() {
		saveGameSettings()
	}

	// The CrosshairSize is ideally a 0-100 (percent) how big the editor
	// crosshair is, but options now are only 0% or 100% so it presents
	// this as a checkbox for now.
	var crosshairEnabled = *c.CrosshairSize > 0

	form := magicform.Form{
		Supervisor: c.Supervisor,
		Engine:     c.Engine,
		Vertical:   true,
		LabelWidth: 150,
	}
	form.Create(tab, []magicform.Field{
		{
			Label: "Game Options",
			Font:  balance.LabelFont,
		},
		{
			Label:        "Hide touchscreen control hints during Play Mode",
			Font:         balance.UIFont,
			BoolVariable: c.HideTouchHints,
			OnClick:      onClick,
		},
		{
			Label: "Level & Doodad Editor",
			Font:  balance.LabelFont,
		},
		{
			Label:        "Horizontal instead of vertical toolbars",
			Font:         balance.UIFont,
			BoolVariable: c.HorizontalToolbars,
			OnClick:      onClick,
			Tooltip: ui.Tooltip{
				Text: "Note: reload your level after changing this option.\n" +
					"Playtesting and returning will do.",
				Edge: ui.Top,
			},
		},
		{
			Label:        "Disable auto-save in the Editor",
			Font:         balance.UIFont,
			BoolVariable: c.DisableAutosave,
			OnClick:      onClick,
		},
		{
			Label:        "Draw a crosshair at the mouse cursor.",
			Font:         balance.UIFont,
			BoolVariable: &crosshairEnabled,
			OnClick: func() {
				if crosshairEnabled {
					*c.CrosshairSize = 100
				} else {
					*c.CrosshairSize = 0
				}
				onClick()
			},
		},
		{
			Type:  magicform.Color,
			Label: "Crosshair color:",
			Font:  balance.UIFont,
			Color: c.CrosshairColor,
			OnClick: func() {
				onClick()
			},
		},
		{
			Label: "My Custom Content",
			Font:  balance.LabelFont,
		},
		{
			Label: "Levels and doodads you create in-game are placed in your\n" +
				"Profile Directory, which you can access below:",
			Font: balance.UIFont,
		},
		{
			Buttons: []magicform.Field{
				{
					Label:       "Open profile directory",
					Font:        balance.UIFont,
					ButtonStyle: &balance.ButtonPrimary,
					OnClick: func() {
						path := strings.ReplaceAll(userdir.ProfileDirectory, "\\", "/")
						if path[0] != '/' {
							path = "/" + path
						}
						native.OpenURL("file://" + path)
					},
				},
			},
		},
	})

	return tab
}

// Settings Window "Controls" Tab
func (c Settings) makeControlsTab(tabFrame *ui.TabFrame, Width, Height int) *ui.Frame {
	frame := tabFrame.AddTab("Controls", ui.NewLabel(ui.Label{
		Text: "Controls",
		Font: balance.TabFont,
	}))
	frame.Resize(render.NewRect(Width-4, Height-frame.Size().H-46))

	var (
		halfWidth        = (Width - 4) / 2 // the 4 is for window borders, TODO
		shortcutTabWidth = float64(halfWidth) * 0.5
		infoTabWidth     = float64(halfWidth) * 0.5
		rowHeight        = 20

		shortcutTabSize = render.NewRect(int(shortcutTabWidth), rowHeight)
		infoTabSize     = render.NewRect(int(infoTabWidth), rowHeight)
	)

	controls := []struct {
		Header   string
		Label    string
		Shortcut string
	}{
		{
			Header: "Universal Shortcut Keys",
		},
		{
			Shortcut: "Escape",
			Label:    "Exit game",
		},
		{
			Shortcut: "F1",
			Label:    "Guidebook",
		},
		{
			Shortcut: "`",
			Label:    "Dev console",
		},
		{
			Header: "Gameplay Controls (Play Mode)",
		},
		{
			Shortcut: "Up or W",
			Label:    "Jump",
		},
		{
			Shortcut: "Space",
			Label:    "Activate",
		},
		{
			Shortcut: "Left or A",
			Label:    "Move left",
		},
		{
			Shortcut: "Right or D",
			Label:    "Move right",
		},
		{
			Header: "Level Editor Shortcuts",
		},
		{
			Shortcut: "Ctrl-N",
			Label:    "New level",
		},
		{
			Shortcut: "Ctrl-O",
			Label:    "Open drawing",
		},
		{
			Shortcut: "Ctrl-S",
			Label:    "Save drawing",
		},
		{
			Shortcut: "Shift-Ctrl-S",
			Label:    "Save a copy",
		},
		{
			Shortcut: "Ctrl-Z",
			Label:    "Undo stroke",
		},
		{
			Shortcut: "Ctrl-Y",
			Label:    "Redo stroke",
		},
		{
			Shortcut: "P",
			Label:    "Playtest",
		},
		{
			Shortcut: "0",
			Label:    "Scroll to origin",
		},
		{
			Shortcut: "q",
			Label:    "Doodads",
		},
		{
			Shortcut: "f",
			Label:    "Pencil Tool",
		},
		{
			Shortcut: "l",
			Label:    "Line Tool",
		},
		{
			Shortcut: "r",
			Label:    "Rectangle Tool",
		},
		{
			Shortcut: "c",
			Label:    "Ellipse Tool",
		},
		{
			Shortcut: "x",
			Label:    "Eraser Tool",
		},
	}
	var curFrame = ui.NewFrame("Frame")
	frame.Pack(curFrame, ui.Pack{
		Side:  ui.N,
		FillX: true,
	})
	var i = -1 // manually controlled
	for _, row := range controls {
		i++
		row := row

		if row.Header != "" {
			// Close out a previous Frame?
			if i != 0 {
				curFrame = ui.NewFrame("Header Row")
				frame.Pack(curFrame, ui.Pack{
					Side:  ui.N,
					FillX: true,
				})
			}

			label := ui.NewLabel(ui.Label{
				Text: row.Header,
				Font: balance.LabelFont,
			})
			curFrame.Pack(label, ui.Pack{
				Side: ui.W,
			})

			// Set up the next series of shortcut keys.
			i = -1
			curFrame = ui.NewFrame("Frame")
			frame.Pack(curFrame, ui.Pack{
				Side:  ui.N,
				FillX: true,
			})
			continue
		}

		// Cut a new frame every 2 items.
		if i > 0 && i%2 == 0 {
			curFrame = ui.NewFrame("Frame")
			frame.Pack(curFrame, ui.Pack{
				Side:  ui.N,
				FillX: true,
				PadY:  1,
			})
		}

		keyLabel := ui.NewLabel(ui.Label{
			Text: row.Shortcut,
			Font: balance.CodeLiteralFont,
		})
		keyLabel.Configure(ui.Config{
			Background:  render.RGBA(255, 255, 220, 255),
			BorderSize:  1,
			BorderStyle: ui.BorderSunken,
			BorderColor: render.DarkGrey,
		})
		keyLabel.Resize(shortcutTabSize)
		curFrame.Pack(keyLabel, ui.Pack{
			Side: ui.W,
			PadX: 1,
		})

		helpLabel := ui.NewLabel(ui.Label{
			Text: row.Label,
			Font: balance.UIFont,
		})
		helpLabel.Resize(infoTabSize)
		curFrame.Pack(helpLabel, ui.Pack{
			Side: ui.W,
			PadX: 1,
		})
	}

	return frame
}

// Settings Window "Experimental" Tab
func (c Settings) makeExperimentalTab(tabFrame *ui.TabFrame, Width, Height int) *ui.Frame {
	tab := tabFrame.AddTab("Experimental", ui.NewLabel(ui.Label{
		Text: "Experimental",
		Font: balance.TabFont,
	}))
	tab.Resize(render.NewRect(Width-4, Height-tab.Size().H-46))

	// Common click handler for all settings,
	// so we can write the updated info to disk.
	onClick := func() {
		saveGameSettings()
	}

	form := magicform.Form{
		Supervisor: c.Supervisor,
		Engine:     c.Engine,
		Vertical:   true,
		LabelWidth: 150,
	}
	form.Create(tab, []magicform.Field{
		{
			Label: "Enable Experimental Features",
			Font:  balance.LabelFont,
		},
		{
			Label: "The setting below can enable experimental features in this\n" +
				"game. These are features which are still in development and\n" +
				"may have unstable or buggy behavior.",
			Font: balance.UIFont,
		},
		{
			Label: "Viewport window",
			Font:  balance.LabelFont,
		},
		{
			Label: "This option in the Level menu opens another view into\n" +
				"the level. Has glitchy wallpaper problems.",
			Font: balance.UIFont,
		},
		{
			BoolVariable: c.EnableFeatures,
			Label:        "Enable experimental features",
			Font:         balance.UIFont,
			OnClick:      onClick,
		},
		{
			Label: "Restart the game for changes to take effect.",
			Font:  balance.UIFont,
		},
	})

	return tab
}

// Settings Window "Controller" Tab
func (c Settings) makeControllerTab(tabFrame *ui.TabFrame, Width, Height int) *ui.Frame {
	tab := tabFrame.AddTab("Gamepad", ui.NewLabel(ui.Label{
		Text: "Gamepad",
		Font: balance.TabFont,
	}))
	tab.Resize(render.NewRect(Width-4, Height-tab.Size().H-46))

	// Render the form.
	form := magicform.Form{
		Supervisor: c.Supervisor,
		Engine:     c.Engine,
		Vertical:   true,
		LabelWidth: 150,
	}
	form.Create(tab, []magicform.Field{
		{
			Label: "Play with an Xbox or Nintendo controller!",
			Font:  balance.LabelFont,
		},
		{
			Label: "If you have a Nintendo-style controller (your A button is on\n" +
				"the right and B button on bottom), pick 'N Style' to reverse\n" +
				"the A/B and X/Y buttons.",
			Font: balance.UIFont,
		},
		{
			Label: "Button Style:",
			Font:  balance.LabelFont,
			Type:  magicform.Selectbox,
			Options: []magicform.Option{
				{
					Label: "X Style (default)",
					Value: int(gamepad.XStyle),
				},
				{
					Label: "N Style",
					Value: int(gamepad.NStyle),
				},
			},
			SelectValue: &c.ControllerStyle,
			OnSelect: func(v interface{}) {
				style, _ := v.(int)
				log.Error("style: %d", style)
				gamepad.SetStyle(gamepad.Style(style))
				*c.ControllerStyle = style
				saveGameSettings()
			},
		},
		{
			Label: "The gamepad controls vary between two modes:",
			Font:  balance.UIFont,
		},
		{
			Label: "Mouse Mode (outside of gameplay)",
			Font:  balance.LabelFont,
		},
		{
			Label: "The left analog stick moves a mouse cursor around.\n" +
				"The right analog stick scrolls the level around.\n" +
				"A or X: Left-click    B or Y: Right-click\n" +
				"L1: Middle-click  L2: Close window",
			Font: balance.UIFont,
		},
		{
			Label: "Gameplay Mode",
			Font:  balance.LabelFont,
		},
		{
			Label: "Left stick or D-Pad to move the player around.\n" +
				"A or X: 'Use'    B or Y: 'Jump'\n" +
				"R1: Toggle between Mouse and Gameplay controls.",
			Font: balance.UIFont,
		},
	})

	return tab
}
