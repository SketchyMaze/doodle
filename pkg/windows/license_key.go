package windows

import (
	"fmt"
	"time"

	"git.kirsle.net/SketchyMaze/doodle/pkg/balance"
	"git.kirsle.net/SketchyMaze/doodle/pkg/branding"
	"git.kirsle.net/SketchyMaze/doodle/pkg/branding/builds"
	"git.kirsle.net/SketchyMaze/doodle/pkg/log"
	"git.kirsle.net/SketchyMaze/doodle/pkg/modal"
	"git.kirsle.net/SketchyMaze/doodle/pkg/native"
	"git.kirsle.net/SketchyMaze/doodle/pkg/plus"
	"git.kirsle.net/SketchyMaze/doodle/pkg/plus/dpp"
	"git.kirsle.net/go/render"
	"git.kirsle.net/go/ui"
	"git.kirsle.net/go/ui/style"
)

// License window.
type License struct {
	// Settings passed in by doodle
	Supervisor *ui.Supervisor
	Engine     render.Engine

	OnLicensed func()
	OnCancel   func()
}

// MakeLicenseWindow initializes a license window for any scene.
// The window width/height are the actual SDL2 window dimensions.
func MakeLicenseWindow(windowWidth, windowHeight int, cfg License) *ui.Window {
	win := NewLicenseWindow(cfg)
	win.Compute(cfg.Engine)
	win.Supervise(cfg.Supervisor)

	// Center the window.
	size := win.Size()
	win.MoveTo(render.Point{
		X: (windowWidth / 2) - (size.W / 2),
		Y: (windowHeight / 2) - (size.H / 2),
	})

	return win
}

