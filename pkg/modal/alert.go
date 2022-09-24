package modal

import (
	"fmt"

	"git.kirsle.net/SketchyMaze/doodle/pkg/balance"
	"git.kirsle.net/SketchyMaze/doodle/pkg/log"
	"git.kirsle.net/go/ui"
)

// Alert pops up an alert box modal.
func Alert(message string, args ...interface{}) *Modal {
	if !ready {
		panic("modal.Alert(): not ready")
	} else if current != nil {
		current.Dismiss(false)
	}

	// Reset the supervisor.
	supervisor = ui.NewSupervisor()

	m := &Modal{
		title:      "Alert",
		message:    fmt.Sprintf(message, args...),
		cancelable: true,
	}
	m.window = makeAlert(m)

	center(m.window)
	current = m

	return m
}

// alertWindow creates the ui.Window for the Alert modal.
func makeAlert(m *Modal) *ui.Window {
	win := ui.NewWindow("Alert")
	_, title := win.TitleBar()
	title.TextVariable = &m.title

	msgFrame := ui.NewFrame("Alert Message")
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

	button := ui.NewButton("Ok Button", ui.NewLabel(ui.Label{
		Text: "Ok",
		Font: balance.MenuFont,
	}))
	button.SetStyle(&balance.ButtonPrimary)
	button.Handle(ui.Click, func(ev ui.EventData) error {
		log.Info("clicked!")
		m.Dismiss(true)
		return nil
	})
	win.Pack(button, ui.Pack{
		Side: ui.N,
		PadY: 4,
	})

	button.Compute(engine)
	supervisor.Add(button)

	win.Compute(engine)
	win.Supervise(supervisor)

	return win
}
