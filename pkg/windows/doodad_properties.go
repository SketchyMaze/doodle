package windows

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"git.kirsle.net/SketchyMaze/doodle/assets"
	"git.kirsle.net/SketchyMaze/doodle/pkg/balance"
	"git.kirsle.net/SketchyMaze/doodle/pkg/doodads"
	"git.kirsle.net/SketchyMaze/doodle/pkg/log"
	"git.kirsle.net/SketchyMaze/doodle/pkg/modal"
	"git.kirsle.net/SketchyMaze/doodle/pkg/native"
	"git.kirsle.net/SketchyMaze/doodle/pkg/shmem"
	"git.kirsle.net/SketchyMaze/doodle/pkg/userdir"
	"git.kirsle.net/go/render"
	"git.kirsle.net/go/ui"
)

// Some generic built-in doodad scripts users can attach.
var GenericScripts = []struct {
	Label    string
	Help     string
	Filename string
	SetTags  map[string]string
}{
	{
		Label: "Generic Solid",
		Help: "The whole canvas of your doodad acts solid.\n" +
			"The player and other mobile doodads can walk on\n" +
			"top of it, and it blocks movement from the sides.",
		Filename: "assets/scripts/generic-solid.js",
	},
	{
		Label: "Generic Fire",
		Help: "The whole canvas of your doodad acts like fire.\n" +
			"Mobile doodads who touch it turn dark, and if\n" +
			"the player touches it - game over! The failure\n" +
			"message says: 'Watch out for (title)!'",
		Filename: "assets/scripts/generic-fire.js",
	},
	{
		Label: "Generic Anvil",
		Help: "This doodad will behave like the Anvil: fall with\n" +
			"gravity and be deadly to any mobile doodad that it\n" +
			"lands on! The failure message says:\n" +
			"'Watch out for (title)!'",
		Filename: "assets/scripts/generic-anvil.js",
	},
	{
		Label: "Generic Collectible Item",
		Help: "This doodad will behave like a pocketable item, like\n" +
			"the Keys. Tip: set a Doodad tag like quantity=0 to set\n" +
			"the item quantity when picked up (default is 1).",
		Filename: "assets/scripts/generic-item.js",
		SetTags: map[string]string{
			"quantity": "1",
		},
	},
}

// DoodadProperties window.
type DoodadProperties struct {
	// Settings passed in by doodle
	Supervisor *ui.Supervisor
	Engine     render.Engine

	// Configuration options.
	EditDoodad *doodads.Doodad
	ActiveTab  string // specify the tab to open
	OnRefresh  func() // caller should rebuild the window

	// Widgets.
	TabFrame *ui.TabFrame
}

// HACKY GLOBAL VARIABLE
var (
	showTagsOnRefreshDoodadPropertiesWindow bool
	showOptsOnRefreshDoodadPropertiesWindow bool
)

// NewSettingsWindow initializes the window.
func NewDoodadPropertiesWindow(cfg *DoodadProperties) *ui.Window {
	var (
		Width  = 400
		Height = 300
	)

	window := ui.NewWindow("Doodad Properties")
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
	cfg.TabFrame = tabFrame

	// Make the tabs.
	cfg.makeMetaTab(tabFrame, Width, Height)
	cfg.makeTagsTab(tabFrame, Width, Height)
	cfg.makeOptionsTab(tabFrame, Width, Height)

	if showTagsOnRefreshDoodadPropertiesWindow {
		tabFrame.SetTab("Tags")
		showTagsOnRefreshDoodadPropertiesWindow = false
	} else if showOptsOnRefreshDoodadPropertiesWindow {
		tabFrame.SetTab("Options")
		showOptsOnRefreshDoodadPropertiesWindow = false
	}

	tabFrame.Supervise(cfg.Supervisor)

	return window
}

