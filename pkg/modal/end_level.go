package modal

import (
	"fmt"
	"time"

	"git.kirsle.net/apps/doodle/pkg/balance"
	"git.kirsle.net/apps/doodle/pkg/log"
	"git.kirsle.net/apps/doodle/pkg/savegame"
	"git.kirsle.net/apps/doodle/pkg/sprites"
	"git.kirsle.net/go/render"
	"git.kirsle.net/go/ui"
)

// ConfigEndLevel sets options for the EndLevel modal.
type ConfigEndLevel struct {
	Engine  render.Engine
	Success bool // false = failure condition

	// Handler functions - what you don't define will not
	// show as buttons in the modal.
	OnRestartLevel    func() // Restart Level
	OnRetryCheckpoint func() // Continue from checkpoint
	OnEditLevel       func()
	OnNextLevel       func() // Next Level
	OnExitToMenu      func() // Exit to Menu

	// Set these values to show the "New Record!" part of the modal.
	NewRecord   bool
	IsPerfect   bool
	TimeElapsed time.Duration
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

	// New Record frame.
	if cfg.NewRecord {
		// Get the gold or silver sprite.
		var (
			coin       = balance.SilverCoin
			recordFont = balance.NewRecordFont
		)
		if cfg.IsPerfect {
			coin = balance.GoldCoin
			recordFont = balance.NewRecordPerfectFont
		}

		recordFrame := ui.NewFrame("New Record")
		msgFrame.Pack(recordFrame, ui.Pack{
			Side: ui.N,
		})

		header := ui.NewLabel(ui.Label{
			Text: "A New Record!",
			Font: recordFont,
		})
		recordFrame.Pack(header, ui.Pack{
			Side:  ui.N,
			FillX: true,
		})

		// A frame to hold the icon and duration elapsed.
		timeFrame := ui.NewFrame("Time Frame")
		recordFrame.Pack(timeFrame, ui.Pack{
			Side: ui.N,
		})

		// Show the coin image.
		if cfg.Engine != nil {
			img, err := sprites.LoadImage(cfg.Engine, coin)
			if err != nil {
				log.Error("Couldn't load %s: %s", coin, err)
			} else {
				timeFrame.Pack(img, ui.Pack{
					Side: ui.W,
				})
			}
		}

		// Show the time duration label.
		dur := ui.NewLabel(ui.Label{
			Text: savegame.FormatDuration(cfg.TimeElapsed),
			Font: balance.MenuFont,
		})
		timeFrame.Pack(dur, ui.Pack{
			Side: ui.W,
		})
	}

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
