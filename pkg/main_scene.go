package doodle

import (
	"fmt"

	"git.kirsle.net/apps/doodle/pkg/balance"
	"git.kirsle.net/apps/doodle/pkg/branding"
	"git.kirsle.net/apps/doodle/pkg/level"
	"git.kirsle.net/apps/doodle/pkg/license"
	"git.kirsle.net/apps/doodle/pkg/log"
	"git.kirsle.net/apps/doodle/pkg/native"
	"git.kirsle.net/apps/doodle/pkg/scripting"
	"git.kirsle.net/apps/doodle/pkg/shmem"
	"git.kirsle.net/apps/doodle/pkg/uix"
	"git.kirsle.net/apps/doodle/pkg/updater"
	"git.kirsle.net/apps/doodle/pkg/windows"
	"git.kirsle.net/go/render"
	"git.kirsle.net/go/render/event"
	"git.kirsle.net/go/ui"
	"git.kirsle.net/go/ui/style"
)

// MainScene implements the main menu of Doodle.
type MainScene struct {
	Supervisor *ui.Supervisor

	// Background wallpaper canvas.
	scripting *scripting.Supervisor
	canvas    *uix.Canvas

	// UI components.
	labelTitle    *ui.Label
	labelSubtitle *ui.Label
	labelVersion  *ui.Label
	labelHint     *ui.Label
	frame         *ui.Frame // Main button frame
	btnRegister   *ui.Button
	winRegister   *ui.Window
	winSettings   *ui.Window

	// Update check variables.
	updateButton *ui.Button
	updateInfo   updater.VersionInfo

	// Lazy scroll variables. See LoopLazyScroll().
	lazyScrollBounce     bool
	lazyScrollTrajectory render.Point
	lazyScrollLastValue  render.Point
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

	// Subtitle/byline.
	s.labelSubtitle = ui.NewLabel(ui.Label{
		Text: branding.Byline,
		Font: balance.TitleScreenSubtitleFont,
	})
	s.labelSubtitle.Compute(d.Engine)

	// Version label.
	var shareware string
	if !license.IsRegistered() {
		shareware = " (shareware)"
	}
	ver := ui.NewLabel(ui.Label{
		Text: fmt.Sprintf("v%s%s", branding.Version, shareware),
		Font: balance.TitleScreenVersionFont,
	})
	ver.Compute(d.Engine)
	s.labelVersion = ver

	// Arrow Keys hint label (scroll the demo level).
	s.labelHint = ui.NewLabel(ui.Label{
		Text: "Hint: press the Arrow keys",
		Font: render.Text{
			Size:   16,
			Color:  render.Grey,
			Shadow: render.Purple,
		},
	})
	s.labelHint.Compute(d.Engine)

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

	// Register button.
	s.btnRegister = ui.NewButton("Register", ui.NewLabel(ui.Label{
		Text: "Register Game",
		Font: balance.LabelFont,
	}))
	s.btnRegister.SetStyle(&balance.ButtonPrimary)
	s.btnRegister.Handle(ui.Click, func(ed ui.EventData) error {
		if s.winRegister == nil {
			cfg := windows.License{
				Supervisor: s.Supervisor,
				Engine:     d.Engine,
				OnCancel: func() {
					s.winRegister.Hide()
				},
			}
			cfg.OnLicensed = func() {
				// License status has changed, reload the window!
				if s.winRegister != nil {
					s.winRegister.Hide()
				}
				s.winRegister = windows.MakeLicenseWindow(d.width, d.height, cfg)
			}

			cfg.OnLicensed()
		}
		s.winRegister.Show()
		return nil
	})
	s.btnRegister.Compute(d.Engine)
	s.Supervisor.Add(s.btnRegister)

	if license.IsRegistered() {
		s.btnRegister.Hide()
	}

	// Main UI button frame.
	frame := ui.NewFrame("frame")
	s.frame = frame

	var buttons = []struct {
		Name  string
		Func  func()
		Style *style.Button
	}{
		// {
		// 	Name: "Story Mode",
		// 	Func: d.GotoStoryMenu,
		// },
		{
			Name:  "Play a Level",
			Func:  d.GotoPlayMenu,
			Style: &balance.ButtonBabyBlue,
		},
		{
			Name:  "Create a Level",
			Func:  d.GotoNewMenu,
			Style: &balance.ButtonPink,
		},
		{
			Name: "Create a Doodad",
			Func: func() {
				d.NewDoodad(0)
			},
			Style: &balance.ButtonPink,
		},
		{
			Name:  "Edit a Drawing",
			Func:  d.GotoLoadMenu,
			Style: &balance.ButtonPrimary,
		},
		{
			Name: "Settings",
			Func: func() {
				if s.winSettings == nil {
					s.winSettings = d.MakeSettingsWindow(s.Supervisor)
				}
				s.winSettings.Show()
			},
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
		if button.Style != nil {
			btn.SetStyle(button.Style)
		}
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
	if shmem.OfflineMode {
		log.Info("OfflineMode: skip updates check")
		return
	}

	info, err := updater.Check()
	if err != nil {
		log.Error(err.Error())
		return
	}

	if info.IsNewerVersionThan(branding.Version) {
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
	if lvl, err := level.LoadFile(balance.DemoLevelName); err == nil {
		s.canvas.LoadLevel(lvl)
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
		log.Error("Error loading demo level %s: %s", balance.DemoLevelName, err)
	}

	return nil
}

// Loop the editor scene.
func (s *MainScene) Loop(d *Doodle, ev *event.State) error {
	s.Supervisor.Loop(ev)

	if err := s.scripting.Loop(); err != nil {
		log.Error("MainScene.Loop: scripting.Loop: %s", err)
	}

	// Lazily scroll the canvas around, slowly.
	s.LoopLazyScroll()

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

// LoopLazyScroll gently scrolls the title screen demo level, called each Loop.
func (s *MainScene) LoopLazyScroll() {
	// The v1 basic sauce algorithm:
	// 1. We scroll diagonally downwards and rightwards.
	// 2. When we scroll downwards far enough, we change direction.
	//    Make a zigzag pattern.
	// 3. When we reach the right bound of the level
	//    OR some max number of px into an unbounded level:
	//    enter a simple ball bouncing mode like a screensaver.
	var (
		zigzagMaxHeight = 512
		maxScrollX      = zigzagMaxHeight * 2
		lastScrollValue = s.lazyScrollLastValue
		currentScroll   = s.canvas.Scroll
	)

	// So we have two states:
	// - Zigzag state (default)
	// - Bounce state (when we hit a wall)
	if !s.lazyScrollBounce {
		// Zigzag state.
		s.lazyScrollTrajectory = render.Point{
			X: -1, // down and right
			Y: -1,
		}

		// When we've gone far enough X, it's also far enough Y.
		if currentScroll.X < -zigzagMaxHeight {
			s.lazyScrollTrajectory.Y = 1 // go back up
		}

		// Have we gotten stuck in a corner? (ending the zigzag phase, for bounded levels)
		if currentScroll.X < 0 && (currentScroll == lastScrollValue) || currentScroll.X < -maxScrollX {
			log.Debug("LoopLazyScroll: Ending zigzag phase, enter bounce phase")
			s.lazyScrollBounce = true
			s.lazyScrollTrajectory = render.Point{
				X: -1,
				Y: -1,
			}
		}
	} else {
		// Lazy bounce algorithm.
		if currentScroll.Y == lastScrollValue.Y {
			log.Debug("LoopLazyScroll: Hit a floor/ceiling")
			s.lazyScrollTrajectory.Y = -s.lazyScrollTrajectory.Y
		}
		if currentScroll.X == lastScrollValue.X {
			log.Debug("LoopLazyScroll: Hit the side of the map!")
			s.lazyScrollTrajectory.X = -s.lazyScrollTrajectory.X
		}
	}

	// Check the scroll.
	s.lazyScrollLastValue = currentScroll
	s.canvas.ScrollBy(s.lazyScrollTrajectory)
}

// Draw the pixels on this frame.
func (s *MainScene) Draw(d *Doodle) error {
	// Clear the canvas and fill it with white.
	d.Engine.Clear(render.White)

	s.canvas.Present(d.Engine, render.Origin)

	// Draw a sheen over the level for clarity.
	d.Engine.DrawBox(render.RGBA(255, 255, 254, 96), render.Rect{
		X: 0,
		Y: 0,
		W: d.width,
		H: d.height,
	})

	// Draw out bounding boxes.
	if DebugCollision {
		for _, actor := range s.canvas.Actors() {
			d.DrawCollisionBox(s.canvas, actor)
		}
	}

	// App title label.
	s.labelTitle.MoveTo(render.Point{
		X: (d.width / 2) - (s.labelTitle.Size().W / 2),
		Y: 120,
	})
	s.labelTitle.Present(d.Engine, s.labelTitle.Point())

	// App subtitle label (byline).
	s.labelSubtitle.MoveTo(render.Point{
		X: (d.width / 2) - (s.labelSubtitle.Size().W / 2),
		Y: s.labelTitle.Point().Y + s.labelTitle.Size().H + 8,
	})
	s.labelSubtitle.Present(d.Engine, s.labelSubtitle.Point())

	// Version label
	s.labelVersion.MoveTo(render.Point{
		X: (d.width) - (s.labelVersion.Size().W) - 20,
		Y: 20,
	})
	s.labelVersion.Present(d.Engine, s.labelVersion.Point())

	// Hint label.
	s.labelHint.MoveTo(render.Point{
		X: (d.width / 2) - (s.labelHint.Size().W / 2),
		Y: d.height - s.labelHint.Size().H - 32,
	})
	s.labelHint.Present(d.Engine, s.labelHint.Point())

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

	// Register button.
	s.btnRegister.MoveTo(render.Point{
		X: d.width - s.btnRegister.Size().W - 24,
		Y: d.height - s.btnRegister.Size().H - 24,
	})
	s.btnRegister.Present(d.Engine, s.btnRegister.Point())

	// Present supervised windows.
	s.Supervisor.Present(d.Engine)

	return nil
}

// Destroy the scene.
func (s *MainScene) Destroy() error {
	return nil
}