// DoodadProperties Window "Metadata" Tab
func (c DoodadProperties) makeMetaTab(tabFrame *ui.TabFrame, Width, Height int) *ui.Frame {
	tab := tabFrame.AddTab("Metadata", ui.NewLabel(ui.Label{
		Text: "Metadata",
		Font: balance.TabFont,
	}))
	tab.Resize(render.NewRect(Width-4, Height-tab.Size().H-46))

	if c.EditDoodad == nil {
		return tab
	}

	//////////////
	// Draw the editable metadata form.
	var hitboxString = c.EditDoodad.Hitbox.String()
	for _, data := range []struct {
		Label    string
		Prompt   string // optional
		Variable *string
		Update   func(string)
	}{
		{
			Label:    "Title:",
			Variable: &c.EditDoodad.Title,
			Update: func(v string) {
				c.EditDoodad.Title = v
			},
		},
		{
			Label:    "Author:",
			Variable: &c.EditDoodad.Author,
			Update: func(v string) {
				c.EditDoodad.Author = v
			},
		},
		{
			Label:    "Hitbox:",
			Prompt:   "Enter hitbox in X,Y,W,H or just W,H format: ",
			Variable: &hitboxString,
			Update: func(v string) {
				// Parse it.
				parts := strings.Split(v, ",")
				var ints []int
				for _, part := range parts {
					a, err := strconv.Atoi(strings.TrimSpace(part))
					if err != nil {
						shmem.Flash("Invalid format for hitbox, using the default")
						return
					}
					ints = append(ints, a)
				}

				if len(ints) == 2 {
					c.EditDoodad.Hitbox = render.NewRect(ints[0], ints[1])
				} else if len(ints) == 4 {
					c.EditDoodad.Hitbox = render.Rect{
						X: ints[0],
						Y: ints[1],
						W: ints[2],
						H: ints[3],
					}
				} else {
					shmem.Flash("Hitbox should be in X,Y,W,H or just W,H format, 2 or 4 numbers.")
					return
				}

				hitboxString = c.EditDoodad.Hitbox.String()
			},
		},
	} {
		data := data
		frame := ui.NewFrame("Metadata " + data.Label + " Frame")
		tab.Pack(frame, ui.Pack{
			Side:  ui.N,
			PadY:  4,
			FillX: true,
		})

		// The label
		label := ui.NewLabel(ui.Label{
			Text: data.Label,
			Font: balance.MenuFont,
		})
		label.Configure(ui.Config{
			Width: 75,
		})
		frame.Pack(label, ui.Pack{
			Side: ui.W,
		})

		// The button.
		btn := ui.NewButton(data.Label, ui.NewLabel(ui.Label{
			TextVariable: data.Variable,
			Font:         balance.MenuFont,
		}))
		btn.Handle(ui.Click, func(ed ui.EventData) error {
			var prompt = data.Prompt
			if prompt == "" {
				prompt = "Enter a new " + data.Label + " "
			}

			shmem.Prompt(prompt, func(answer string) {
				if answer != "" {
					data.Update(answer)
				}
			})
			return nil
		})
		c.Supervisor.Add(btn)
		frame.Pack(btn, ui.Pack{
			Side:   ui.W,
			Expand: true,
			PadX:   2,
		})
	}

	//////////////////////////////////
	// Draw the JavaScript management

	scriptHeader := ui.NewLabel(ui.Label{
		Text: "Doodad Script",
		Font: balance.LargeLabelFont,
	})
	tab.Pack(scriptHeader, ui.Pack{
		Side:  ui.N,
		FillX: true,
		PadY:  8,
	})

	// Frame for if a script does exist on the doodad.
	var (
		ifScript   *ui.Frame
		elseScript *ui.Frame
	)

	// "If Script" Frame
	{
		ifScript = ui.NewFrame("If Script")
		tab.Pack(ifScript, ui.Pack{
			Side:  ui.N,
			FillX: true,
		})

		label := ui.NewLabel(ui.Label{
			Text: "This Doodad has a script attached.",
			Font: balance.MenuFont,
		})
		ifScript.Pack(label, ui.Pack{
			Side: ui.W,
		})

		// Delete Button
		deleteBtn := ui.NewButton("Save", ui.NewLabel(ui.Label{
			Text: "Delete",
			Font: balance.MenuFont,
		}))
		deleteBtn.SetStyle(&balance.ButtonDanger)
		deleteBtn.Handle(ui.Click, func(ed ui.EventData) error {
			modal.Confirm("Are you sure you want to delete this script?").Then(func() {
				c.EditDoodad.Script = ""
				ifScript.Hide()
				elseScript.Show()
			})
			return nil
		})
		c.Supervisor.Add(deleteBtn)
		ifScript.Pack(deleteBtn, ui.Pack{
			Side: ui.E,
			PadX: 2,
		})

		// Open Button
		saveBtn := ui.NewButton("Open", ui.NewLabel(ui.Label{
			Text: "View",
			Font: balance.MenuFont,
		}))
		saveBtn.SetStyle(&balance.ButtonPrimary)
		saveBtn.Handle(ui.Click, func(ed ui.EventData) error {
			// Write the js file to cache and try and open it in the user's
			// native text editor program.
			outname := filepath.Join(userdir.CacheDirectory, c.EditDoodad.Filename+".js")
			err := ioutil.WriteFile(outname, []byte(c.EditDoodad.Script), 0644)
			if err == nil {
				native.OpenLocalURL(outname)
				return nil
			}

			// Otherwise, prompt the user for their filepath.
			shmem.Prompt("Save script as (*.js): ", func(answer string) {
				if answer != "" {
					cwd, _ := os.Getwd()
					err := ioutil.WriteFile(answer, []byte(c.EditDoodad.Script), 0644)
					if err != nil {
						shmem.Flash(err.Error())
					} else {
						shmem.Flash("Written to: %s (%d bytes)", filepath.Join(cwd, answer), len(c.EditDoodad.Script))
						native.OpenLocalURL(filepath.Join(cwd, answer))
					}
				}
			})
			return nil
		})
		c.Supervisor.Add(saveBtn)
		ifScript.Pack(saveBtn, ui.Pack{
			Side: ui.E,
			PadX: 2,
		})
	}

	// "Else Script" Frame
	{
		elseScript = ui.NewFrame("If Script")
		tab.Pack(elseScript, ui.Pack{
			Side:  ui.N,
			FillX: true,
		})

		label := ui.NewLabel(ui.Label{
			Text: "There is no script attached to this doodad.",
			Font: balance.MenuFont,
		})
		elseScript.Pack(label, ui.Pack{
			Side: ui.W,
		})
	}

	// Attaching a Script Frame
	{
		label := ui.NewLabel(ui.Label{
			Text: "Attach a Script",
			Font: balance.LabelFont,
		})
		tab.Pack(label, ui.Pack{
			Side:  ui.N,
			FillX: true,
		})

		frame := ui.NewFrame("Attach Script Frame")
		tab.Pack(frame, ui.Pack{
			Side:  ui.N,
			FillX: true,
		})

		// Browse Script label.
		lblBrowse := ui.NewLabel(ui.Label{
			Text: "Browse and attach a .js file:",
			Font: balance.MenuFont,
		})
		frame.Pack(lblBrowse, ui.Pack{
			Side: ui.W,
		})

		// Browse Script button.
		btnBrowse := ui.NewButton("Browse Script", ui.NewLabel(ui.Label{
			Text: "Attach a script...",
			Font: balance.MenuFont,
		}))
		btnBrowse.SetStyle(&balance.ButtonPrimary)
		btnBrowse.Handle(ui.Click, func(ed ui.EventData) error {
			filename, err := native.OpenFile("Choose a .js file", "*.js")
			if err != nil {
				shmem.Flash("Couldn't show file dialog: %s", err)
				return nil
			}

			data, err := ioutil.ReadFile(filename)
			if err != nil {
				shmem.Flash("Couldn't read file: %s", err)
				return nil
			}

			c.EditDoodad.Script = string(data)
			shmem.Flash("Attached %d-byte script to this doodad.", len(c.EditDoodad.Script))

			// Toggle the if/else frames.
			ifScript.Show()
			elseScript.Hide()

			return nil
		})
		c.Supervisor.Add(btnBrowse)
		frame.Pack(btnBrowse, ui.Pack{
			Side: ui.E,
		})
	}

	// Built-in Generic Scripts Frame
	{
		frame := ui.NewFrame("Generic Scripts Frame")
		tab.Pack(frame, ui.Pack{
			Side:  ui.N,
			FillX: true,
			PadY:  4,
		})

		label := ui.NewLabel(ui.Label{
			Text: "Or select from a generic script:",
			Font: ui.MenuFont,
		})
		frame.Pack(label, ui.Pack{
			Side: ui.W,
		})

		// SelectBox for the built-ins.
		sb := ui.NewSelectBox("Select", ui.Label{
			Font: ui.MenuFont,
		})
		tab.Pack(sb, ui.Pack{
			Side:  ui.N,
			FillX: true,
		})

		for _, script := range GenericScripts {
			sb.AddItem(script.Label, script.Filename, func() {})
		}
		sb.SetValue(GenericScripts[0].Filename)
		sb.AlwaysChange = true
		sb.Handle(ui.Change, func(ed ui.EventData) error {
			if selection, ok := sb.GetValue(); ok {
				if filename, ok := selection.Value.(string); ok {
					// Get this script from the built-in assets.
					data, err := assets.Asset(filename)
					if err != nil {
						shmem.Flash("Couldn't get script: %s", err)
						return nil
					}

					// Find the data from the builtins.
					var label, help string
					var setTags map[string]string
					for _, script := range GenericScripts {
						if script.Filename == filename {
							label = script.Label
							help = script.Help
							setTags = script.SetTags
							break
						}
					}

					// Prompt the user + a description of this option.
					var (
						basename    = filepath.Base(filename)
						description = fmt.Sprintf(
							"Do you want to install %s to your doodad?\n\n"+
								"%s\n\n%s",
							basename,
							label,
							help,
						)
					)

					modal.Confirm(description).Then(func() {
						c.EditDoodad.Script = string(data)

						shmem.Flash("Attached %s to your doodad", filepath.Base(filename))

						// Set any tags that come with this script.
						if setTags != nil && len(setTags) > 0 {
							for k, v := range setTags {
								log.Info("Set doodad tag %s=%s", k, v)
								c.EditDoodad.Tags[k] = v
							}
						}

						// Toggle the if/else frames.
						ifScript.Show()
						elseScript.Hide()
					})
				}
			}

			return nil
		})

		sb.Supervise(c.Supervisor)
		c.Supervisor.Add(sb)
	}

	// Show/hide appropriate frames.
	if c.EditDoodad.Script == "" {
		ifScript.Hide()
		elseScript.Show()
	} else {
		ifScript.Show()
		elseScript.Hide()
	}

	return tab
}

