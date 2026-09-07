package doodle

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"git.kirsle.net/SketchyMaze/doodle/pkg/balance"
	"git.kirsle.net/SketchyMaze/doodle/pkg/branding"
	"git.kirsle.net/SketchyMaze/doodle/pkg/cursor"
	"git.kirsle.net/SketchyMaze/doodle/pkg/enum"
	"git.kirsle.net/SketchyMaze/doodle/pkg/gamepad"
	"git.kirsle.net/SketchyMaze/doodle/pkg/keybind"
	"git.kirsle.net/SketchyMaze/doodle/pkg/levelpack"
	"git.kirsle.net/SketchyMaze/doodle/pkg/log"
	"git.kirsle.net/SketchyMaze/doodle/pkg/modal"
	"git.kirsle.net/SketchyMaze/doodle/pkg/modal/loadscreen"
	"git.kirsle.net/SketchyMaze/doodle/pkg/native"
	"git.kirsle.net/SketchyMaze/doodle/pkg/pattern"
	"git.kirsle.net/SketchyMaze/doodle/pkg/scripting/exceptions"
	"git.kirsle.net/SketchyMaze/doodle/pkg/shmem"
	"git.kirsle.net/SketchyMaze/doodle/pkg/usercfg"
	"git.kirsle.net/SketchyMaze/doodle/pkg/windows"
	golog "git.kirsle.net/go/log"
	"git.kirsle.net/go/render"
	"git.kirsle.net/go/render/event"
	"git.kirsle.net/go/ui"
)

