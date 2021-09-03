package windows

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"

	"git.kirsle.net/apps/doodle/pkg/balance"
	"git.kirsle.net/apps/doodle/pkg/doodads"
	"git.kirsle.net/apps/doodle/pkg/log"
	"git.kirsle.net/apps/doodle/pkg/modal"
	"git.kirsle.net/apps/doodle/pkg/native"
	"git.kirsle.net/apps/doodle/pkg/shmem"
	"git.kirsle.net/go/render"
	"git.kirsle.net/go/ui"
)

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
var showTagsOnRefreshDoodadPropertiesWindow bool

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

	if showTagsOnRefreshDoodadPropertiesWindow {
		tabFrame.SetTab("Tags")
		showTagsOnRefreshDoodadPropertiesWindow = false
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
	for _, data := range []struct {
		Label    string
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
			shmem.Prompt("Enter a new "+data.Label+" ", func(answer string) {
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

		// Save Button
		saveBtn := ui.NewButton("Save", ui.NewLabel(ui.Label{
			Text: "Save",
			Font: balance.MenuFont,
		}))
		saveBtn.SetStyle(&balance.ButtonPrimary)
		saveBtn.Handle(ui.Click, func(ed ui.EventData) error {
			shmem.Prompt("Save script as (*.js): ", func(answer string) {
				if answer != "" {
					cwd, _ := os.Getwd()
					err := ioutil.WriteFile(answer, []byte(c.EditDoodad.Script), 0644)
					if err != nil {
						shmem.Flash(err.Error())
					} else {
						shmem.Flash("Written to: %s (%d bytes)", filepath.Join(cwd, answer), len(c.EditDoodad.Script))
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
	tab.Pack(btnBrowse, ui.Pack{
		Side:    ui.N,
		Padding: 4,
	})

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

func (c DoodadProperties) reloadTagFrame() {

}
