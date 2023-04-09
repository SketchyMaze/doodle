package windows

import (
	"git.kirsle.net/SketchyMaze/doodle/pkg/balance"
	"git.kirsle.net/SketchyMaze/doodle/pkg/log"
	"git.kirsle.net/SketchyMaze/doodle/pkg/userdir"
	"git.kirsle.net/go/render"
	"git.kirsle.net/go/ui"
)

// OpenDrawing window lets the user open a drawing or play a user level.
type OpenDrawing struct {
	Supervisor *ui.Supervisor
	Engine     render.Engine
	LevelsOnly bool

	// Callback functions.
	OnOpenDrawing func(filename string)
	OnCloseWindow func()

	// Internal variables
	window *ui.Window
}

// NewOpenDrawingWindow initializes the window.
func NewOpenDrawingWindow(config OpenDrawing) *ui.Window {
	// Default options.
	var (
		title = "Open a drawing"

		// size of the popup window
		width  = 320
		height = 320
	)

	window := ui.NewWindow(title)
	window.SetButtons(ui.CloseButton)
	window.Configure(ui.Config{
		Width:      width,
		Height:     height,
		Background: render.Grey,
	})
	config.window = window

	frame := ui.NewFrame("Window Body Frame")
	window.Pack(frame, ui.Pack{
		Side:   ui.N,
		Fill:   true,
		Expand: true,
	})

	// Divide the Levels and Doodads into tabs.
	tabs := ui.NewTabFrame("Tabs")
	tabs.SetBackground(render.DarkGrey)
	config.setupLevelTab(width, height, tabs)
	if !config.LevelsOnly {
		config.setupDoodadTab(width, height, tabs)
	}
	tabs.Supervise(config.Supervisor)

	window.Pack(tabs, ui.Pack{
		Side:   ui.N,
		Expand: true,
	})

	// Close button.
	if config.OnCloseWindow != nil {
		closeBtn := ui.NewButton("Close Window", ui.NewLabel(ui.Label{
			Text: "Close",
			Font: balance.MenuFont,
		}))
		closeBtn.Handle(ui.Click, func(ed ui.EventData) error {
			config.OnCloseWindow()
			return nil
		})
		config.Supervisor.Add(closeBtn)
		window.Place(closeBtn, ui.Place{
			Bottom: 15,
			Center: true,
		})
	}

	window.Supervise(config.Supervisor)
	window.Hide()
	return window
}

func (config OpenDrawing) setupLevelTab(width, height int, tabs *ui.TabFrame) *ui.Frame {
	frame := tabs.AddTab("Levels", ui.NewLabel(ui.Label{
		Text: "My Levels",
		Font: balance.TabFont,
	}))

	// Levels Listbox.
	levelList := ui.NewListBox("Levels", ui.ListBox{})
	levelList.Resize(render.NewRect(width-10, height-80))
	frame.Pack(levelList, ui.Pack{
		Side:   ui.N,
		Expand: true,
	})

	levelList.Handle(ui.Change, func(ed ui.EventData) error {
		filename, _ := ed.Value.(string)
		log.Info("Clicked on: %s", filename)
		if config.OnOpenDrawing != nil {
			config.OnOpenDrawing(filename)
		}
		return nil
	})

	// Get the user's levels.
	levels, _ := userdir.ListLevels()

	for _, lvl := range levels {
		levelList.AddLabel(lvl, lvl, func() {})
	}

	levelList.Supervise(config.Supervisor)

	return frame
}

func (config OpenDrawing) setupDoodadTab(width, height int, tabs *ui.TabFrame) *ui.Frame {
	frame := tabs.AddTab("Doodads", ui.NewLabel(ui.Label{
		Text: "My Custom Doodads",
		Font: balance.TabFont,
	}))

	// Doodads Listbox.
	doodadList := ui.NewListBox("Doodads", ui.ListBox{})
	doodadList.Resize(render.NewRect(width-10, height-80))
	frame.Pack(doodadList, ui.Pack{
		Side:   ui.N,
		Expand: true,
	})

	doodadList.Handle(ui.Change, func(ed ui.EventData) error {
		filename, _ := ed.Value.(string)
		log.Info("Clicked on: %s", filename)
		if config.OnOpenDrawing != nil {
			config.OnOpenDrawing(filename)
		}
		return nil
	})

	// Get the user's doodads.
	doodads, _ := userdir.ListDoodads()

	for _, lvl := range doodads {
		doodadList.AddLabel(lvl, lvl, func() {})
	}

	doodadList.Supervise(config.Supervisor)

	return frame
}