const (
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
	shmem.PromptPre = d.PromptPre

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

	// Fixed-timestep bookkeeping: game logic (shmem.Tick, Scene.Loop and
	// everything keyed off ticks -- physics, animations, doodad script
	// timers) advances at a constant balance.TargetFPS ticks/sec of
	// simulated time, no matter how fast or slow the render loop actually
	// spins. This is what keeps gameplay speed identical whether the frame
	// rate is capped at 60 or uncapped to 700+ (the "unleash the beast"
	// cheat) or struggling on slow hardware.
	var (
		tickDuration    = time.Second / time.Duration(balance.TargetFPS)
		tickAccumulator time.Duration
		lastFrameTime   = time.Now()
	)
	for d.running {
		// d.Engine.Clear(render.White)

		start := time.Now() // Record how long this frame took.

		// Figure out how many simulation ticks have elapsed in real time
		// since the last render frame. Clamp the per-frame delta so a long
		// stall (breakpoint, window drag, OS suspend) doesn't queue up a
		// huge burst of catch-up ticks: a visible hitch is preferable to a
		// "spiral of death" where processing the backlog takes longer than
		// real time, pushing the backlog even further behind.
		frameDt := start.Sub(lastFrameTime)
		lastFrameTime = start
		if frameDt > balance.MaxFrameDelta {
			frameDt = balance.MaxFrameDelta
		}
		tickAccumulator += frameDt

		var ticksElapsed int
		for tickAccumulator >= tickDuration && ticksElapsed < balance.MaxTicksPerFrame {
			tickAccumulator -= tickDuration
			ticksElapsed++
		}

		// Poll for events.
		ev, err := d.Engine.Poll()
		if err != nil {
			log.Error("event poll error: %s", err)
			d.running = false
			break
		}
		d.event = ev

		// Always have an accurate idea of the window size.
		if ev.WindowResized {
			d.width, d.height = d.Engine.WindowSize()
		}

		// Update touch screen control (mainly whether to show the mouse cursor).
		native.UpdateTouchScreenMode(ev)

		// Let the gamepad controller check for events, if it's in MouseMode
		// it will fake the mouse cursor.
		gamepad.Loop(ev)

		// Globally store the cursor position.
		shmem.Cursor = render.NewPoint(ev.CursorX, ev.CursorY)

		// Whether the scene's simulation should advance this frame. Set by
		// whichever branch below applies, then acted on in a single unified
		// tick loop after the branching resolves (see the loop below the
		// branches for why this must be interleaved with the shmem.Tick
		// increment rather than done separately).
		var runScene bool

		// Command line shell.
		if d.shell.open {
			// The shell intentionally doesn't forward input to the scene
			// here (so typing in the console doesn't also move the player,
			// click buttons, etc. in the background) -- but that also meant
			// a window resize while the shell was open never reached the
			// scene's own layout code: Scene.Loop() is what recomputes
			// widget positions for the new size (Draw() below still runs
			// every frame regardless of shell state, but it only
			// re-renders whatever positions were last computed, which
			// don't change on their own). This matters especially on
			// Android, where opening the shell also requests the on-screen
			// keyboard (see pkg/native/keyboard_android.go), which resizes
			// the window to make room for it.
			//
			// Give the scene a chance to relayout without also processing
			// input meant for the shell: forward a fresh, mostly-empty
			// event.State with only WindowResized (and the cursor
			// position, in case layout code depends on it) set, rather
			// than the real `ev`, which may carry clicks/keys the shell
			// itself is handling this frame.
			if ev.WindowResized {
				resizeOnly := event.NewState()
				resizeOnly.WindowResized = true
				resizeOnly.CursorX = ev.CursorX
				resizeOnly.CursorY = ev.CursorY
				if err := d.Scene.Loop(d, resizeOnly); err != nil {
					return err
				}
			}
		} else if keybind.ShellKey(ev) {
			log.Debug("Shell: opening shell")
			d.shell.Open()
		} else {
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
			if !modal.Handled(ev) && !exceptions.Handled(ev) {
				// Global event handlers.
				if keybind.Shutdown(ev) {
					if d.Debug { // fast exit in -debug mode.
						d.running = false
					} else {
						d.ConfirmExit()
					}
					continue
				}

				// Run the scene's logic below, in the unified tick loop.
				runScene = true
			}

		}

		// Advance the simulation and, if applicable, run the scene's logic
		// -- once per elapsed simulation tick, so movement speeds,
		// animations, physics, etc. all advance at a fixed rate of
		// simulated time regardless of how fast or slow frames are
		// actually being rendered. On a normal 60 FPS frame this runs
		// exactly once, same as before; an uncapped high-FPS frame may run
		// it zero times, and a slow frame that fell behind may run it more
		// than once to catch back up to real time.
		//
		// shmem.Tick is incremented right here, interleaved with each
		// individual Scene.Loop() call, rather than being bulk-advanced by
		// ticksElapsed beforehand: code keyed off shmem.Tick (animation
		// frame timing, doodad setTimeout/setInterval, jump cooldowns...)
		// expects it to climb by exactly 1 between successive ticks. If it
		// were bulk-advanced first, every call within a multi-tick
		// catch-up burst would see the *same*, already fully-advanced
		// value, quantizing all of that timing to the burst size instead
		// of single ticks. That was a real, observed bug: on a sustained
		// low FPS device (e.g. ~10-12 FPS on Android) every real frame
		// hits the MaxTicksPerFrame catch-up cap, so an animation like the
		// electric door's "close" (timed in-ticks) would finish later in
		// real time than intended -- giving a player extra real-world time
		// to run through a door that should have already sealed behind
		// them, since their movement (also gated on the same shmem.Tick)
		// no longer advanced in lockstep with the door's timer.
		//
		// The shmem.Tick increment happens even when runScene is false
		// (e.g. while a modal or the dev shell is open) since some UI bits
		// -- e.g. the shell's blinking cursor, flash message expiry -- key
		// off shmem.Tick too and should keep ticking at a steady rate
		// independent of render FPS.
		//
		// WindowResized is a special case needing at least one pass even
		// with zero elapsed ticks: it's an edge-triggered flag that is
		// only ever true on the single real frame the resize happened (the
		// engine clears it on the very next Poll()), so if it landed on a
		// frame with zero elapsed ticks -- e.g. the first frame or two
		// after startup, if the window opens maximized, before the
		// accumulator has filled -- it would otherwise be silently lost
		// forever and the scene would never learn its canvas needs to
		// resize.
		loopCount := ticksElapsed
		if loopCount == 0 && runScene && ev.WindowResized {
			loopCount = 1
		}
		for i := 0; i < loopCount; i++ {
			shmem.Tick++
			if runScene {
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
		exceptions.Draw(d.Engine)
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

		// Draw our custom mouse cursor to the screen.
		cursor.Draw(d.Engine)

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
			if balance.TargetClockRate-int(tmp) > 0 { // make sure it won't roll under
				delay = uint32(balance.TargetClockRate - int(tmp))
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
		OnOpenCheatsWindow: func() *ui.Window {
			return d.MakeCheatsWindow(supervisor)
		},

		// Boolean checkbox bindings
		DebugOverlay:       &DebugOverlay,
		DebugCollision:     &DebugCollision,
		HorizontalToolbars: &usercfg.Current.HorizontalToolbars,
		EnableFeatures:     &usercfg.Current.EnableFeatures,
		CrosshairSize:      &usercfg.Current.CrosshairSize,
		CrosshairColor:     &usercfg.Current.CrosshairColor,
		HideTouchHints:     &usercfg.Current.HideTouchHints,
		DisableAutosave:    &usercfg.Current.DisableAutosave,
		ControllerStyle:    &usercfg.Current.ControllerStyle,
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

// Shutdown flags the main loop (Run) to exit gracefully after the current
// frame, the same way the in-game quit keybind does in -debug mode. Exported
// so external callers (e.g. cmd/doodle's SIGINT/SIGTERM handler) can stop the
// loop cleanly -- notably so a -pprof CPU profile in progress still gets
// flushed to disk via the deferred pprof.StopCPUProfile(), which a bare
// process kill would skip entirely.
func (d *Doodle) Shutdown() {
	d.running = false
}

// NewMap loads a new map in Edit Mode.
func (d *Doodle) NewMap() {
	log.Info("Starting a new map")
	scene := &EditorScene{}
	d.Goto(scene)
}

// NewDoodad loads a new Doodad in Edit Mode.
// If size is zero, it prompts the user to select a size or accept the default size.
func (d *Doodle) NewDoodad(width, height int) {
	if width+height == 0 {
		width = balance.DoodadSize
		height = width
	}

	log.Info("Starting a new doodad")
	scene := &EditorScene{
		DrawingType: enum.DoodadDrawing,
		DoodadSize:  render.NewRect(width, height),
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

// PlayFromLevelpack initializes the Play Scene from a level as part of
// a levelpack.
func (d *Doodle) PlayFromLevelpack(pack *levelpack.LevelPack, which levelpack.Level) error {
	log.Info("Loading level %s from levelpack %s", which.Filename, pack.Title)
	scene := &PlayScene{
		Filename:  which.Filename,
		LevelPack: pack,
	}
	d.Goto(scene)
	return nil
}
