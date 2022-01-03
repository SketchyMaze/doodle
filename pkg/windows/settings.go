package windows

import (
	"strconv"
	"strings"

	"git.kirsle.net/apps/doodle/pkg/balance"
	"git.kirsle.net/apps/doodle/pkg/log"
	"git.kirsle.net/apps/doodle/pkg/native"
	"git.kirsle.net/apps/doodle/pkg/shmem"
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
	EnableFeatures     *bool
	CrosshairSize      *int
	CrosshairColor     *render.Color
	HideTouchHints     *bool
	DisableAutosave    *bool

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
		Height = 400
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
	onClick := func(ed ui.EventData) error {
		saveGameSettings()
		return nil
	}

	var inputBoxWidth = 120
	rows := []struct {
		Header       string
		Text         string
		Boolean      *bool
		Integer      *int
		TextVariable *string
		Color        *render.Color
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
			Boolean: c.HideTouchHints,
			Text:    "Hide touchscreen control hints during Play Mode",
			PadX:    4,
			name:    "toolbars",
		},
		{
			Boolean: c.DisableAutosave,
			Text:    "Disable auto-save in the Editor",
			PadX:    4,
			name:    "autosave",
		},
		{
			Integer: c.CrosshairSize,
			Text:    "Editor: Crosshair size (0 to disable):",
			PadX:    4,
		},
		{
			Color: c.CrosshairColor,
			Text:  "Editor: Crosshair color:",
			PadX:  4,
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
				"Profile Directory, which you can access below:",
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
		} else {
			// Reserve indented space where the checkbox would have gone.
			spacer := ui.NewFrame("Spacer")
			spacer.Resize(render.NewRect(9, 9)) // TODO: ugly UI hack ;)
			frame.Pack(spacer, ui.Pack{
				Side: ui.W,
				PadX: row.PadX,
			})
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

		// Int variables draw as a button to prompt for new value.
		// In future: TextVariable works here too.
		if row.Integer != nil {
			varButton := ui.NewButton("VarButton", ui.NewLabel(ui.Label{
				IntVariable: row.Integer,
				Font:        ui.MenuFont,
			}))
			varButton.Handle(ui.Click, func(ed ui.EventData) error {
				shmem.Prompt(row.Text+" ", func(answer string) {
					if answer == "" {
						return
					}

					a, err := strconv.Atoi(answer)
					if err != nil {
						shmem.FlashError(err.Error())
						return
					}

					if a < 0 {
						a = 0
					} else if a > 100 {
						a = 100
					}

					*row.Integer = a
					shmem.Flash("Crosshair size set to %d%% (WIP)", a)

					// call onClick to save change to disk now
					onClick(ed)
				})
				return nil
			})

			varButton.Compute(c.Engine)
			varButton.Resize(render.Rect{
				W: inputBoxWidth,
				H: varButton.Size().H,
			})

			c.Supervisor.Add(varButton)
			frame.Pack(varButton, ui.Pack{
				Side: ui.E,
				PadX: row.PadX,
			})
		}

		// Color picker button.
		if row.Color != nil {
			btn := ui.NewButton("ColorBtn", ui.NewFrame(""))
			style := style.DefaultButton
			style.Background = *row.Color
			style.HoverBackground = style.Background.Lighten(20)
			btn.SetStyle(&style)
			btn.Handle(ui.Click, func(ed ui.EventData) error {
				// Open a ColorPicker widget.
				picker, err := ui.NewColorPicker(ui.ColorPicker{
					Title:      "Select a color",
					Supervisor: c.Supervisor,
					Engine:     c.Engine,
					Color:      *row.Color,
					OnManualInput: func(callback func(render.Color)) {
						// Prompt the user to enter a hex color using the developer shell.
						shmem.Prompt("New color in hex notation: ", func(answer string) {
							if answer != "" {
								// XXX: pure white renders as invisible, fudge it a bit.
								if answer == "FFFFFF" {
									answer = "FFFFFE"
								}

								color, err := render.HexColor(answer)
								if err != nil {
									shmem.Flash("Error with that color code: %s", err)
									return
								}

								callback(color)
							}
						})
					},
				})
				if err != nil {
					log.Error("Couldn't open ColorPicker: %s", err)
					return err
				}

				picker.Then(func(color render.Color) {
					*row.Color = color
					style.Background = color
					style.HoverBackground = style.Background.Lighten(20)

					// call onClick to save change to disk now
					onClick(ed)
				})

				picker.Center(shmem.CurrentRenderEngine.WindowSize())
				picker.Show()

				return nil
			})

			btn.Compute(c.Engine)
			btn.Resize(render.Rect{
				W: inputBoxWidth,
				H: 20, // TODO
			})

			c.Supervisor.Add(btn)
			frame.Pack(btn, ui.Pack{
				Side: ui.E,
				PadX: row.PadX,
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

// Settings Window "Experimental" Tab
func (c Settings) makeExperimentalTab(tabFrame *ui.TabFrame, Width, Height int) *ui.Frame {
	tab := tabFrame.AddTab("Experimental", ui.NewLabel(ui.Label{
		Text: "Experimental",
		Font: balance.TabFont,
	}))
	tab.Resize(render.NewRect(Width-4, Height-tab.Size().H-46))

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
			Header: "Enable Experimental Features",
		},
		{
			Text: "The setting below can enable experimental features in this\n" +
				"game. These are features which are still in development and\n" +
				"may have unstable or buggy behavior.",
			PadY: 2,
		},
		{
			Header: "Viewport window",
		},
		{
			Text: "This option in the Level menu opens another view into\n" +
				"the level. Has glitchy wallpaper problems.",
			PadY: 2,
		},
		{
			Boolean: c.EnableFeatures,
			Text:    "Enable experimental features",
			PadX:    4,
		},
		{
			Text: "Restart the game for changes to take effect.",
			PadY: 2,
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

	return tab
}
