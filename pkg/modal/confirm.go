package modal

import (
	"fmt"

	"git.kirsle.net/apps/doodle/pkg/balance"
	"git.kirsle.net/go/ui"
)

// Confirm pops up an Ok/Cancel modal.
func Confirm(message string, args ...interface{}) *Modal {
	if !ready {
		panic("modal.Confirm(): not ready")
	} else if current != nil {
		return current
	}

	// Reset the supervisor.
	supervisor = ui.NewSupervisor()

	m := &Modal{
		title:   "Confirm",
		message: fmt.Sprintf(message, args...),
	}
	m.window = makeConfirm(m)

	center(m.window)
	current = m

	return m
}

// makeConfirm creates the ui.Window for the Confirm modal.
func makeConfirm(m *Modal) *ui.Window {
	win := ui.NewWindow("Confirm")
	_, title := win.TitleBar()
	title.TextVariable = &m.title

	msgFrame := ui.NewFrame("Confirm Message")
	win.Pack(msgFrame, ui.Pack{
		Side: ui.N,
	})

	msg := ui.NewLabel(ui.Label{
		TextVariable: &m.message,
		Font:         balance.UIFont,
	})
	msgFrame.Pack(msg, ui.Pack{
		Side: ui.N,
	})

	// Ok/Cancel button bar.
	btnBar := ui.NewFrame("Button Bar")
	msgFrame.Pack(btnBar, ui.Pack{
		Side: ui.N,
		PadY: 4,
	})

	for _, btn := range []struct {
		Label string
		F     func(ui.EventData) error
	}{
		{"Ok", func(ev ui.EventData) error {
			m.Dismiss(true)
			return nil
		}},
		{"Cancel", func(ev ui.EventData) error {
			m.Dismiss(false)
			return nil
		}},
	} {
		btn := btn
		button := ui.NewButton(btn.Label+"Button", ui.NewLabel(ui.Label{
			Text: btn.Label,
			Font: balance.MenuFont,
		}))
		button.Handle(ui.Click, btn.F)
		button.Compute(engine)
		supervisor.Add(button)

		btnBar.Pack(button, ui.Pack{
			Side: ui.W,
			PadX: 2,
		})
	}

	win.Compute(engine)
	win.Supervise(supervisor)

	return win
}
