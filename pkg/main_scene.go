package doodle

import (
	"fmt"
	"math/rand"

	"git.kirsle.net/apps/doodle/pkg/balance"
	"git.kirsle.net/apps/doodle/pkg/branding"
	"git.kirsle.net/apps/doodle/pkg/level"
	"git.kirsle.net/apps/doodle/pkg/levelpack"
	"git.kirsle.net/apps/doodle/pkg/license"
	"git.kirsle.net/apps/doodle/pkg/log"
	"git.kirsle.net/apps/doodle/pkg/modal/loadscreen"
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
	Supervisor    *ui.Supervisor
	LevelFilename string // custom level filename to load in background

	// Background wallpaper canvas.
	scripting *scripting.Supervisor
	canvas    *uix.Canvas

	// UI components.
	labelTitle    *ui.Label
	labelSubtitle *ui.Label
	labelVersion  *ui.Label
	labelHint     *ui.Label
	frame         *ui.Frame // Main button frame
	winRegister   *ui.Window
	winSettings   *ui.Window
	winLevelPacks *ui.Window

	// Update check variables.
	updateButton *ui.Button
	updateInfo   updater.VersionInfo

	// Lazy scroll variables. See LoopLazyScroll().
	lazyScrollBounce     bool
	lazyScrollTrajectory render.Point
	lazyScrollLastValue  render.Point

	// Landscape mode: if the screen isn't tall enough to see the main
	// menu we redo the layout to be landscape friendly. NOTE: this only
	// happens one time, and does not re-adapt when the window is made
	// tall enough again.
	landscapeMode bool
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

	// Main UI button frame.
	frame := ui.NewFrame("frame")
	s.frame = frame

	var buttons = []struct {
		Name  string
		If    func() bool
		Func  func()
		Style *style.Button
	}{
		{
			Name: "Story Mode",
			Func: func() {
				if s.winLevelPacks == nil {
					s.winLevelPacks = windows.NewLevelPackWindow(windows.LevelPack{
						Supervisor: s.Supervisor,
						Engine:     d.Engine,

						OnPlayLevel: func(lp levelpack.LevelPack, which levelpack.Level) {
							if err := d.PlayFromLevelpack(lp, which); err != nil {
								shmem.FlashError(err.Error())
							}
						},
						OnCloseWindow: func() {
							s.winLevelPacks.Destroy()
							s.winLevelPacks = nil
						},
					})
				}
				s.winLevelPacks.MoveTo(render.Point{
					X: (d.width / 2) - (s.winLevelPacks.Size().W / 2),
					Y: (d.height / 2) - (s.winLevelPacks.Size().H / 2),
				})
				s.winLevelPacks.Show()
			},
			Style: &balance.ButtonBabyBlue,
		},
		{
			Name:  "Play a Level",
			Func:  d.GotoPlayMenu,
			Style: &balance.ButtonBabyBlue,
		},
		{
			Name:  "New Drawing",
			Func:  d.GotoNewMenu,
			Style: &balance.ButtonPink,
		},
		{
			Name:  "Edit Drawing",
			Func:  d.GotoLoadMenu,
			Style: &balance.ButtonPink,
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
		{
			Name: "Register",
			If: func() bool {
				return !license.IsRegistered()
			},
			Func: func() {
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
			},
			Style: &balance.ButtonPrimary,
		},
	}
	for _, button := range buttons {
		if check := button.If; check != nil && !check() {
			continue
		}

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

	// Eager load the level in background, no time for load screen.
	go func() {
		if err := s.setupAsync(d); err != nil {
			log.Error("MainScene.setupAsync: %s", err)
		}
	}()

	// Trigger our "Window Resized" function so we can check if the
	// layout needs to be switched to landscape mode for mobile.
	s.Resized(d.width, d.height)

	return nil
}

// setupAsync runs background tasks from setup, e.g. eager load
// chunks of the level for cache.
func (s *MainScene) setupAsync(d *Doodle) error {
	loadscreen.PreloadAllChunkBitmaps(s.canvas.Chunker())
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

	// Title screen level to load. Pick a random level.
	levelName := balance.DemoLevelName[0]
	if s.LevelFilename != "" {
		levelName = s.LevelFilename
	} else if len(balance.DemoLevelName) > 1 {
		randIndex := rand.Intn(len(balance.DemoLevelName))
		levelName = balance.DemoLevelName[randIndex]
	}

	if lvl, err := level.LoadFile(levelName); err == nil {
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

		// Create a basic notebook level.
		s.canvas.LoadLevel(&level.Level{
			Chunker:   level.NewChunker(100),
			Palette:   level.NewPalette(),
			PageType:  level.Bounded,
			MaxWidth:  42,
			MaxHeight: 42,
			Wallpaper: "notebook.png",
		})
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
		s.Resized(w, h)
	}

	return nil
}

// Resized the app window.
func (s *MainScene) Resized(width, height int) {
	log.Info("Resized to %dx%d", width, height)

	// If the height is not tall enough for the menu, switch to the horizontal layout.
	if height < balance.TitleScreenResponsiveHeight {
		log.Error("Switch to landscape mode")
		s.landscapeMode = true
	} else {
		s.landscapeMode = false
	}

	s.canvas.Resize(render.Rect{
		W: width,
		H: height,
	})
}

// Move things into position for the main menu. This function arranges
// the Title, Subtitle, Buttons, etc. into screen relative positions every
// tick. This function sets their 'default' values, but if the window is
// not tall enough and needs the landscape orientation, positionMenuLandscape()
// will override these defaults.
func (s *MainScene) positionMenuPortrait(d *Doodle) {
	// App title label.
	s.labelTitle.MoveTo(render.Point{
		X: (d.width / 2) - (s.labelTitle.Size().W / 2),
		Y: 120,
	})

	// App subtitle label (byline).
	s.labelSubtitle.MoveTo(render.Point{
		X: (d.width / 2) - (s.labelSubtitle.Size().W / 2),
		Y: s.labelTitle.Point().Y + s.labelTitle.Size().H + 8,
	})

	// Version label
	s.labelVersion.MoveTo(render.Point{
		X: (d.width) - (s.labelVersion.Size().W) - 20,
		Y: 20,
	})

	// Hint label.
	s.labelHint.MoveTo(render.Point{
		X: (d.width / 2) - (s.labelHint.Size().W / 2),
		Y: d.height - s.labelHint.Size().H - 32,
	})

	// Update button.
	s.updateButton.MoveTo(render.Point{
		X: 24,
		Y: d.height - s.updateButton.Size().H - 24,
	})

	// Button frame.
	s.frame.MoveTo(render.Point{
		X: (d.width / 2) - (s.frame.Size().W / 2),
		Y: 260,
	})
}

func (s *MainScene) positionMenuLandscape(d *Doodle) {
	s.positionMenuPortrait(d)

	var (
		col1 = render.Rect{
			X: 0,
			Y: 0,
			W: d.width / 2,
			H: d.height,
		}
		col2 = render.Rect{
			X: d.width,
			Y: 0,
			W: d.width - col1.W,
			H: d.height,
		}
	)

	// Title and subtitle move to the left.
	s.labelTitle.MoveTo(render.Point{
		X: (col1.W / 2) - (s.labelTitle.Size().W / 2),
		Y: s.labelTitle.Point().Y,
	})
	s.labelSubtitle.MoveTo(render.Point{
		X: (col1.W / 2) - (s.labelSubtitle.Size().W / 2),
		Y: s.labelTitle.Point().Y + s.labelTitle.Size().H + 8,
	})

	// Button frame to the right.
	s.frame.MoveTo(render.Point{
		X: (col2.X+col2.W)/2 - (s.frame.Size().W / 2),
		Y: (d.height / 2) - (s.frame.Size().H / 2),
	})
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

	// Arrange the main widgets by Portrait or Landscape mode.
	if s.landscapeMode {
		s.positionMenuLandscape(d)
	} else {
		s.positionMenuPortrait(d)
	}

	// App title label.
	s.labelTitle.Present(d.Engine, s.labelTitle.Point())

	// App subtitle label (byline).
	s.labelSubtitle.Present(d.Engine, s.labelSubtitle.Point())

	// Version label
	s.labelVersion.Present(d.Engine, s.labelVersion.Point())

	// Hint label.
	s.labelHint.Present(d.Engine, s.labelHint.Point())

	// Update button.
	s.updateButton.Present(d.Engine, s.updateButton.Point())

	s.frame.Compute(d.Engine)
	s.frame.Present(d.Engine, s.frame.Point())

	// Present supervised windows.
	s.Supervisor.Present(d.Engine)

	return nil
}

// Destroy the scene.
func (s *MainScene) Destroy() error {
	return nil
}
