package doodle

import "github.com/veandco/go-sdl2/sdl"

// Version number.
const Version = "0.0.0-alpha"

// Doodle is the game object.
type Doodle struct {
	Debug bool

	window   *sdl.Window
	renderer *sdl.Renderer
}

// New initializes the game object.
func New(debug bool) *Doodle {
	d := &Doodle{
		Debug: debug,
	}

	// Initialize SDL.
	if err := sdl.Init(sdl.INIT_EVERYTHING); err != nil {
		panic(err)
	}
	defer sdl.Quit()

	// Create our window.
	window, err := sdl.CreateWindow(
		"Doodle v"+Version,
		sdl.WINDOWPOS_CENTERED,
		sdl.WINDOWPOS_CENTERED,
		800,
		600,
		sdl.WINDOW_SHOWN,
	)
	if err != nil {
		panic(err)
	}
	defer window.Destroy()

	surface, err := window.GetSurface()
	if err != nil {
		panic(err)
	}

	rect := sdl.Rect{
		X: 0,
		Y: 0,
		W: 200,
		H: 200,
	}
	surface.FillRect(&rect, 0xffff0000)
	window.UpdateSurface()

	sdl.Delay(2500)

	return d
}
