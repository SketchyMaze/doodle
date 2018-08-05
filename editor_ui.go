package doodle

import (
	"fmt"

	"git.kirsle.net/apps/doodle/balance"
	"git.kirsle.net/apps/doodle/events"
	"git.kirsle.net/apps/doodle/render"
	"git.kirsle.net/apps/doodle/ui"
)

// EditorUI manages the user interface for the Editor Scene.
type EditorUI struct {
	d *Doodle

	// Variables
	StatusMouseText string

	// Widgets
	Supervisor *ui.Supervisor
	StatusBar  *ui.Frame
}

// NewEditorUI initializes the Editor UI.
func NewEditorUI(d *Doodle) *EditorUI {
	u := &EditorUI{
		d:               d,
		Supervisor:      ui.NewSupervisor(),
		StatusMouseText: ".",
	}
	u.StatusBar = u.SetupStatusBar(d)
	return u
}

// Loop to process events and update the UI.
func (u *EditorUI) Loop(ev *events.State) {
	u.StatusMouseText = fmt.Sprintf("Mouse: (%d,%d)",
		ev.CursorX.Now,
		ev.CursorY.Now,
	)
	u.StatusBar.Compute(u.d.Engine)
	u.Supervisor.Loop(ev)
}

// Present the UI to the screen.
func (u *EditorUI) Present(e render.Engine) {
	u.StatusBar.Present(e, u.StatusBar.Point())
}

// SetupStatusBar sets up the status bar widget along the bottom of the window.
func (u *EditorUI) SetupStatusBar(d *Doodle) *ui.Frame {
	frame := ui.NewFrame("Status Bar")
	frame.Configure(ui.Config{
		BorderStyle: ui.BorderRaised,
		Background:  render.Grey,
		BorderSize:  2,
		Width:       d.width,
	})

	cursorLabel := ui.NewLabel(ui.Label{
		TextVariable: &u.StatusMouseText,
		Font:         balance.StatusFont,
	})
	cursorLabel.Configure(ui.Config{
		Background:  render.Grey,
		BorderStyle: ui.BorderSunken,
		BorderColor: render.Grey,
		BorderSize:  1,
	})
	cursorLabel.Compute(d.Engine)
	frame.Pack(cursorLabel, ui.Pack{
		Anchor: ui.W,
	})

	filenameLabel := ui.NewLabel(ui.Label{
		Text: "Filename: untitled.map",
		Font: balance.StatusFont,
	})
	filenameLabel.Configure(ui.Config{
		Background:  render.Grey,
		BorderStyle: ui.BorderSunken,
		BorderColor: render.Grey,
		BorderSize:  1,
	})
	filenameLabel.Compute(d.Engine)
	frame.Pack(filenameLabel, ui.Pack{
		Anchor: ui.W,
	})

	extraLabel := ui.NewLabel(ui.Label{
		Text: "blah",
		Font: balance.StatusFont,
	})
	extraLabel.Configure(ui.Config{
		Background:  render.Grey,
		BorderStyle: ui.BorderSunken,
		BorderColor: render.Grey,
		BorderSize:  1,
	})
	extraLabel.Compute(d.Engine)
	frame.Pack(extraLabel, ui.Pack{
		Anchor: ui.E,
	})

	frame.Resize(render.Rect{
		W: d.width,
		H: cursorLabel.BoxSize().H + frame.BoxThickness(1),
	})
	frame.Compute(d.Engine)
	frame.MoveTo(render.Point{
		X: 0,
		Y: d.height - frame.Size().H,
	})

	return frame
}
