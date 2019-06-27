package canvas

import (
	"syscall/js"
	"time"

	"git.kirsle.net/apps/doodle/lib/events"
)

// Engine implements a rendering engine targeting an HTML canvas for
// WebAssembly targets.
type Engine struct {
	canvas    Canvas
	startTime time.Time
	width     int
	height    int
	ticks     uint32

	// Private fields.
	events   *events.State
	running  bool
	textures map[string]*Texture // cached texture PNG images

	// Event channel. WASM subscribes to events asynchronously using the
	// JavaScript APIs, whereas SDL2 polls the event queue which orders them
	// all up for processing. This channel will order and queue the events.
	queue chan Event
}

// New creates the Canvas Engine.
func New(canvasID string) (*Engine, error) {
	canvas := GetCanvas(canvasID)

	engine := &Engine{
		canvas:    canvas,
		startTime: time.Now(),
		events:    events.New(),
		width:     canvas.ClientW(),
		height:    canvas.ClientH(),
		queue:     make(chan Event, 1024),
		textures:  map[string]*Texture{},
	}

	return engine, nil
}

// WindowSize returns the size of the canvas window.
func (e *Engine) WindowSize() (w, h int) {
	// Good time to recompute it first?
	var (
		window = js.Global().Get("window")
		width  = window.Get("innerWidth").Int()
		height = window.Get("innerHeight").Int()
	)
	e.canvas.Value.Set("width", width)
	e.canvas.Value.Set("height", height)
	return e.canvas.ClientW(), e.canvas.ClientH()
}

// GetTicks returns the number of milliseconds since the engine started.
func (e *Engine) GetTicks() uint32 {
	ms := time.Since(e.startTime) * time.Millisecond
	return uint32(ms)
}

// TO BE IMPLEMENTED...

func (e *Engine) Setup() error {
	return nil
}

func (e *Engine) Present() error {
	return nil
}

// Delay for a moment.
func (e *Engine) Delay(delay uint32) {
	time.Sleep(time.Duration(delay) * time.Millisecond)
}

// Teardown tasks.
func (e *Engine) Teardown() {}

func (e *Engine) Loop() error {
	return nil
}
