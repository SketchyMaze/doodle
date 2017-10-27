package doodle

import (
	"fmt"

	"github.com/veandco/go-sdl2/sdl"
)

// Version number.
const Version = "0.0.0-alpha"

// Doodle is the game object.
type Doodle struct {
	Debug bool

	running bool
	events  EventState
	width   int
	height  int

	window   *sdl.Window
	surface  *sdl.Surface
	renderer *sdl.Renderer
}

// EventState keeps track of important events.
type EventState struct {
	CursorX    int32
	CursorY    int32
	LastX      int32
	LastY      int32
	LeftClick  bool
	LastLeft   bool
	RightClick bool
	LastRight  bool
}

// New initializes the game object.
func New(debug bool) *Doodle {
	d := &Doodle{
		Debug:   debug,
		running: true,
		width:   800,
		height:  600,
	}
	return d
}

// Run initializes SDL and starts the main loop.
func (d *Doodle) Run() error {
	// Initialize SDL.
	if err := sdl.Init(sdl.INIT_EVERYTHING); err != nil {
		return err
	}
	defer sdl.Quit()

	// Create our window.
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
	renderer, err := sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED)
	if err != nil {
		panic(err)
	}
	d.renderer = renderer
	defer renderer.Destroy()

	for i := 0; i < 10; i++ {
		d.Loop()
		// renderer.Clear()
		// rect := sdl.Rect{
		// 	X: 0,
		// 	Y: 0,
		// 	W: 800,
		// 	H: 600,
		// }
		// renderer.SetDrawColor(0, 0, 0, 255)
		// renderer.FillRect(&rect)
		//
		// renderer.SetDrawColor(0, 255, 0, 255)
		// renderer.DrawPoint(10*i, 10*i)
		//
		// renderer.Present()
		//
		// sdl.Delay(250)
	}

	for d.running {
		err = d.Loop()
		if err != nil {
			return err
		}
	}

	// surface, err := window.GetSurface()
	// if err != nil {
	// 	panic(err)
	// }
	// d.surface = surface
	//
	// rect := sdl.Rect{
	// 	X: 0,
	// 	Y: 0,
	// 	W: 200,
	// 	H: 200,
	// }
	// surface.FillRect(&rect, 0xffff0000)
	// window.UpdateSurface()
	//
	// sdl.Delay(2500)
	return nil
}

// TODO: not a global
type Pixel struct {
	start bool
	x     int32
	y     int32
}

var pixelHistory []Pixel

// Loop runs one loop of the game engine.
func (d *Doodle) Loop() error {
	// Poll for events.
	d.PollEvents()

	d.renderer.Clear()
	rect := sdl.Rect{
		X: 0,
		Y: 0,
		W: 800,
		H: 600,
	}
	d.renderer.SetDrawColor(255, 255, 255, 255)
	d.renderer.FillRect(&rect)

	// Clicking? Log all the pixels while doing so.
	if d.events.LeftClick {
		fmt.Printf("Pixel at %dx%d\n", d.events.CursorX, d.events.CursorY)
		pixel := Pixel{
			start: d.events.LeftClick && !d.events.LastLeft,
			x:     d.events.CursorX,
			y:     d.events.CursorY,
		}
		pixelHistory = append(pixelHistory, pixel)
	}

	// Colorize all those pixels.
	d.renderer.SetDrawColor(0, 0, 0, 255)
	for i, pixel := range pixelHistory {
		fmt.Printf("Draw: %v\n", pixel)
		if pixel.start == false && i > 0 {
			start := pixelHistory[i-1]
			fmt.Printf("Line from %dx%d -> %dx%d\n", start.x, start.y, pixel.x, pixel.y)
			d.renderer.DrawLine(
				int(start.x),
				int(start.y),
				int(pixel.x),
				int(pixel.y),
			)
		} else {
			d.renderer.DrawPoint(
				int(pixel.x), int(pixel.y),
			)
		}
		// d.renderer.FillRect(&sdl.Rect{pixel.x, pixel.y, 10, 10})
	}

	d.renderer.Present()
	sdl.Delay(1000 / 60)

	return nil
}

// PollEvents checks for keyboard/mouse/etc. events.
func (d *Doodle) PollEvents() {
	for {
		event := sdl.PollEvent()
		if event == nil {
			break
		}

		// Handle the event.
		switch t := event.(type) {
		case *sdl.QuitEvent:
			d.running = false
		case *sdl.MouseMotionEvent:
			fmt.Printf("[%d ms] MouseMotion  type:%d  id:%d  x:%d  y:%d  xrel:%d  yrel:%d\n",
				t.Timestamp, t.Type, t.Which, t.X, t.Y, t.XRel, t.YRel,
			)
			d.events.LastX = d.events.CursorX
			d.events.LastY = d.events.CursorY
			d.events.CursorX = t.X
			d.events.CursorY = t.Y
		case *sdl.MouseButtonEvent:
			fmt.Printf("[%d ms] MouseButton  type:%d  id:%d  x:%d  y:%d  button:%d  state:%d\n",
				t.Timestamp, t.Type, t.Which, t.X, t.Y, t.Button, t.State,
			)

			d.events.LastX = d.events.CursorX
			d.events.LastY = d.events.CursorY
			d.events.CursorX = t.X
			d.events.CursorY = t.Y

			d.events.LastLeft = d.events.LeftClick
			d.events.LastRight = d.events.RightClick

			// Clicking?
			if t.Button == 1 {
				if t.State == 1 && d.events.LeftClick == false {
					d.events.LeftClick = true
				} else if t.State == 0 && d.events.LeftClick == true {
					d.events.LeftClick = false
				}
			}
			d.events.RightClick = t.Button == 3 && t.State == 1
		case *sdl.MouseWheelEvent:
			fmt.Printf("[%d ms] MouseWheel  type:%d  id:%d  x:%d  y:%d\n",
				t.Timestamp, t.Type, t.Which, t.X, t.Y,
			)
		case *sdl.KeyDownEvent:
			fmt.Printf("[%d ms] Keyboard  type:%d  sym:%c  modifiers:%d  state:%d  repeat:%d\n",
				t.Timestamp, t.Type, t.Keysym.Sym, t.Keysym.Mod, t.State, t.Repeat,
			)
		case *sdl.KeyUpEvent:
			fmt.Printf("[%d ms] Keyboard  type:%d  sym:%c  modifiers:%d  state:%d  repeat:%d\n",
				t.Timestamp, t.Type, t.Keysym.Sym, t.Keysym.Mod, t.State, t.Repeat,
			)
		}
	}
}
