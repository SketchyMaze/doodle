package windows

import (
	"strings"

	"git.kirsle.net/apps/doodle/pkg/balance"
	"git.kirsle.net/apps/doodle/pkg/log"
	"git.kirsle.net/apps/doodle/pkg/native"
	"git.kirsle.net/apps/doodle/pkg/usercfg"
	"git.kirsle.net/apps/doodle/pkg/userdir"
	"git.kirsle.net/go/render"
	"git.kirsle.net/go/ui"
	"git.kirsle.net/go/ui/style"
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
		Width     = 400
		Height    = 400
		ActiveTab = "index"

		// The tab frames
		TabOptions  *ui.Frame // index
		TabControls *ui.Frame // controls
	)
	if cfg.ActiveTab != "" {
		ActiveTab = cfg.ActiveTab
	}

	window := ui.NewWindow("Settings")
	window.SetButtons(ui.CloseButton)
	window.Configure(ui.Config{
		Width:      Width,
		Height:     Height,
		Background: render.Grey,
	})

	///////////
	// Tab Bar
	tabFrame := ui.NewFrame("Tab Bar")
	tabFrame.SetBackground(render.DarkGrey)
	window.Pack(tabFrame, ui.Pack{
		Side:  ui.N,
		FillX: true,
	})
	for _, tab := range []struct {
		label string
		value string
	}{
		{"Options", "index"},
		{"Controls", "controls"},
	} {
		radio := ui.NewRadioButton(tab.label, &ActiveTab, tab.value, ui.NewLabel(ui.Label{
			Text: tab.label,
			Font: balance.UIFont,
		}))
		radio.SetStyle(&balance.ButtonBabyBlue)
		radio.Handle(ui.Click, func(ed ui.EventData) error {
			switch ActiveTab {
			case "index":
				TabOptions.Show()
				TabControls.Hide()
			case "controls":
				TabOptions.Hide()
				TabControls.Show()
			}
			return nil
		})
		cfg.Supervisor.Add(radio)
		tabFrame.Pack(radio, ui.Pack{
			Side:   ui.W,
			Expand: true,
		})
	}

	///////////
	// Options (index) Tab
	TabOptions = cfg.makeOptionsTab(Width, Height)
	if ActiveTab != "index" {
		TabOptions.Hide()
	}
	window.Pack(TabOptions, ui.Pack{
		Side:  ui.N,
		FillX: true,
	})

	///////////
	// Controls Tab
	TabControls = cfg.makeControlsTab(Width, Height)
	if ActiveTab != "controls" {
		TabControls.Hide()
	}
	window.Pack(TabControls, ui.Pack{
		Side:  ui.N,
		FillX: true,
	})

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
func (c Settings) makeOptionsTab(Width, Height int) *ui.Frame {
	tab := ui.NewFrame("Options Tab")

	// Common click handler for all settings,
	// so we can write the updated info to disk.
	onClick := func(ed ui.EventData) error {
		saveGameSettings()
		return nil
	}

	rows := []struct {
		Header       string
		Text         string
		Boolean      *bool
		TextVariable *string
		PadY         int
		PadX         int
		name         string // for special cases
	}{
		{
			Text: "Notice: all settings are temporary and controls are not editable.",
			PadY: 2,
		},
		{
			Header: "Game Options",
		},
		{
			Boolean: c.HorizontalToolbars,
			Text:    "Editor: Horizontal instead of vertical toolbars",
			PadX:    4,
			name:    "toolbars",
		},
		{
			Header: "Debug Options (temporary)",
		},
		{
			Boolean: c.DebugOverlay,
			Text:    "Show debug text overlay (F3)",
			PadX:    4,
		},
		{
			Boolean: c.DebugCollision,
			Text:    "Show collision hitboxes (F4)",
			PadX:    4,
		},
		{
			Header: "My Custom Content",
		},
		{
			Text: "Levels and doodads you create in-game are placed in your\n" +
				"Profile Directory. This is also where you can place content made\n" +
				"by others to use them in your game. Click on the button below\n" +
				"to (hopefully) be taken to your Profile Directory:",
		},
	}
	for _, row := range rows {
		row := row
		frame := ui.NewFrame("Frame")
		tab.Pack(frame, ui.Pack{
			Side:  ui.N,
			FillX: true,
			PadY:  row.PadY,
		})

		// Headers get their own row to themselves.
		if row.Header != "" {
			label := ui.NewLabel(ui.Label{
				Text: row.Header,
				Font: balance.LabelFont,
			})
			frame.Pack(label, ui.Pack{
				Side: ui.W,
				PadX: row.PadX,
			})
			continue
		}

		// Checkboxes get their own row.
		if row.Boolean != nil {
			cb := ui.NewCheckbox(row.Text, row.Boolean, ui.NewLabel(ui.Label{
				Text: row.Text,
				Font: balance.UIFont,
			}))
			cb.Handle(ui.Click, onClick)
			cb.Supervise(c.Supervisor)

			// Add warning to the toolbars option if the EditMode is currently active.
			if row.name == "toolbars" && c.SceneName == "Edit" {
				ui.NewTooltip(cb, ui.Tooltip{
					Text: "Note: reload your level after changing this option.\n" +
						"Playtesting and returning will do.",
					Edge: ui.Top,
				})
			}

			frame.Pack(cb, ui.Pack{
				Side: ui.W,
				PadX: row.PadX,
			})
			continue
		}

		// Any leftover Text gets packed to the left.
		if row.Text != "" {
			tf := ui.NewFrame("TextFrame")
			label := ui.NewLabel(ui.Label{
				Text: row.Text,
				Font: balance.UIFont,
			})
			tf.Pack(label, ui.Pack{
				Side: ui.W,
			})
			frame.Pack(tf, ui.Pack{
				Side: ui.W,
			})
		}
	}

	// Button toolbar.
	btnFrame := ui.NewFrame("Button Frame")
	tab.Pack(btnFrame, ui.Pack{
		Side:  ui.N,
		FillX: true,
		PadY:  4,
	})
	for _, button := range []struct {
		Label string
		Fn    func()
		Style *style.Button
	}{
		{
			Label: "Open profile directory",
			Fn: func() {
				path := strings.ReplaceAll(userdir.ProfileDirectory, "\\", "/")
				if path[0] != '/' {
					path = "/" + path
				}
				native.OpenURL("file://" + path)
			},
			Style: &balance.ButtonPrimary,
		},
	} {
		btn := ui.NewButton(button.Label, ui.NewLabel(ui.Label{
			Text: button.Label,
			Font: balance.UIFont,
		}))
		if button.Style != nil {
			btn.SetStyle(button.Style)
		}
		btn.Handle(ui.Click, func(ed ui.EventData) error {
			button.Fn()
			return nil
		})
		c.Supervisor.Add(btn)
		btnFrame.Pack(btn, ui.Pack{
			Side:   ui.W,
			Expand: true,
		})
	}

	return tab
}

// Settings Window "Controls" Tab
func (c Settings) makeControlsTab(Width, Height int) *ui.Frame {
	var (
		halfWidth        = (Width - 4) / 2 // the 4 is for window borders, TODO
		shortcutTabWidth = float64(halfWidth) * 0.5
		infoTabWidth     = float64(halfWidth) * 0.5
		rowHeight        = 20

		shortcutTabSize = render.NewRect(int(shortcutTabWidth), rowHeight)
		infoTabSize     = render.NewRect(int(infoTabWidth), rowHeight)
	)
	frame := ui.NewFrame("Controls Tab")

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
				PadY:  2,
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
