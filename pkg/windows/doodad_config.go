package windows

import (
	"fmt"
	"sort"

	"git.kirsle.net/SketchyMaze/doodle/pkg/balance"
	"git.kirsle.net/SketchyMaze/doodle/pkg/level"
	"git.kirsle.net/SketchyMaze/doodle/pkg/log"
	"git.kirsle.net/SketchyMaze/doodle/pkg/shmem"
	"git.kirsle.net/SketchyMaze/doodle/pkg/uix"
	magicform "git.kirsle.net/SketchyMaze/doodle/pkg/uix/magic-form"
	"git.kirsle.net/go/render"
	"git.kirsle.net/go/ui"
)

// DoodadConfig window is what pops up in Edit Mode when you mouse over
// an actor and click its properties button.
type DoodadConfig struct {
	// Settings passed in by doodle
	Supervisor *ui.Supervisor
	Engine     render.Engine

	// Configuration options.
	EditActor *uix.Actor
	ActiveTab string // specify the tab to open
	OnRefresh func() // caller should rebuild the window

	// Widgets.
	TabFrame *ui.TabFrame
}

// NewSettingsWindow initializes the window.
func NewDoodadConfigWindow(cfg *DoodadConfig) *ui.Window {
	var (
		Width  = 400
		Height = 300
	)

	window := ui.NewWindow(cfg.EditActor.Doodad().Title + " - Actor Properties")
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
	cfg.makeOptionsTab(tabFrame, Width, Height)

	tabFrame.Supervise(cfg.Supervisor)

	return window
}

// DoodadConfig Window "Metadata" Tab
func (c *DoodadConfig) makeMetaTab(tabFrame *ui.TabFrame, Width, Height int) *ui.Frame {
	tab := tabFrame.AddTab("Metadata", ui.NewLabel(ui.Label{
		Text: "Metadata",
		Font: balance.TabFont,
	}))
	tab.Resize(render.NewRect(Width-4, Height-tab.Size().H-46))

	if c.EditActor == nil {
		return tab
	}

	var (
		doodad  = c.EditActor.Doodad()
		actorID = c.EditActor.ID()
		// actorPos        = c.EditActor.Position().String()
	)

	form := magicform.Form{
		Supervisor: c.Supervisor,
		Engine:     c.Engine,
		Vertical:   true,
		LabelWidth: 110,
		PadY:       2,
	}
	fields := []magicform.Field{
		{
			Label: "Doodad",
			Font:  balance.LabelFont,
		},
		{
			Label:        "Title:",
			Type:         magicform.Value,
			Font:         balance.UIFont,
			TextVariable: &doodad.Title,
		},
		{
			Label:        "Author:",
			Type:         magicform.Value,
			Font:         balance.UIFont,
			TextVariable: &doodad.Author,
		},

		{
			Label: "Actor (Doodad instance in level)",
			Font:  balance.LabelFont,
		},
		{
			Label:        "Actor ID:",
			Type:         magicform.Value,
			Font:         balance.UIFont,
			TextVariable: &actorID,
		},
		/* TODO: doesn't update dynamically enough
		{
			Label:        "World Position:",
			Type:         magicform.Value,
			Font:         balance.UIFont,
			TextVariable: actorPos,
		},
		*/
	}

	form.Create(tab, fields)

	return tab
}

// SetTextable is a Button or Checkbox widget having a SetText function,
// to support the reset button on the Doodad Options tab.
type SetTextable interface {
	SetText(string) error
}

// DoodadConfig Window "Tags" Tab
func (c DoodadConfig) makeOptionsTab(tabFrame *ui.TabFrame, Width, Height int) *ui.Frame {
	tab := tabFrame.AddTab("Options", ui.NewLabel(ui.Label{
		Text: "Options",
		Font: balance.TabFont,
	}))
	tab.Resize(render.NewRect(Width-4, Height-tab.Size().H-46))

	if c.EditActor == nil {
		return tab
	}

	// Draw a table view of the current tags on this doodad.
	var (
		doodad  = c.EditActor.Doodad()
		headers = []string{"Type", "Name", "Value", "Reset"}
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
	if len(doodad.Options) == 0 {
		label := ui.NewLabel(ui.Label{
			Text: "There are no options on this doodad.",
			Font: balance.MenuFont,
		})
		tab.Pack(label, ui.Pack{
			Side:  ui.N,
			FillX: true,
		})
	} else {
		// Initialize the Actor Options if nil
		if c.EditActor.Actor.Options == nil {
			c.EditActor.Actor.Options = map[string]*level.Option{}
		}

		// Draw the rows for each tag.
		var sortedOpts []string
		for name := range doodad.Options {
			sortedOpts = append(sortedOpts, name)
		}
		sort.Strings(sortedOpts)

		for _, optName := range sortedOpts {
			var (
				name  = optName
				value = c.EditActor.GetOption(name)
			)

			if value == nil {
				continue
			}

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
			var cbValue bool
			if value.Type == "bool" {
				cbValue = value.Value.(bool)
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
					c.EditActor.Actor.SetOption(name, value.Type, label)
					checkbox.SetText(label)
					return nil
				})
				checkbox.Supervise(c.Supervisor)
				btnValue = checkbox
			} else {
				button := ui.NewButton("Tag Button", ui.NewLabel(ui.Label{
					Text: fmt.Sprintf("%v", value.Value),
					Font: balance.MenuFont,
				}))
				button.Resize(render.NewRect(columns[2], height))
				button.Handle(ui.Click, func(ed ui.EventData) error {
					shmem.Prompt("Enter new value: ", func(answer string) {
						if answer == "" {
							return
						}
						answer = c.EditActor.Actor.SetOption(name, value.Type, answer)
						button.SetText(answer)
					})
					return nil
				})
				c.Supervisor.Add(button)
				btnValue = button
			}

			// "Delete" / Reset Button: removes the Actor Option so it falls
			// back onto the default Doodad Option.
			btnDelete := ui.NewButton("Delete Button", ui.NewLabel(ui.Label{
				Text: "Reset",
				Font: balance.MenuFont,
			}))
			btnDelete.Resize(render.NewRect(columns[3], height))
			btnDelete.SetStyle(&balance.ButtonDanger)
			btnDelete.Handle(ui.Click, func(ed ui.EventData) error {
				log.Info("Delete option: %s", name)
				delete(c.EditActor.Actor.Options, name)

				// Update the value button's text label.
				if stt, ok := btnValue.(SetTextable); ok {
					value := c.EditActor.GetOption(name)
					if value != nil {
						stt.SetText(fmt.Sprintf("%v", value.Value))

						// Set the correct boolean checkbox state.
						if value.Type == "bool" {
							cbValue = value.Value.(bool)
						}
					}
				}

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

	return tab
}
