package doodle

import (
	"time"

	"git.kirsle.net/apps/doodle/render"
	"github.com/kirsle/golog"
)

const (
	// Version number.
	Version = "0.0.1-alpha"

	// TargetFPS is the frame rate to cap the game to.
	TargetFPS = 1000 / 60 // 60 FPS

	// Millisecond64 is a time.Millisecond casted to float64.
	Millisecond64 = float64(time.Millisecond)
)

// Doodle is the game object.
type Doodle struct {
	Debug  bool
	Engine render.Engine

	startTime time.Time
	running   bool
	ticks     uint64
	width     int32
	height    int32

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
		width:     800,
		height:    600,
	}
	d.shell = NewShell(d)

	if !debug {
		log.Config.Level = golog.InfoLevel
	}

	return d
}

// Run initializes SDL and starts the main loop.
func (d *Doodle) Run() error {
	// Set up the render engine.
	if err := d.Engine.Setup(); err != nil {
		return err
	}

	// Set up the default scene.
	if d.Scene == nil {
		d.Goto(&MainScene{})
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
		elapsed := time.Now().Sub(start)
		tmp := elapsed / time.Millisecond
		var delay uint32
		if TargetFPS-int(tmp) > 0 { // make sure it won't roll under
			delay = uint32(TargetFPS - int(tmp))
		}
		d.Engine.Delay(delay)

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

// EditLevel loads a map from JSON into the EditorScene.
func (d *Doodle) EditLevel(filename string) error {
	log.Info("Loading level from file: %s", filename)
	scene := &EditorScene{}
	err := scene.LoadLevel(filename)
	if err != nil {
		return err
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
