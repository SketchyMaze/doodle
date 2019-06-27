package canvas

import (
	"bytes"
	"encoding/base64"
	"errors"
	"image"
	"image/png"
	"syscall/js"

	"git.kirsle.net/apps/doodle/lib/render"
)

// Texture can hold on to cached image textures.
type Texture struct {
	data   string   // data:image/png URI
	image  js.Value // DOM image element
	canvas js.Value // Warmed up canvas element
	ctx2d  js.Value // 2D drawing context for the canvas.
	width  int
	height int
}

// StoreTexture caches a texture from a bitmap.
func (e *Engine) StoreTexture(name string, img image.Image) (render.Texturer, error) {
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
	e.textures[name] = tex

	return tex, nil
}

// Size returns the dimensions of the texture.
func (t *Texture) Size() render.Rect {
	return render.NewRect(int32(t.width), int32(t.height))
}

// LoadTexture recalls a cached texture image.
func (e *Engine) LoadTexture(name string) (render.Texturer, error) {
	if tex, ok := e.textures[name]; ok {
		return tex, nil
	}

	return nil, errors.New("no bitmap data stored for " + name)
}

// Copy a texturer bitmap onto the canvas.
func (e *Engine) Copy(t render.Texturer, src, dist render.Rect) {
	tex := t.(*Texture)

	// e.canvas.ctx2d.Call("drawImage", tex.image, dist.X, dist.Y)
	e.canvas.ctx2d.Call("drawImage", tex.canvas, dist.X, dist.Y)

}
