package canvas

import (
	"time"

	"git.kirsle.net/apps/doodle/lib/events"
	"git.kirsle.net/apps/doodle/lib/render"
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
	events  *events.State
	running bool

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
	}

	return engine, nil
}

// WindowSize returns the size of the canvas window.
func (e *Engine) WindowSize() (w, h int) {
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

func (e *Engine) NewBitmap(filename string) (render.Texturer, error) {
	return nil, nil
}

func (e *Engine) Copy(t render.Texturer, src, dist render.Rect) {

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
