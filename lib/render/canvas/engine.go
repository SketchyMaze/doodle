package canvas

import (
	"errors"
	"syscall/js"
	"time"

	"git.kirsle.net/apps/doodle/lib/events"
	"git.kirsle.net/apps/doodle/lib/render"
	"git.kirsle.net/apps/doodle/pkg/wasm"
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

// Texture can hold on to cached image textures.
type Texture struct {
	data   string   // data:image/png URI
	image  js.Value // DOM image element
	width  int
	height int
}

// Size returns the dimensions of the texture.
func (t *Texture) Size() render.Rect {
	return render.NewRect(int32(t.width), int32(t.height))
}

// NewBitmap initializes a texture from a bitmap image. The image is stored
// in HTML5 Session Storage.
func (e *Engine) NewBitmap(filename string) (render.Texturer, error) {
	if data, ok := wasm.GetSession(filename); ok {
		img := js.Global().Get("document").Call("createElement", "img")
		img.Set("src", data)
		return &Texture{
			data:   data,
			image:  img,
			width:  60, // TODO
			height: 60,
		}, nil
	}

	return nil, errors.New("no bitmap data stored for " + filename)

}

var TODO int

// Copy a texturer bitmap onto the canvas.
func (e *Engine) Copy(t render.Texturer, src, dist render.Rect) {
	tex := t.(*Texture)

	// image := js.Global().Get("document").Call("createElement", "img")
	// image.Set("src", tex.data)

	// log.Info("drawing image just this once")
	e.canvas.ctx2d.Call("drawImage", tex.image, dist.X, dist.Y)
	// TODO++
	// if TODO > 200 {
	// 	log.Info("I exited at engine.Copy for canvas engine")
	// 	os.Exit(0)
	// }
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
