package modal

import (
	"fmt"

	"git.kirsle.net/apps/doodle/pkg/balance"
	"git.kirsle.net/go/ui"
)

// ConfigEndLevel sets options for the EndLevel modal.
type ConfigEndLevel struct {
	Success bool // false = failure condition

	// Handler functions - what you don't define will not
	// show as buttons in the modal.
	OnRestartLevel    func() // Restart Level
	OnRetryCheckpoint func() // Continue from checkpoint
	OnEditLevel       func()
	OnNextLevel       func() // Next Level
	OnExitToMenu      func() // Exit to Menu
}

// EndLevel shows the End Level modal.
func EndLevel(cfg ConfigEndLevel, title, message string, args ...interface{}) *Modal {
	if !ready {
		panic("modal.EndLevel(): not ready")
	} else if current != nil {
		return current
	}

	// Reset the supervisor.
	supervisor = ui.NewSupervisor()

	m := &Modal{
		title:   title,
		message: fmt.Sprintf(message, args...),
	}
	m.window = makeEndLevel(m, cfg)

	center(m.window)
	current = m

	return m
}

// makeEndLevel creates the ui.Window for the Confirm modal.
func makeEndLevel(m *Modal, cfg ConfigEndLevel) *ui.Window {
	win := ui.NewWindow("EndLevel")
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

	var buttons []*ui.Button
	var primaryFunc func()
	for _, btn := range []struct {
		Label string
		F     func()
	}{
		{
			Label: "Next Level",
			F:     cfg.OnNextLevel,
		},
		{
			Label: "Retry from Checkpoint",
			F:     cfg.OnRetryCheckpoint,
		},
		{
			Label: "Restart Level",
			F:     cfg.OnRestartLevel,
		},
		{
			Label: "Edit Level",
			F:     cfg.OnEditLevel,
		},
		{
			Label: "Exit to Menu",
			F:     cfg.OnExitToMenu,
		},
	} {
		btn := btn
		if btn.F == nil {
			continue
		}

		if primaryFunc == nil {
			primaryFunc = btn.F
		}

		button := ui.NewButton(btn.Label+"Button", ui.NewLabel(ui.Label{
			Text: btn.Label,
			Font: balance.MenuFont,
		}))
		button.Handle(ui.Click, func(ed ui.EventData) error {
			btn.F()
			m.Dismiss(false)
			return nil
		})
		button.Compute(engine)
		buttons = append(buttons, button)
		supervisor.Add(button)

		btnBar.Pack(button, ui.Pack{
			Side:  ui.N,
			PadY:  2,
			FillX: true,
		})

		// // Make a new row of buttons?
		// if i > 0 && i%3 == 0 {
		// 	btnBar = ui.NewFrame("Button Bar")
		// 	msgFrame.Pack(btnBar, ui.Pack{
		// 		Side: ui.N,
		// 		PadY: 0,
		// 	})
		// }
	}

	// Mark the first button the primary button.
	if primaryFunc != nil {
		m.Then(primaryFunc)
	}
	buttons[0].SetStyle(&balance.ButtonPrimary)

	win.Compute(engine)
	win.Supervise(supervisor)

	return win
}