// DoodadProperties Window "Tags" Tab
func (c DoodadProperties) makeTagsTab(tabFrame *ui.TabFrame, Width, Height int) *ui.Frame {
	tab := tabFrame.AddTab("Tags", ui.NewLabel(ui.Label{
		Text: "Tags",
		Font: balance.TabFont,
	}))
	tab.Resize(render.NewRect(Width-4, Height-tab.Size().H-46))

	if c.EditDoodad == nil {
		return tab
	}

	// Draw a table view of the current tags on this doodad.
	var (
		headers = []string{"Name", "Value", "Del."}
		columns = []int{150, 150, 80} // TODO, Width=400
		height  = 24
		row     = ui.NewFrame("HeaderRow")
	)
	tab.Pack(row, ui.Pack{
		Side:  ui.N,
		FillX: true,
	})
	for i, value := range headers {
		cell := ui.NewLabel(ui.Label{
			Text: value,
			Font: balance.MenuFontBold,
		})
		cell.Resize(render.NewRect(columns[i], height))
		row.Pack(cell, ui.Pack{
			Side: ui.W,
		})
	}

	// No tags?
	if len(c.EditDoodad.Tags) == 0 {
		label := ui.NewLabel(ui.Label{
			Text: "There are no tags on this doodad.",
			Font: balance.MenuFont,
		})
		tab.Pack(label, ui.Pack{
			Side:  ui.N,
			FillX: true,
		})
	} else {
		// Draw the rows for each tag.
		var sortedTags []string
		for name := range c.EditDoodad.Tags {
			sortedTags = append(sortedTags, name)
		}
		sort.Strings(sortedTags)

		for _, tagName := range sortedTags {
			var (
				name  = tagName
				value = c.EditDoodad.Tags[name]
			)

			row = ui.NewFrame("Tag Row")
			tab.Pack(row, ui.Pack{
				Side:  ui.N,
				FillX: true,
				PadY:  2,
			})

			lblName := ui.NewLabel(ui.Label{
				Text: name,
				Font: balance.MenuFont,
			})
			lblName.Resize(render.NewRect(columns[0], height))

			btnValue := ui.NewButton("Tag Button", ui.NewLabel(ui.Label{
				Text: value,
				Font: balance.MenuFont,
			}))
			btnValue.Resize(render.NewRect(columns[1], height))
			btnValue.Handle(ui.Click, func(ed ui.EventData) error {
				shmem.Prompt("Enter new value: ", func(answer string) {
					if answer == "" {
						return
					}
					c.EditDoodad.Tags[name] = answer
					btnValue.SetText(answer)
				})
				return nil
			})
			c.Supervisor.Add(btnValue)

			btnDelete := ui.NewButton("Delete Button", ui.NewLabel(ui.Label{
				Text: "Delete",
				Font: balance.MenuFont,
			}))
			btnDelete.Resize(render.NewRect(columns[2], height))
			btnDelete.SetStyle(&balance.ButtonDanger)
			btnDelete.Handle(ui.Click, func(ed ui.EventData) error {
				modal.Confirm("Delete tag %s?", name).Then(func() {
					log.Info("Delete tag: %s", name)
					delete(c.EditDoodad.Tags, name)

					// Trigger a refresh.
					if c.OnRefresh != nil {
						showTagsOnRefreshDoodadPropertiesWindow = true
						c.OnRefresh()
					}
				})
				return nil
			})
			c.Supervisor.Add(btnDelete)

			// Pack the widgets.
			row.Pack(lblName, ui.Pack{
				Side: ui.W,
			})
			row.Pack(btnValue, ui.Pack{
				Side: ui.W,
				PadX: 4,
			})
			row.Pack(btnDelete, ui.Pack{
				Side: ui.W,
			})
		}
	}

	// Add Tag button.
	row = ui.NewFrame("Button Frame")
	tab.Pack(row, ui.Pack{
		Side:  ui.N,
		FillX: true,
	})
	btnAdd := ui.NewButton("New Tag", ui.NewLabel(ui.Label{
		Text: "Add Tag",
		Font: balance.MenuFont,
	}))
	btnAdd.SetStyle(&balance.ButtonPrimary)
	btnAdd.Handle(ui.Click, func(ed ui.EventData) error {
		shmem.Prompt("Enter name of the new tag: ", func(answer string) {
			if answer == "" {
				return
			}

			log.Info("Adding doodad tag: %s", answer)
			c.EditDoodad.Tags[answer] = ""
			if c.OnRefresh != nil {
				showTagsOnRefreshDoodadPropertiesWindow = true
				c.OnRefresh()
			}
		})
		return nil
	})
	c.Supervisor.Add(btnAdd)
	row.Pack(btnAdd, ui.Pack{
		Side: ui.E,
	})

	return tab
}

