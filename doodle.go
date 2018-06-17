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
	Version = "0.0.0-alpha"

	// TargetFPS is the frame rate to cap the game to.
	TargetFPS = uint32(1000 / 60) // 60 FPS

	// Millisecond64 is a time.Millisecond casted to float64.
	Millisecond64 = float64(time.Millisecond)
)

// Doodle is the game object.
type Doodle struct {
	Debug bool

	startTime time.Time
	running   bool
	events    *events.State
	width     int32
	height    int32

	nextSecond time.Time
	canvas     Grid

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
		canvas:    Grid{},

		nextSecond: time.Now().Add(1 * time.Second),
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

	log.Info("Enter Main Loop")
	for d.running {
		// Draw a frame and log how long it took.
		start := time.Now()
		err = d.Loop()
		elapsed := time.Now().Sub(start)

		tmp := elapsed / time.Millisecond
		delay := TargetFPS - uint32(tmp)
		sdl.Delay(delay)

		d.TrackFPS(delay)
		if err != nil {
			return err
		}
	}

	log.Warn("Main Loop Exited! Shutting down...")
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

var pixelHistory []Pixel

// Loop runs one loop of the game engine.
func (d *Doodle) Loop() error {
	// Poll for events.
	ev, err := d.events.Poll()
	if err != nil {
		log.Error("event poll error: %s", err)
		return err
	}

	// Clear the canvas and fill it with white.
	d.renderer.SetDrawColor(255, 255, 255, 255)
	d.renderer.Clear()

	// Clicking? Log all the pixels while doing so.
	if ev.Button1.Now {
		pixel := Pixel{
			start: ev.Button1.Now && !ev.Button1.Last,
			x:     ev.CursorX.Now,
			y:     ev.CursorY.Now,
			dx:    ev.CursorX.Last,
			dy:    ev.CursorY.Last,
		}

		if len(pixelHistory) == 0 || pixelHistory[len(pixelHistory)-1] != pixel {
			pixelHistory = append(pixelHistory, pixel)
		}
	}

	d.renderer.SetDrawColor(0, 0, 0, 255)
	for i, pixel := range pixelHistory {
		if !pixel.start && i > 0 {
			prev := pixelHistory[i-1]
			if prev.x == pixel.x && prev.y == pixel.y {
				d.renderer.DrawPoint(pixel.x, pixel.y)
			} else {
				d.renderer.DrawLine(
					pixel.x,
					pixel.y,
					prev.x,
					prev.y,
				)
			}
		}
		d.renderer.DrawPoint(pixel.x, pixel.y)
	}

	// Draw the FPS.
	d.DrawDebugOverlay()

	d.renderer.Present()

	return nil
}
