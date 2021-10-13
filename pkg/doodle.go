package doodle

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"git.kirsle.net/apps/doodle/pkg/balance"
	"git.kirsle.net/apps/doodle/pkg/branding"
	"git.kirsle.net/apps/doodle/pkg/enum"
	"git.kirsle.net/apps/doodle/pkg/keybind"
	"git.kirsle.net/apps/doodle/pkg/log"
	"git.kirsle.net/apps/doodle/pkg/modal"
	"git.kirsle.net/apps/doodle/pkg/modal/loadscreen"
	"git.kirsle.net/apps/doodle/pkg/native"
	"git.kirsle.net/apps/doodle/pkg/pattern"
	"git.kirsle.net/apps/doodle/pkg/shmem"
	"git.kirsle.net/apps/doodle/pkg/usercfg"
	"git.kirsle.net/apps/doodle/pkg/windows"
	golog "git.kirsle.net/go/log"
	"git.kirsle.net/go/render"
	"git.kirsle.net/go/render/event"
	"git.kirsle.net/go/ui"
)

const (
	// TargetFPS is the frame rate to cap the game to.
	TargetFPS = 1000 / 60 // 60 FPS

	// Millisecond64 is a time.Millisecond casted to float64.
	Millisecond64 = float64(time.Millisecond)
)

// Doodle is the game object.
type Doodle struct {
	Debug       bool
	Engine      render.Engine
	engineReady bool

	// Easy access to the event state, for the debug overlay to use.
	// Might not be thread safe.
	event *event.State

	startTime time.Time
	running   bool
	width     int
	height    int

	// Command line shell options.
	shell Shell

	Scene Scene
}

// New initializes the game object.
func New(debug bool, engine render.Engine) *Doodle {
	d := &Doodle{
		Debug:     debug,
		Engine:    engine,
		startTime: time.Now(),
		running:   true,
		width:     balance.Width,
		height:    balance.Height,
	}
	d.shell = NewShell(d)

	// Make the render engine globally available. TODO: for wasm/ToBitmap
	shmem.CurrentRenderEngine = engine
	shmem.Flash = d.Flash
	shmem.FlashError = d.FlashError
	shmem.Prompt = d.Prompt

	if debug {
		log.Logger.Config.Level = golog.DebugLevel
		// DebugOverlay = true // on by default in debug mode, F3 to disable
	}

	return d
}

// SetWindowSize sets the size of the Doodle window.
func (d *Doodle) SetWindowSize(width, height int) {
	d.width = width
	d.height = height
}

// Title returns the game's preferred window title.
func (d *Doodle) Title() string {
	return fmt.Sprintf("%s v%s", branding.AppName, branding.Version)
}

// SetupEngine sets up the rendering engine.
func (d *Doodle) SetupEngine() error {
	// Set up the rendering engine (SDL2, etc.)
	if err := d.Engine.Setup(); err != nil {
		return err
	}
	d.engineReady = true

	// Initialize the UI modal manager.
	modal.Initialize(d.Engine)

	// Preload the builtin brush patterns.
	pattern.LoadBuiltins(d.Engine)

	return nil
}

// Run initializes SDL and starts the main loop.
func (d *Doodle) Run() error {
	if !d.engineReady {
		if err := d.SetupEngine(); err != nil {
			return err
		}
	}

	// Set up the default scene.
	if d.Scene == nil {
		d.Goto(&MainScene{})
	}

	log.Info("Enter Main Loop")
	for d.running {
		// d.Engine.Clear(render.White)

		start := time.Now() // Record how long this frame took.
		shmem.Tick++

		// Poll for events.
		ev, err := d.Engine.Poll()
		shmem.Cursor = render.NewPoint(ev.CursorX, ev.CursorY)
		if err != nil {
			log.Error("event poll error: %s", err)
			d.running = false
			break
		}
		d.event = ev

		// Command line shell.
		if d.shell.Open {
		} else if keybind.ShellKey(ev) {
			log.Debug("Shell: opening shell")
			d.shell.Open = true
		} else {
			// Global event handlers.
			if keybind.Shutdown(ev) {
				if d.Debug { // fast exit in -debug mode.
					d.running = false
				} else {
					d.ConfirmExit()
				}
				continue
			}

			if keybind.Help(ev) {
				// Launch the local guidebook
				native.OpenLocalURL(balance.GuidebookPath)
			} else if keybind.DebugOverlay(ev) {
				DebugOverlay = !DebugOverlay
			} else if keybind.DebugCollision(ev) {
				DebugCollision = !DebugCollision
			}

			// Make sure no UI modals (alerts, confirms)
			// or loadscreen are currently visible.
			if !modal.Handled(ev) {
				// Run the scene's logic.
				err = d.Scene.Loop(d, ev)
				if err != nil {
					return err
				}
			}

		}

		// Draw the scene.
		d.Scene.Draw(d)

		// Draw the loadscreen if it is active.
		loadscreen.Loop(render.NewRect(d.width, d.height), d.Engine)

		// Draw modals on top of the game UI.
		modal.Draw()

		// Draw the shell, always on top of UI and modals.
		err = d.shell.Draw(d, ev)
		if err != nil {
			log.Error("shell error: %s", err)
			d.running = false
			break
		}

		// Draw the debug overlay over all scenes.
		d.DrawDebugOverlay()

		// Render the pixels to the screen.
		err = d.Engine.Present()
		if err != nil {
			log.Error("draw error: %s", err)
			d.running = false
			break
		}

		// Delay to maintain the target frames per second.
		var delay uint32
		if !fpsDoNotCap {
			elapsed := time.Now().Sub(start)
			tmp := elapsed / time.Millisecond
			if TargetFPS-int(tmp) > 0 { // make sure it won't roll under
				delay = uint32(TargetFPS - int(tmp))
			}
			d.Engine.Delay(delay)
		}

		// Track how long this frame took to measure FPS over time.
		d.TrackFPS(delay)

		// Consume any lingering key sym.
		// ev.ResetKeyDown()
	}

	log.Warn("Main Loop Exited! Shutting down...")
	return nil
}

