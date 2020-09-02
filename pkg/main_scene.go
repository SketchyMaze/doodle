package doodle

import (
	"fmt"

	"git.kirsle.net/apps/doodle/pkg/balance"
	"git.kirsle.net/apps/doodle/pkg/branding"
	"git.kirsle.net/apps/doodle/pkg/level"
	"git.kirsle.net/apps/doodle/pkg/log"
	"git.kirsle.net/apps/doodle/pkg/native"
	"git.kirsle.net/apps/doodle/pkg/scripting"
	"git.kirsle.net/apps/doodle/pkg/uix"
	"git.kirsle.net/apps/doodle/pkg/updater"
	"git.kirsle.net/go/render"
	"git.kirsle.net/go/render/event"
	"git.kirsle.net/go/ui"
)

// MainScene implements the main menu of Doodle.
type MainScene struct {
	Supervisor *ui.Supervisor

	// Background wallpaper canvas.
	scripting *scripting.Supervisor
	canvas    *uix.Canvas

	// UI components.
	labelTitle   *ui.Label
	labelVersion *ui.Label
	frame        *ui.Frame // Main button frame

	// Update check variables.
	updateButton *ui.Button
	updateInfo   updater.VersionInfo
}

// Name of the scene.
func (s *MainScene) Name() string {
	return "Main"
}

// Setup the scene.
func (s *MainScene) Setup(d *Doodle) error {
	s.Supervisor = ui.NewSupervisor()

	if err := s.SetupDemoLevel(d); err != nil {
		return err
	}

	// Main title label
	s.labelTitle = ui.NewLabel(ui.Label{
		Text: branding.AppName,
		Font: balance.TitleScreenFont,
	})
	s.labelTitle.Compute(d.Engine)

	// Version label.
	var shareware string
	if balance.FreeVersion {
		shareware = " (shareware)"
	}
	ver := ui.NewLabel(ui.Label{
		Text: fmt.Sprintf("v%s%s", branding.Version, shareware),
		Font: render.Text{
			Size:   18,
			Color:  render.Grey,
			Shadow: render.Black,
		},
	})
	ver.Compute(d.Engine)
	s.labelVersion = ver

	// "Update Available" button.
	s.updateButton = ui.NewButton("Update Button", ui.NewLabel(ui.Label{
		Text: "An update is available!",
		Font: render.Text{
			FontFilename: "DejaVuSans-Bold.ttf",
			Size:         16,
			Color:        render.Blue,
			Padding:      4,
		},
	}))
	s.updateButton.Handle(ui.Click, func(ed ui.EventData) error {
		native.OpenURL(s.updateInfo.DownloadURL)
		return nil
	})
	s.updateButton.Compute(d.Engine)
	s.updateButton.Hide()
	s.Supervisor.Add(s.updateButton)

	// Main UI button frame.
	frame := ui.NewFrame("frame")
	s.frame = frame

	var buttons = []struct {
		Name string
		Func func()
	}{
		// {
		// 	Name: "Story Mode",
		// 	Func: d.GotoStoryMenu,
		// },
		{
			Name: "Play a Level",
			Func: d.GotoPlayMenu,
		},
		{
			Name: "Create a New Level",
			Func: d.GotoNewMenu,
		},
		{
			Name: "Edit a Level",
			Func: d.GotoLoadMenu,
		},
		{
			Name: "Settings",
			Func: d.GotoSettingsMenu,
		},
	}
	for _, button := range buttons {
		button := button
		btn := ui.NewButton(button.Name, ui.NewLabel(ui.Label{
			Text: button.Name,
			Font: balance.StatusFont,
		}))
		btn.Handle(ui.Click, func(ed ui.EventData) error {
			button.Func()
			return nil
		})
		s.Supervisor.Add(btn)
		frame.Pack(btn, ui.Pack{
			Side: ui.N,
			PadY: 8,
			// Fill:   true,
			FillX: true,
		})
	}

	// Check for update in the background.
	go s.checkUpdate()

	return nil
}

// checkUpdate checks for a version update and shows the button.
func (s *MainScene) checkUpdate() {
	info, err := updater.Check()
	if err != nil {
		log.Error(err.Error())
		return
	}

	if info.LatestVersion != branding.Version {
		s.updateInfo = info
		s.updateButton.Show()
	}
}

// SetupDemoLevel configures the wallpaper behind the New screen,
// which demos a title screen demo level.
func (s *MainScene) SetupDemoLevel(d *Doodle) error {
	// Set up the background wallpaper canvas.
	s.canvas = uix.NewCanvas(100, false)
	s.canvas.Scrollable = true
	s.canvas.Resize(render.Rect{
		W: d.width,
		H: d.height,
	})

	s.scripting = scripting.NewSupervisor()
	s.canvas.SetScriptSupervisor(s.scripting)

	// Title screen level to load.
	if lvl, err := level.LoadFile("example1.level"); err == nil {
		s.canvas.LoadLevel(d.Engine, lvl)
		s.canvas.InstallActors(lvl.Actors)

		// Load all actor scripts.
		if err := s.scripting.InstallScripts(lvl); err != nil {
			log.Error("Error with title screen level scripts: %s", err)
		}

		// Run all actors scripts main function to start them off.
		if err := s.canvas.InstallScripts(); err != nil {
			log.Error("Error running actor main() functions: %s", err)
		}
	} else {
		log.Error("Error loading title-screen.level: %s", err)
	}

	return nil
}

// Loop the editor scene.
func (s *MainScene) Loop(d *Doodle, ev *event.State) error {
	s.Supervisor.Loop(ev)

	if err := s.scripting.Loop(); err != nil {
		log.Error("MainScene.Loop: scripting.Loop: %s", err)
	}

	s.canvas.Loop(ev)

	if ev.WindowResized {
		w, h := d.Engine.WindowSize()
		d.width = w
		d.height = h
		log.Info("Resized to %dx%d", d.width, d.height)
		s.canvas.Resize(render.Rect{
			W: d.width,
			H: d.height,
		})
	}

	return nil
}

// Draw the pixels on this frame.
func (s *MainScene) Draw(d *Doodle) error {
	// Clear the canvas and fill it with white.
	d.Engine.Clear(render.White)

	s.canvas.Present(d.Engine, render.Origin)

	// Draw a sheen over the level for clarity.
	d.Engine.DrawBox(render.RGBA(255, 255, 254, 128), render.Rect{
		X: 0,
		Y: 0,
		W: d.width,
		H: d.height,
	})

	// App title label.
	s.labelTitle.MoveTo(render.Point{
		X: (d.width / 2) - (s.labelTitle.Size().W / 2),
		Y: 120,
	})
	s.labelTitle.Present(d.Engine, s.labelTitle.Point())

	// Version label
	s.labelVersion.MoveTo(render.Point{
		X: (d.width / 2) - (s.labelVersion.Size().W / 2),
		Y: s.labelTitle.Point().Y + s.labelTitle.Size().H + 8,
	})
	s.labelVersion.Present(d.Engine, s.labelVersion.Point())

	// Update button.
	s.updateButton.MoveTo(render.Point{
		X: 24,
		Y: d.height - s.updateButton.Size().H - 24,
	})
	s.updateButton.Present(d.Engine, s.updateButton.Point())

	s.frame.Compute(d.Engine)
	s.frame.MoveTo(render.Point{
		X: (d.width / 2) - (s.frame.Size().W / 2),
		Y: 260,
	})
	s.frame.Present(d.Engine, s.frame.Point())

	return nil
}

// Destroy the scene.
func (s *MainScene) Destroy() error {
	return nil
}
