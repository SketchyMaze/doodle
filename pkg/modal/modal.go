// Package modal provides global pop-up confirmation modals (Alert, Confirm).
package modal

import (
	"git.kirsle.net/go/render"
	"git.kirsle.net/go/ui"
)

// Global state variables.
var (
	// Darkening screen that covers the whole window.
	screen = *ui.Frame
	modalActive bool
)

func init() {
	screen = ui.NewFrame("screen")
	screen.SetBackground(render.RGBA(0, 0, 0, 180))
}

type Modal struct {
	type string
	title string
	message string
}

func (m *Modal) Then(fn func) {

}

// Present the modal view.
func Present(e render.Engine) {

}

// Alert the user to an important message. The callback function is invoked
// after the user confirms the message.
func Alert(message string, v ...interface{}) Modal {
	return Modal{
		type: "alert",
		title: "Alert!",
		message: fmt.Sprintf(message, v...),
	}
}
