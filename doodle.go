package doodle

import (
	"fmt"
	"time"

	"git.kirsle.net/apps/doodle/events"
	"git.kirsle.net/apps/doodle/render"
	"github.com/kirsle/golog"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
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
	Debug bool

	startTime time.Time
	running   bool
	ticks     uint64
	events    *events.State
	width     int32
	height    int32

	scene Scene

	window   *sdl.Window
	renderer *sdl.Renderer
}

// New initializes the game object.
func New(debug bool) *Doodle {
	d := &Doodle{
		Debug:     debug,
		startTime: time.Now(),
		events:    events.New(),
		running:   true,
		width:     800,
		height:    600,
	}

	if !debug {
		log.Config.Level = golog.InfoLevel
	}

	return d
}

// Run initializes SDL and starts the main loop.
func (d *Doodle) Run() error {
	// Initialize SDL.
	log.Info("Initializing SDL")
	if err := sdl.Init(sdl.INIT_EVERYTHING); err != nil {
		return err
	}
	defer sdl.Quit()

	// Initialize SDL_TTF.
	log.Info("Initializing SDL_TTF")
	if err := ttf.Init(); err != nil {
		return err
	}

	// Create our window.
	log.Info("Creating the Main Window")
	window, err := sdl.CreateWindow(
		"Doodle v"+Version,
		sdl.WINDOWPOS_CENTERED,
		sdl.WINDOWPOS_CENTERED,
		d.width,
		d.height,
		sdl.WINDOW_SHOWN,
	)
	if err != nil {
		return err
	}
	defer window.Destroy()
	d.window = window

	// Blank out the window in white.
	log.Info("Creating the SDL Renderer")
	renderer, err := sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED)
	if err != nil {
		panic(err)
	}
	d.renderer = renderer
	render.Renderer = renderer
	defer renderer.Destroy()

	// Set up the default scene.
	if d.scene == nil {
		d.Goto(&EditorScene{})
	}

	log.Info("Enter Main Loop")
	for d.running {
		d.ticks++

		// Poll for events.
		_, err := d.events.Poll(d.ticks)
		if err != nil {
			log.Error("event poll error: %s", err)
			return err
		}

		// Draw a frame and log how long it took.
		start := time.Now()
		err = d.scene.Loop(d)
		if err != nil {
			return err
		}

		elapsed := time.Now().Sub(start)

		// Delay to maintain the target frames per second.
		tmp := elapsed / time.Millisecond
		var delay uint32
		if TargetFPS-int(tmp) > 0 { // make sure it won't roll under
			delay = uint32(TargetFPS - int(tmp))
		}
		sdl.Delay(delay)

		// Track how long this frame took to measure FPS over time.
		d.TrackFPS(delay)
	}

	log.Warn("Main Loop Exited! Shutting down...")
	return nil
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
	scene := &PlayScene{}
	err := scene.LoadLevel(filename)
	if err != nil {
		return err
	}
	d.Goto(scene)
	return nil
}

// TODO: not a global
type Pixel struct {
	start bool
	x     int32
	y     int32
	dx    int32
	dy    int32
}

func (p Pixel) String() string {
	return fmt.Sprintf("(%d,%d) delta (%d,%d)",
		p.x, p.y,
		p.dx, p.dy,
	)
}

// Grid is a 2D grid of pixels in X,Y notation.
type Grid map[Pixel]interface{}
