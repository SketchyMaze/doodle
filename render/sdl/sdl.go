// Package sdl provides an SDL2 renderer for Doodle.
package sdl

import (
	"errors"
	"time"

	"git.kirsle.net/apps/doodle/events"
	"git.kirsle.net/apps/doodle/render"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

// Renderer manages the SDL state.
type Renderer struct {
	// Configurable fields.
	title     string
	width     int32
	height    int32
	startTime time.Time

	// Private fields.
	events   *events.State
	window   *sdl.Window
	renderer *sdl.Renderer
	running  bool
	ticks    uint64

	// Optimizations to minimize SDL calls.
	lastColor render.Color
}

// New creates the SDL renderer.
func New(title string, width, height int32) *Renderer {
	return &Renderer{
		events: events.New(),
		title:  title,
		width:  width,
		height: height,
	}
}

// Teardown tasks when exiting the program.
func (r *Renderer) Teardown() {
	r.renderer.Destroy()
	r.window.Destroy()
	sdl.Quit()
}

// Setup the renderer.
func (r *Renderer) Setup() error {
	// Initialize SDL.
	log.Info("Initializing SDL")
	if err := sdl.Init(sdl.INIT_EVERYTHING); err != nil {
		return err
	}

	// Initialize SDL_TTF.
	log.Info("Initializing SDL_TTF")
	if err := ttf.Init(); err != nil {
		return err
	}

	// Create our window.
	log.Info("Creating the Main Window")
	window, err := sdl.CreateWindow(
		r.title,
		sdl.WINDOWPOS_CENTERED,
		sdl.WINDOWPOS_CENTERED,
		r.width,
		r.height,
		sdl.WINDOW_SHOWN,
	)
	if err != nil {
		return err
	}
	r.window = window

	// Blank out the window in white.
	log.Info("Creating the SDL Renderer")
	renderer, err := sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED)
	if err != nil {
		panic(err)
	}
	r.renderer = renderer

	return nil
}

// GetTicks gets SDL's current tick count.
func (r *Renderer) GetTicks() uint32 {
	return sdl.GetTicks()
}

// Poll for events.
func (r *Renderer) Poll() (*events.State, error) {
	s := r.events
	for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
		switch t := event.(type) {
		case *sdl.QuitEvent:
			return s, errors.New("quit")
		case *sdl.MouseMotionEvent:
			if DebugMouseEvents {
				log.Debug("[%d ms] tick:%d MouseMotion  type:%d  id:%d  x:%d  y:%d  xrel:%d  yrel:%d",
					t.Timestamp, r.ticks, t.Type, t.Which, t.X, t.Y, t.XRel, t.YRel,
				)
			}

			// Push the cursor position.
			s.CursorX.Push(t.X)
			s.CursorY.Push(t.Y)
			s.Button1.Push(t.State == 1)
		case *sdl.MouseButtonEvent:
			if DebugClickEvents {
				log.Debug("[%d ms] tick:%d MouseButton  type:%d  id:%d  x:%d  y:%d  button:%d  state:%d",
					t.Timestamp, r.ticks, t.Type, t.Which, t.X, t.Y, t.Button, t.State,
				)
			}

			// Push the cursor position.
			s.CursorX.Push(t.X)
			s.CursorY.Push(t.Y)

			// Is a mouse button pressed down?
			if t.Button == 1 {
				var eventName string
				if DebugClickEvents {
					if t.State == 1 && s.Button1.Now == false {
						eventName = "DOWN"
					} else if t.State == 0 && s.Button1.Now == true {
						eventName = "UP"
					}
				}

				if eventName != "" {
					log.Debug("tick:%d  Mouse Button1 %s BEFORE: %+v",
						r.ticks,
						eventName,
						s.Button1,
					)
					s.Button1.Push(eventName == "DOWN")
					log.Debug("tick:%d  Mouse Button1 %s AFTER: %+v",
						r.ticks,
						eventName,
						s.Button1,
					)

					// Return the event immediately.
					return s, nil
				}
			}

			// s.Button2.Push(t.Button == 3 && t.State == 1)
		case *sdl.MouseWheelEvent:
			if DebugMouseEvents {
				log.Debug("[%d ms] tick:%d MouseWheel  type:%d  id:%d  x:%d  y:%d",
					t.Timestamp, r.ticks, t.Type, t.Which, t.X, t.Y,
				)
			}
		case *sdl.KeyboardEvent:
			if DebugKeyEvents {
				log.Debug("[%d ms] tick:%d Keyboard  type:%d  sym:%c  modifiers:%d  state:%d  repeat:%d\n",
					t.Timestamp, r.ticks, t.Type, t.Keysym.Sym, t.Keysym.Mod, t.State, t.Repeat,
				)
			}

			switch t.Keysym.Scancode {
			case sdl.SCANCODE_ESCAPE:
				if t.Repeat == 1 {
					continue
				}
				s.EscapeKey.Push(t.State == 1)
			case sdl.SCANCODE_RETURN:
				if t.Repeat == 1 {
					continue
				}
				s.EnterKey.Push(t.State == 1)
			case sdl.SCANCODE_F12:
				s.ScreenshotKey.Push(t.State == 1)
			case sdl.SCANCODE_UP:
				s.Up.Push(t.State == 1)
			case sdl.SCANCODE_LEFT:
				s.Left.Push(t.State == 1)
			case sdl.SCANCODE_RIGHT:
				s.Right.Push(t.State == 1)
			case sdl.SCANCODE_DOWN:
				s.Down.Push(t.State == 1)
			case sdl.SCANCODE_LSHIFT:
			case sdl.SCANCODE_RSHIFT:
				s.ShiftActive.Push(t.State == 1)
				continue
			case sdl.SCANCODE_LALT:
			case sdl.SCANCODE_RALT:
			case sdl.SCANCODE_LCTRL:
			case sdl.SCANCODE_RCTRL:
				continue
			case sdl.SCANCODE_BACKSPACE:
				// Make it a key event with "\b" as the sequence.
				if t.State == 1 || t.Repeat == 1 {
					s.KeyName.Push(`\b`)
				}
			default:
				// Push the string value of the key.
				if t.State == 1 {
					s.KeyName.Push(string(t.Keysym.Sym))
				}
			}
		}
	}

	return s, nil
}

// Present the current frame.
func (r *Renderer) Present() error {
	r.renderer.Present()
	return nil
}

// Delay using sdl.Delay
func (r *Renderer) Delay(time uint32) {
	sdl.Delay(time)
}

// Loop is the main loop.
func (r *Renderer) Loop() error {
	return nil
}