// MakeSettingsWindow initializes the windows/settings.go window
// from anywhere you need it, binding all the variables in.
func (d *Doodle) MakeSettingsWindow(supervisor *ui.Supervisor) *ui.Window {
	cfg := windows.Settings{
		Supervisor: supervisor,
		Engine:     d.Engine,
		SceneName:  d.Scene.Name(),
		OnApply: func() {

		},

		// Boolean checkbox bindings
		DebugOverlay:       &DebugOverlay,
		DebugCollision:     &DebugCollision,
		HorizontalToolbars: &usercfg.Current.HorizontalToolbars,
		EnableFeatures:     &usercfg.Current.EnableFeatures,
		CrosshairSize:      &usercfg.Current.CrosshairSize,
		CrosshairColor:     &usercfg.Current.CrosshairColor,
		HideTouchHints:     &usercfg.Current.HideTouchHints,
	}
	return windows.MakeSettingsWindow(d.width, d.height, cfg)
}

// ConfirmExit may shut down Doodle gracefully after showing the user a
// confirmation modal.
func (d *Doodle) ConfirmExit() {
	modal.Confirm("Are you sure you want to quit %s?", branding.AppName).
		WithTitle("Confirm Quit").Then(func() {
		d.running = false
	})
}

// NewMap loads a new map in Edit Mode.
func (d *Doodle) NewMap() {
	log.Info("Starting a new map")
	scene := &EditorScene{}
	d.Goto(scene)
}

// NewDoodad loads a new Doodad in Edit Mode.
// If size is zero, it prompts the user to select a size or accept the default size.
func (d *Doodle) NewDoodad(size int) {
	if size == 0 {
		d.Prompt(fmt.Sprintf("Doodad size or %d>", balance.DoodadSize), func(answer string) {
			size := balance.DoodadSize
			if answer != "" {
				i, err := strconv.Atoi(answer)
				if err != nil {
					d.FlashError("Error: Doodad size must be a number.")
					return
				}
				size = i
			}

			// Recurse with the proper answer.
			if size <= 0 {
				d.FlashError("Error: Doodad size must be a positive number.")
			}
			d.NewDoodad(size)
		})
		return
	}

	log.Info("Starting a new doodad")
	scene := &EditorScene{
		DrawingType: enum.DoodadDrawing,
		DoodadSize:  size,
	}
	d.Goto(scene)
}

// EditDrawing loads a drawing (Level or Doodad) in Edit Mode.
func (d *Doodle) EditDrawing(filename string) error {
	log.Info("Loading drawing from file: %s", filename)
	ext := strings.ToLower(filepath.Ext(filename))

	scene := &EditorScene{
		Filename: filename,
		OpenFile: true,
	}

	switch ext {
	case ".level":
	case ".map":
		log.Info("is a Level type")
		scene.DrawingType = enum.LevelDrawing
	case ".doodad":
		scene.DrawingType = enum.DoodadDrawing
	default:
		return fmt.Errorf("file extension '%s' doesn't indicate its drawing type", ext)
	}

	d.Goto(scene)
	return nil
}

// PlayLevel loads a map from JSON into the PlayScene.
func (d *Doodle) PlayLevel(filename string) error {
	log.Info("Loading level from file: %s", filename)
	scene := &PlayScene{
		Filename: filename,
	}
	d.Goto(scene)
	return nil
}
