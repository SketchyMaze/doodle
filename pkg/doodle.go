package doodle

import (
	"fmt"
	"strings"
	"time"

	"git.kirsle.net/apps/doodle/lib/events"
	"git.kirsle.net/apps/doodle/lib/render"
	"git.kirsle.net/apps/doodle/pkg/balance"
	"git.kirsle.net/apps/doodle/pkg/enum"
	"git.kirsle.net/apps/doodle/pkg/log"
	"github.com/kirsle/golog"
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
	event *events.State

	startTime time.Time
	running   bool
	ticks     uint64
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

	if debug {
		log.Logger.Config.Level = golog.DebugLevel
		DebugOverlay = true // on by default in debug mode, F3 to disable
	}

	return d
}

// SetupEngine sets up the rendering engine.
func (d *Doodle) SetupEngine() error {
	if err := d.Engine.Setup(); err != nil {
		return err
	}
	d.engineReady = true
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
		// d.Goto(&GUITestScene{})
		d.NewMap()
		// d.Goto(&MainScene{})
	}

	log.Info("Enter Main Loop")
	for d.running {
		d.Engine.Clear(render.White)

		start := time.Now() // Record how long this frame took.
		d.ticks++

		// Poll for events.
		ev, err := d.Engine.Poll()
		if err != nil {
			log.Error("event poll error: %s", err)
			d.running = false
			break
		}
		d.event = ev

		// Command line shell.
		if d.shell.Open {
		} else if ev.EnterKey.Read() {
			log.Debug("Shell: opening shell")
			d.shell.Open = true
		} else {
			// Global event handlers.
			if ev.EscapeKey.Read() {
				log.Error("Escape key pressed, shutting down")
				d.running = false
				break
			}

			if ev.KeyName.Now == "F3" {
				DebugOverlay = !DebugOverlay
				ev.KeyName.Read()
			} else if ev.KeyName.Now == "F4" {
				DebugCollision = !DebugCollision
				ev.KeyName.Read()
			}

			// Run the scene's logic.
			err = d.Scene.Loop(d, ev)
			if err != nil {
				return err
			}
		}

		// Draw the scene.
		d.Scene.Draw(d)

		// Draw the shell.
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
		ev.KeyName.Read()
	}

	log.Warn("Main Loop Exited! Shutting down...")
	return nil
}

// NewMap loads a new map in Edit Mode.
func (d *Doodle) NewMap() {
	log.Info("Starting a new map")
	scene := &EditorScene{}
	d.Goto(scene)
}

// NewDoodad loads a new Doodad in Edit Mode.
func (d *Doodle) NewDoodad(size int) {
	if balance.FreeVersion {
		d.Flash("Doodad editor is not available in your version of the game.")
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
	parts := strings.Split(filename, ".")
	if len(parts) < 2 {
		return fmt.Errorf("filename `%s` has no file extension", filename)
	}
	ext := strings.ToLower(parts[len(parts)-1])

	scene := &EditorScene{
		Filename: filename,
		OpenFile: true,
	}

	switch ext {
	case "level":
	case "map":
		log.Info("is a Level type")
		scene.DrawingType = enum.LevelDrawing
	case "doodad":
		if balance.FreeVersion {
			return fmt.Errorf("Doodad editor not supported in your version of the game")
		}
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
