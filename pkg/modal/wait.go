package modal

import (
	"fmt"
	"time"

	"git.kirsle.net/SketchyMaze/doodle/pkg/balance"
	"git.kirsle.net/go/render"
	"git.kirsle.net/go/ui"
)

// Wait pops up a non-dismissable modal that the caller can close when they're ready.
func Wait(message string, args ...interface{}) *Modal {
	if !ready {
		panic("modal.Wait(): not ready")
	} else if current != nil {
		current.Dismiss(false)
	}

	// Reset the supervisor.
	supervisor = ui.NewSupervisor()

	m := &Modal{
		title:   "Wait",
		message: fmt.Sprintf(message, args...),
		force:   true,
	}
	m.window = makeWaitModal(m)

	center(m.window)
	current = m

	return m
}

// creates the ui.Window for the Wait modal.
func makeWaitModal(m *Modal) *ui.Window {
	win := ui.NewWindow("Wait")
	_, title := win.TitleBar()
	title.TextVariable = &m.title

	msgFrame := ui.NewFrame("Wait Message")
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

	// Create a bouncing progress bar.
	var (
		trough        *ui.Frame
		troughW       = 250
		progressBar   *ui.Frame
		progressX     int
		progressW     = 64
		progressH     = 30
		progressSpeed = 8
		progressFreq  = 16 * time.Millisecond
	)

	trough = ui.NewFrame("Progress Trough")
	trough.Configure(ui.Config{
		Width:       troughW,
		Height:      progressH,
		BorderSize:  1,
		BorderStyle: ui.BorderSunken,
		Background:  render.Grey,
	})
	win.Pack(trough, ui.Pack{
		Side:    ui.N,
		Padding: 4,
	})

	progressBar = ui.NewFrame("Progress Bar")
	progressBar.Configure(ui.Config{
		Width:      progressW,
		Height:     30,
		Background: render.Green,
	})
	trough.Place(progressBar, ui.Place{
		Left: progressX,
		Top:  0,
	})

	trough.Compute(engine)

	win.Compute(engine)
	win.Supervise(supervisor)

	// Animate the bouncing of the progress bar in a background goroutine,
	// and allow canceling it when the modal is dismissed.
	var (
		cancel = make(chan interface{})
		ping   = time.NewTicker(progressFreq)
	)
	go func() {
		for {
			select {
			case <-cancel:
				ping.Stop()
				return
			case <-ping.C:
				// Have room to move the progress bar?
				progressX += progressSpeed

				// Cap it to within bounds.
				if progressX+progressW >= troughW {
					progressX = troughW - progressW
					if progressSpeed > 0 {
						progressSpeed *= -1
					}
				} else if progressX < 0 {
					progressX = 0
					if progressSpeed < 0 {
						progressSpeed *= -1
					}
				}

				trough.Place(progressBar, ui.Place{
					Left: progressX,
					Top:  0,
				})
				trough.Compute(engine)
			}
		}
	}()

	// Cancel the goroutine on modal teardown.
	m.teardown = func() {
		cancel <- nil
	}

	return win
}