// DoodadProperties Window "Options" Tab
func (c DoodadProperties) makeOptionsTab(tabFrame *ui.TabFrame, Width, Height int) *ui.Frame {
	tab := tabFrame.AddTab("Options", ui.NewLabel(ui.Label{
		Text: "Options",
		Font: balance.TabFont,
	}))
	tab.Resize(render.NewRect(Width-4, Height-tab.Size().H-46))

	if c.EditDoodad == nil {
		return tab
	}

	// Draw a table view of the current tags on this doodad.
	var (
		headers = []string{"Type", "Name", "Default", "Del."}
		columns = []int{40, 130, 130, 80} // TODO, Width=400
		height  = 24
		row     = ui.NewFrame("HeaderRow")
	)
	tab.Pack(row, ui.Pack{
		Side:  ui.N,
		FillX: true,
	})
	for i, value := range headers {
		cell := ui.NewLabel(ui.Label{
			Text: value,
			Font: balance.MenuFontBold,
		})
		cell.Resize(render.NewRect(columns[i], height))
		row.Pack(cell, ui.Pack{
			Side: ui.W,
		})
	}

	// No tags?
	if len(c.EditDoodad.Options) == 0 {
		label := ui.NewLabel(ui.Label{
			Text: "There are no options on this doodad.",
			Font: balance.MenuFont,
		})
		tab.Pack(label, ui.Pack{
			Side:  ui.N,
			FillX: true,
		})
	} else {
		// Draw the rows for each tag.
		var sortedOpts []string
		for name := range c.EditDoodad.Options {
			sortedOpts = append(sortedOpts, name)
		}
		sort.Strings(sortedOpts)

		for _, optName := range sortedOpts {
			var (
				name  = optName
				value = c.EditDoodad.Options[name]
			)

			row = ui.NewFrame("Option Row")
			tab.Pack(row, ui.Pack{
				Side:  ui.N,
				FillX: true,
				PadY:  2,
			})

			lblType := ui.NewLabel(ui.Label{
				Text: value.Type,
				Font: balance.MenuFont,
			})
			lblType.Resize(render.NewRect(columns[0], height))

			lblName := ui.NewLabel(ui.Label{
				Text: name,
				Font: balance.MenuFont,
			})
			lblName.Resize(render.NewRect(columns[1], height))

			// Value button: show a checkbox for booleans or a clickable
			// button for other types (prompts user for value)
			var btnValue ui.Widget
			if value.Type == "bool" {
				var cbValue = value.Default.(bool)
				checkbox := ui.NewCheckbox("Bool Box", &cbValue, ui.NewLabel(ui.Label{
					Text: fmt.Sprintf("%v", cbValue),
					Font: balance.MenuFont,
				}))
				checkbox.Resize(render.NewRect(columns[2], height))
				checkbox.Handle(ui.Click, func(ed ui.EventData) error {
					var label string
					if cbValue {
						label = "true"
					} else {
						label = "false"
					}
					c.EditDoodad.Options[name].Set(label)
					checkbox.SetText(label)
					return nil
				})
				checkbox.Supervise(c.Supervisor)
				btnValue = checkbox
			} else {
				button := ui.NewButton("Tag Button", ui.NewLabel(ui.Label{
					Text: fmt.Sprintf("%v", value.Default),
					Font: balance.MenuFont,
				}))
				button.Resize(render.NewRect(columns[2], height))
				button.Handle(ui.Click, func(ed ui.EventData) error {
					shmem.Prompt("Enter new value: ", func(answer string) {
						if answer == "" {
							return
						}
						answer = c.EditDoodad.Options[name].Set(answer)
						button.SetText(answer)
					})
					return nil
				})
				c.Supervisor.Add(button)
				btnValue = button
			}

			btnDelete := ui.NewButton("Delete Button", ui.NewLabel(ui.Label{
				Text: "Del",
				Font: balance.MenuFont,
			}))
			btnDelete.Resize(render.NewRect(columns[3], height))
			btnDelete.SetStyle(&balance.ButtonDanger)
			btnDelete.Handle(ui.Click, func(ed ui.EventData) error {
				modal.Confirm("Delete option %s?", name).Then(func() {
					log.Info("Delete option: %s", name)
					delete(c.EditDoodad.Options, name)

					// Trigger a refresh.
					if c.OnRefresh != nil {
						showOptsOnRefreshDoodadPropertiesWindow = true
						c.OnRefresh()
					}
				})
				return nil
			})
			c.Supervisor.Add(btnDelete)

			// Pack the widgets.
			row.Pack(lblType, ui.Pack{
				Side: ui.W,
			})
			row.Pack(lblName, ui.Pack{
				Side: ui.W,
			})
			row.Pack(btnValue, ui.Pack{
				Side: ui.W,
				PadX: 4,
			})
			row.Pack(btnDelete, ui.Pack{
				Side: ui.W,
			})
		}
	}

	// Add Option menu button.
	row = ui.NewFrame("Button Frame")
	tab.Pack(row, ui.Pack{
		Side:  ui.N,
		FillX: true,
	})
	btnAdd := ui.NewMenuButton("New Option", ui.NewLabel(ui.Label{
		Text: "Add Option",
		Font: balance.MenuFont,
	}))
	btnAdd.SetStyle(&balance.ButtonPrimary)

	// Types of options
	for _, item := range []struct {
		label  string
		typing string
		value  interface{}
	}{
		{"Boolean", "bool", false},
		{"String", "str", ""},
		{"Integer", "int", 0},
	} {
		item := item
		btnAdd.AddItem(item.label, func() {
			shmem.Prompt("Enter name of the new boolean: ", func(answer string) {
				if answer == "" {
					return
				}

				c.EditDoodad.Options[answer] = &doodads.Option{
					Name:    answer,
					Type:    item.typing,
					Default: item.value,
				}
				if c.OnRefresh != nil {
					showOptsOnRefreshDoodadPropertiesWindow = true
					c.OnRefresh()
				}
			})
		})
	}

	btnAdd.Supervise(c.Supervisor)
	c.Supervisor.Add(btnAdd)
	row.Pack(btnAdd, ui.Pack{
		Side: ui.E,
	})

	return tab
}

func (c DoodadProperties) reloadTagFrame() {

}
