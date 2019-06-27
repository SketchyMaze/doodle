package canvas

// Text rendering functions using the HTML 5 canvas.

import (
	"fmt"
	"path/filepath"
	"strings"

	"git.kirsle.net/apps/doodle/lib/render"
)

// FontFilenameToName converts a FontFilename to its CSS font name.
//
// The CSS font name is set to the base of the filename, without the .ttf
// file extension. For example, "fonts/DejaVuSans.ttf" uses the CSS font
// family name "DejaVuSans" and that's what this function returns.
//
// Fonts must be defined in the index.html style sheet when serving the
// wasm build of Doodle.
//
// If filename is "", returns "serif" as a sensible default.
func FontFilenameToName(filename string) string {
	if filename == "" {
		return "DejaVuSans,serif"
	}
	return strings.TrimSuffix(filepath.Base(filename), filepath.Ext(filename))
}

// DrawText draws text on the canvas.
func (e *Engine) DrawText(text render.Text, point render.Point) error {
	font := FontFilenameToName(text.FontFilename)
	e.canvas.ctx2d.Set("font",
		fmt.Sprintf("%dpx %s,serif", text.Size, font),
	)

	e.canvas.ctx2d.Set("textBaseline", "top")

	write := func(dx, dy int, color render.Color) {
		e.canvas.ctx2d.Set("fillStyle", color.ToHex())
		e.canvas.ctx2d.Call("fillText",
			text.Text,
			int(point.X)+dx,
			int(point.Y)+dy,
		)
	}

	// Does the text have a stroke around it?
	if text.Stroke != render.Invisible {
		e.canvas.ctx2d.Set("fillStyle", text.Stroke.ToHex())
		write(-1, -1, text.Stroke)
		write(-1, 0, text.Stroke)
		write(-1, 1, text.Stroke)
		write(1, -1, text.Stroke)
		write(1, 0, text.Stroke)
		write(1, 1, text.Stroke)
		write(0, -1, text.Stroke)
		write(0, 1, text.Stroke)
	}

	// Does it have a drop shadow?
	if text.Shadow != render.Invisible {
		write(1, 1, text.Shadow)
	}

	// Draw the text itself.
	write(0, 0, text.Color)

	return nil
}

// ComputeTextRect computes and returns a Rect for how large the text would
// appear if rendered.
func (e *Engine) ComputeTextRect(text render.Text) (render.Rect, error) {
	font := FontFilenameToName(text.FontFilename)
	e.canvas.ctx2d.Set("font",
		fmt.Sprintf("%dpx %s,serif", text.Size, font),
	)

	measure := e.canvas.ctx2d.Call("measureText", text.Text)
	rect := render.Rect{
		// TODO: the only TextMetrics widely supported in browsers is
		// the width. For height, use the text size for now.
		W: int32(measure.Get("width").Int()),
		H: int32(text.Size),
	}
	return rect, nil
}
