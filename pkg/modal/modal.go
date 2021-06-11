// Package modal provides UI pop-up modals for Doodle.
package modal

import (
	"git.kirsle.net/apps/doodle/pkg/balance"
	"git.kirsle.net/apps/doodle/pkg/keybind"
	"git.kirsle.net/go/render"
	"git.kirsle.net/go/render/event"
	"git.kirsle.net/go/ui"
)

// Package global variables.
var (
	ready   bool   // Has been initialized with a render.Engine
	current *Modal // Current modal object, nil if no modal active.
	engine  render.Engine
	window  render.Rect // cached window dimensions

	supervisor *ui.Supervisor
	screen     *ui.Frame
)

// Initialize the modal package.
func Initialize(e render.Engine) {
	engine = e
	supervisor = ui.NewSupervisor()

	width, height := engine.WindowSize()
	window = render.NewRect(width, height)

	screen = ui.NewFrame("Modal Screen")
	screen.SetBackground(balance.ModalBackdrop)
	screen.Resize(window)
	screen.Compute(e)

	ready = true
}

// Reset the modal state (closing all modals).
func Reset() {
	supervisor = nil
	current = nil
}

// Handled runs the modal manager's logic. Returns true if a modal
// is presently active, to signal to Doodle not to run game logic.
func Handled(ev *event.State) bool {
	if !ready || current == nil {
		return false
	}

	// Enter key submits the default button.
	if keybind.Enter(ev) {
		current.Dismiss(true)
		return true
	}

	supervisor.Loop(ev)

	// Has the window changed size?
	size := render.NewRect(engine.WindowSize())
	if size != window {
		window = size
		screen.Resize(window)
	}

	return true
}

// Draw the modal UI to the screen.
func Draw() {
	if ready && current != nil {
		screen.Present(engine, render.Origin)
		supervisor.Present(engine)
	}
}

// Center the window on screen.
func center(win *ui.Window) {
	var modSize = win.Size()
	var moveTo = render.Point{
		X: (window.W / 2) - (modSize.W / 2),
		Y: (window.H / 4) - (modSize.H / 2),
	}
	win.MoveTo(moveTo)

	// HACK: ideally the modal should auto-size itself, but currently
	// the body of the window juts out the right and bottom side by
	// a few pixels. Fix the underlying problem later, for now we
	// set the modal size to big enough to hide the problem.
	win.Children()[0].Resize(render.NewRect(modSize.W+12, modSize.H+12))
}

// Modal is an instance of a modal, i.e. Alert or Confirm.
type Modal struct {
	title    string
	message  string
	window   *ui.Window
	callback func()
}

// WithTitle sets the title of the modal.
func (m *Modal) WithTitle(title string) *Modal {
	m.title = title
	return m
}

// Then calls a function after the modal is answered.
func (m *Modal) Then(f func()) *Modal {
	m.callback = f
	return m
}

// Dismiss the modal and optionally call the callback function.
func (m *Modal) Dismiss(call bool) {
	if call && m.callback != nil {
		m.callback()
	}
	Reset()
}