// NewLicenseWindow initializes the window.
func NewLicenseWindow(cfg License) *ui.Window {
	var (
		windowWidth  = 340
		windowHeight = 320
		labelSize    = render.NewRect(100, 16)
		valueSize    = render.NewRect(windowWidth-labelSize.W-4, labelSize.H)
		isRegistered bool
		registration plus.Registration
		summary      = "Unregistered" + builds.VersionSuffix
	)

	// Get our current registration status.
	if reg, err := dpp.Driver.GetRegistration(); err == nil {
		isRegistered = true
		registration = reg
		windowHeight = 200
		summary = "Registered"
	}

	window := ui.NewWindow("Registration")
	window.SetButtons(ui.CloseButton)
	window.Configure(ui.Config{
		Width:      windowWidth,
		Height:     windowHeight,
		Background: render.RGBA(255, 200, 255, 255),
	})

	var rows = []struct {
		IfRegistered   bool
		IfUnregistered bool

		Label       string
		Text        string
		Button      *ui.Button
		ButtonStyle *style.Button
		Func        func()
		PadY        int
		PadX        int
	}{
		{
			Label: "Version:",
			Text:  fmt.Sprintf("%s v%s", branding.AppName, branding.Version),
		},
		{
			Label: "Status:",
			Text:  summary,
		},
		{
			IfRegistered: true,
			Label:        "Name:",
			Text:         registration.Name,
		},
		{
			IfRegistered: true,
			Label:        "Email:",
			Text:         registration.Email,
		},
		{
			IfRegistered: true,
			Label:        "Issued:",
			Text:         time.Unix(registration.IssuedAt, 0).Format("Jan 2, 2006 15:04:05 MST"),
		},
		{
			IfUnregistered: true,
			Text: "Register your game today! By purchasing the full\n" +
				"version of Sketchy Maze, you will unlock additional\n" +
				"features including improved support for custom\n" +
				"doodads that attach with your level files for easy\n" +
				"sharing between multiple computers.\n\n" +
				"When you purchase the game you will receive a\n" +
				"license key file; click the button below to browse\n" +
				"and select the key file to register this copy of\n" +
				branding.AppName + ".",
			PadY: 8,
			PadX: 2,
		},
		{
			IfRegistered: true,
			Text:         "Thank you for your support!",
			PadY:         8,
			PadX:         2,
		},
		{
			IfUnregistered: true,
			Button: ui.NewButton("Key Browse", ui.NewLabel(ui.Label{
				Text: "Browse for License Key",
				Font: balance.UIFont,
			})),
			ButtonStyle: &balance.ButtonPrimary,
			Func: func() {
				filename, err := native.OpenFile("Select License File", "*.key *.txt")
				if err != nil {
					log.Error(err.Error())
					return
				}

				// Upload and validate the license key.
				reg, err := dpp.Driver.UploadLicenseFile(filename)
				if err != nil {
					modal.Alert("That license key didn't seem quite right.").WithTitle("License Error")
					return
				}

				modal.Alert("Thank you, %s!", reg.Name).WithTitle("Registration OK!")
				if cfg.OnLicensed != nil {
					cfg.OnLicensed()
				}
			},
		},
	}
	for _, row := range rows {
		row := row

		// It has a conditional?
		if (row.IfRegistered && !isRegistered) ||
			(row.IfUnregistered && isRegistered) {
			continue
		}

		frame := ui.NewFrame("Frame")
		if row.Label != "" {
			lf := ui.NewFrame("LabelFrame")
			lf.Resize(labelSize)
			label := ui.NewLabel(ui.Label{
				Text: row.Label,
				Font: balance.LabelFont,
			})
			lf.Pack(label, ui.Pack{
				Side: ui.E,
			})
			frame.Pack(lf, ui.Pack{
				Side: ui.W,
			})
		}

		if row.Text != "" {
			tf := ui.NewFrame("TextFrame")
			if row.Label != "" {
				tf.Resize(valueSize)
			}
			label := ui.NewLabel(ui.Label{
				Text: row.Text,
				Font: balance.UIFont,
			})
			tf.Pack(label, ui.Pack{
				Side: ui.W,
			})
			frame.Pack(tf, ui.Pack{
				Side: ui.W,
			})
		}

		if row.Button != nil {
			btn := row.Button
			if row.ButtonStyle != nil {
				btn.SetStyle(row.ButtonStyle)
			}
			btn.Handle(ui.Click, func(ed ui.EventData) error {
				if row.Func != nil {
					row.Func()
				}
				return nil
			})
			btn.Compute(cfg.Engine)
			cfg.Supervisor.Add(btn)
			frame.Pack(btn, ui.Pack{
				Side: ui.N,
			})
		}
		window.Pack(frame, ui.Pack{
			Side:  ui.N,
			FillX: true,
			PadY:  row.PadY,
			PadX:  row.PadX,
		})
	}

	/////////////
	// Buttons at bottom of window

	bottomFrame := ui.NewFrame("Button Frame")
	window.Pack(bottomFrame, ui.Pack{
		Side:  ui.N,
		FillX: true,
	})

	frame := ui.NewFrame("Button frame")
	buttons := []struct {
		label string
		f     func()
	}{
		{"Website", func() {
			native.OpenURL(branding.Website)
		}},
		{"Close", func() {
			if cfg.OnCancel != nil {
				cfg.OnCancel()
			}
		}},
	}
	for _, button := range buttons {
		button := button

		btn := ui.NewButton(button.label, ui.NewLabel(ui.Label{
			Text: button.label,
			Font: balance.MenuFont,
		}))

		btn.Handle(ui.Click, func(ed ui.EventData) error {
			button.f()
			return nil
		})

		btn.Compute(cfg.Engine)
		cfg.Supervisor.Add(btn)

		frame.Pack(btn, ui.Pack{
			Side:   ui.W,
			PadX:   4,
			Expand: true,
			Fill:   true,
		})
	}
	bottomFrame.Pack(frame, ui.Pack{
		Side: ui.N,
		PadX: 8,
		PadY: 12,
	})

	return window
}
