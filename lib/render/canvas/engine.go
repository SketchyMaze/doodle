package canvas

import (
	"bytes"
	"encoding/base64"
	"errors"
	"image"
	"image/png"
	"syscall/js"
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

// Texture can hold on to cached image textures.
type Texture struct {
	data   string   // data:image/png URI
	image  js.Value // DOM image element
	canvas js.Value // Warmed up canvas element
	ctx2d  js.Value // 2D drawing context for the canvas.
	width  int
	height int
}

// NewTexture caches a texture from a bitmap.
func (e *Engine) NewTexture(filename string, img image.Image) (render.Texturer, error) {
	var (
		fh        = bytes.NewBuffer([]byte{})
		imageSize = img.Bounds().Size()
		width     = imageSize.X
		height    = imageSize.Y
	)

	// Encode to PNG format.
	if err := png.Encode(fh, img); err != nil {
		return nil, err
	}

	var dataURI = "data:image/png;base64," + base64.StdEncoding.EncodeToString(fh.Bytes())

	tex := &Texture{
		data:   dataURI,
		width:  width,
		height: height,
	}

	// Preheat a cached Canvas object.
	canvas := js.Global().Get("document").Call("createElement", "canvas")
	canvas.Set("width", width)
	canvas.Set("height", height)
	tex.canvas = canvas

	ctx2d := canvas.Call("getContext", "2d")
	tex.ctx2d = ctx2d

	// Load as a JS Image object.
	image := js.Global().Call("eval", "new Image()")
	image.Call("addEventListener", "load", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		ctx2d.Call("drawImage", image, 0, 0)
		return nil
	}))
	image.Set("src", tex.data)
	tex.image = image

	// Cache the texture in memory.
	e.textures[filename] = tex

	return tex, nil
}

// Size returns the dimensions of the texture.
func (t *Texture) Size() render.Rect {
	return render.NewRect(int32(t.width), int32(t.height))
}

// NewBitmap initializes a texture from a bitmap image. The image is stored
// in HTML5 Session Storage.
func (e *Engine) NewBitmap(filename string) (render.Texturer, error) {
	if tex, ok := e.textures[filename]; ok {
		return tex, nil
	}

	panic("no bitmap for " + filename)
	return nil, errors.New("no bitmap data stored for " + filename)
}

// Copy a texturer bitmap onto the canvas.
func (e *Engine) Copy(t render.Texturer, src, dist render.Rect) {
	tex := t.(*Texture)

	// e.canvas.ctx2d.Call("drawImage", tex.image, dist.X, dist.Y)
	e.canvas.ctx2d.Call("drawImage", tex.canvas, dist.X, dist.Y)

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
