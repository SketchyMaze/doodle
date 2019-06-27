package canvas

import (
	"syscall/js"
)

// Canvas represents an HTML5 Canvas object.
type Canvas struct {
	Value js.Value
	ctx2d js.Value
}

// GetCanvas gets an HTML5 Canvas object from the DOM.
func GetCanvas(id string) Canvas {
	canvasEl := js.Global().Get("document").Call("getElementById", id)
	canvas2d := canvasEl.Call("getContext", "2d")

	c := Canvas{
		Value: canvasEl,
		ctx2d: canvas2d,
	}

	canvasEl.Set("width", c.ClientW())
	canvasEl.Set("height", c.ClientH())

	return c
}

// ClientW returns the client width.
func (c Canvas) ClientW() int {
	return c.Value.Get("clientWidth").Int()
}

// ClientH returns the client height.
func (c Canvas) ClientH() int {
	return c.Value.Get("clientHeight").Int()
}
