package level

import (
	"encoding/json"
	"fmt"
	"image"
	"math"
	"os"
	"runtime"

	"git.kirsle.net/apps/doodle/lib/render"
	"git.kirsle.net/apps/doodle/pkg/balance"
	"git.kirsle.net/apps/doodle/pkg/log"
	"git.kirsle.net/apps/doodle/pkg/userdir"
	"git.kirsle.net/apps/doodle/pkg/wasm"
	"github.com/satori/go.uuid"
	"github.com/vmihailenco/msgpack"
)

// Types of chunks.
const (
	MapType int = iota
	GridType
)

// Chunk holds a single portion of the pixel canvas.
type Chunk struct {
	Type int // map vs. 2D array.
	Accessor

	// Values told to it from higher up, not stored in JSON.
	Point render.Point
	Size  int

	// Texture cache properties so we don't redraw pixel-by-pixel every frame.
	uuid               uuid.UUID
	texture            render.Texturer
	textureMasked      render.Texturer
	textureMaskedColor render.Color
	dirty              bool
}

// JSONChunk holds a lightweight (interface-free) copy of the Chunk for
// unmarshalling JSON files from disk.
type JSONChunk struct {
	Type    int             `json:"type" msgpack:"0"`
	Data    json.RawMessage `json:"data" msgpack:"-"`
	BinData interface{}     `json:"-" msgpack:"1"`
}

// Accessor provides a high-level API to interact with absolute pixel coordinates
// while abstracting away the details of how they're stored.
type Accessor interface {
	Inflate(*Palette) error
	Iter() <-chan Pixel
	IterViewport(viewport render.Rect) <-chan Pixel
	Get(render.Point) (*Swatch, error)
	Set(render.Point, *Swatch) error
	Delete(render.Point) error
	Len() int
	MarshalJSON() ([]byte, error)
	UnmarshalJSON([]byte) error
	// MarshalMsgpack() ([]byte, error)
	// UnmarshalMsgpack([]byte) error
	// Serialize() interface{}
}

// NewChunk creates a new chunk.
func NewChunk() *Chunk {
	return &Chunk{
		Type:     MapType,
		Accessor: NewMapAccessor(),
	}
}

// Texture will return a cached texture for the rendering engine for this
// chunk's pixel data. If the cache is dirty it will be rebuilt in this func.
//
// Texture cache can be disabled with balance.DisableChunkTextureCache=true.
func (c *Chunk) Texture(e render.Engine) render.Texturer {
	if c.texture == nil || c.dirty {
		// Generate the normal bitmap and one with a color mask if applicable.
		bitmap := c.toBitmap(render.Invisible)
		if runtime.GOOS != "js" { // WASM
			defer os.Remove(bitmap)
		}
		tex, err := e.NewBitmap(bitmap)
		if err != nil {
			log.Error("Texture: %s", err)
		}

		c.texture = tex
		c.textureMasked = nil // invalidate until next call
		c.dirty = false
	}
	return c.texture
}

// TextureMasked returns a cached texture with the ColorMask applied.
func (c *Chunk) TextureMasked(e render.Engine, mask render.Color) render.Texturer {
	if c.textureMasked == nil || c.textureMaskedColor != mask {
		// Generate the normal bitmap and one with a color mask if applicable.
		bitmap := c.toBitmap(mask)
		if runtime.GOOS != "js" { // WASM
			defer os.Remove(bitmap)
		}
		tex, err := e.NewBitmap(bitmap)
		if err != nil {
			log.Error("Texture: %s", err)
		}

		c.textureMasked = tex
		c.textureMaskedColor = mask
	}
	return c.textureMasked
}

// toBitmap puts the texture in a well named bitmap path in the cache folder.
func (c *Chunk) toBitmap(mask render.Color) string {
	// Generate a unique filename for this chunk cache.
	var filename string
	if c.uuid == uuid.Nil {
		c.uuid = uuid.Must(uuid.NewV4())
	}
	filename = c.uuid.String()

	if mask != render.Invisible {
		filename += fmt.Sprintf("-%02x%02x%02x%02x",
			mask.Red, mask.Green, mask.Blue, mask.Alpha,
		)
	}

	// Get the temp bitmap image.
	bitmap := userdir.CacheFilename("chunk", filename+".bmp")
	err := c.ToBitmap(bitmap, mask)
	if err != nil {
		log.Error("Texture: %s", err)
	}

	return bitmap
}

// ToBitmap exports the chunk's pixels as a bitmap image.
func (c *Chunk) ToBitmap(filename string, mask render.Color) error {
	canvas := c.SizePositive()
	imgSize := image.Rectangle{
		Min: image.Point{},
		Max: image.Point{
			X: c.Size,
			Y: c.Size,
		},
	}

	if imgSize.Max.X == 0 {
		imgSize.Max.X = int(canvas.W)
	}
	if imgSize.Max.Y == 0 {
		imgSize.Max.Y = int(canvas.H)
	}

	img := image.NewRGBA(imgSize)

	// Blank out the pixels.
	for x := 0; x < img.Bounds().Max.X; x++ {
		for y := 0; y < img.Bounds().Max.Y; y++ {
			img.Set(x, y, balance.DebugChunkBitmapBackground.ToColor())
		}
	}

	// Pixel coordinate offset to map the Chunk World Position to the
	// smaller image boundaries.
	pointOffset := render.Point{
		X: int32(c.Point.X * int32(c.Size)),
		Y: int32(c.Point.Y * int32(c.Size)),
	}

	// Blot all the pixels onto it.
	for px := range c.Iter() {
		var color = px.Swatch.Color
		if mask != render.Invisible {
			color = mask
		}
		img.Set(
			int(px.X-pointOffset.X),
			int(px.Y-pointOffset.Y),
			color.ToColor(),
		)
	}

	// Write the image to bitmap file.
	return wasm.StoreBitmap(filename, img)
}

// Set proxies to the accessor and flags the texture as dirty.
func (c *Chunk) Set(p render.Point, sw *Swatch) error {
	c.dirty = true
	return c.Accessor.Set(p, sw)
}

// Delete proxies to the accessor and flags the texture as dirty.
func (c *Chunk) Delete(p render.Point) error {
	c.dirty = true
	return c.Accessor.Delete(p)
}

// Rect returns the bounding coordinates that the Chunk has pixels for.
func (c *Chunk) Rect() render.Rect {
	// Lowest and highest chunks.
	var (
		lowest  render.Point
		highest render.Point
	)

	for coord := range c.Iter() {
		if coord.X < lowest.X {
			lowest.X = coord.X
		}
		if coord.Y < lowest.Y {
			lowest.Y = coord.Y
		}

		if coord.X > highest.X {
			highest.X = coord.X
		}
		if coord.Y > highest.Y {
			highest.Y = coord.Y
		}
	}

	return render.Rect{
		X: lowest.X,
		Y: lowest.Y,
		W: highest.X,
		H: highest.Y,
	}
}

// SizePositive returns the Size anchored to 0,0 with only positive
// coordinates.
func (c *Chunk) SizePositive() render.Rect {
	S := c.Rect()
	return render.Rect{
		W: int32(math.Abs(float64(S.X))) + S.W,
		H: int32(math.Abs(float64(S.Y))) + S.H,
	}
}

// Usage returns the percent of free space vs. allocated pixels in the chunk.
func (c *Chunk) Usage(size int) float64 {
	return float64(c.Len()) / float64(size)
}

// MarshalJSON writes the chunk to JSON.
func (c *Chunk) MarshalJSON() ([]byte, error) {
	data, err := c.Accessor.MarshalJSON()
	if err != nil {
		return []byte{}, err
	}

	generic := &JSONChunk{
		Type: c.Type,
		Data: data,
	}
	b, err := json.Marshal(generic)
	return b, err
}

// UnmarshalJSON loads the chunk from JSON and uses the correct accessor to
// parse the inner details.
func (c *Chunk) UnmarshalJSON(b []byte) error {
	// Parse it generically so we can hand off the inner "data" object to the
	// right accessor for unmarshalling.
	generic := &JSONChunk{}
	err := json.Unmarshal(b, generic)
	if err != nil {
		return fmt.Errorf("Chunk.UnmarshalJSON: failed to unmarshal into generic JSONChunk type: %s", err)
	}

	switch c.Type {
	case MapType:
		c.Accessor = NewMapAccessor()
		return c.Accessor.UnmarshalJSON(generic.Data)
	default:
		return fmt.Errorf("Chunk.UnmarshalJSON: unsupported chunk type '%d'", c.Type)
	}
}

func (c *Chunk) EncodeMsgpack(enc *msgpack.Encoder) error {
	data := c.Accessor

	generic := &JSONChunk{
		Type:    c.Type,
		BinData: data,
	}

	return enc.Encode(generic)
}

func (c *Chunk) DecodeMsgpack(dec *msgpack.Decoder) error {
	generic := &JSONChunk{}
	err := dec.Decode(generic)
	if err != nil {
		return fmt.Errorf("Chunk.DecodeMsgpack: %s", err)
	}

	switch c.Type {
	case MapType:
		c.Accessor = generic.BinData.(MapAccessor)
	default:
		return fmt.Errorf("Chunk.DecodeMsgpack: unsupported chunk type '%d'", c.Type)
	}

	return nil
}

// // MarshalMsgpack writes the chunk to msgpack format.
// func (c *Chunk) MarshalMsgpack() ([]byte, error) {
// 	// data, err := c.Accessor.MarshalMsgpack()
// 	// if err != nil {
// 	// 	return []byte{}, err
// 	// }
// 	data := c.Accessor
//
// 	generic := &JSONChunk{
// 		Type:    c.Type,
// 		BinData: data,
// 	}
// 	b, err := msgpack.Marshal(generic)
// 	return b, err
// }
//
// // UnmarshalMsgpack loads the chunk from msgpack format.
// func (c *Chunk) UnmarshalMsgpack(b []byte) error {
// 	// Parse it generically so we can hand off the inner "data" object to the
// 	// right accessor for unmarshalling.
// 	generic := &JSONChunk{}
// 	err := msgpack.Unmarshal(b, generic)
// 	if err != nil {
// 		return fmt.Errorf("Chunk.UnmarshalMsgpack: failed to unmarshal into generic JSONChunk type: %s", err)
// 	}
//
// 	switch c.Type {
// 	case MapType:
// 		c.Accessor = NewMapAccessor()
// 		return c.Accessor.UnmarshalMsgpack(generic.Data)
// 	default:
// 		return fmt.Errorf("Chunk.UnmarshalMsgpack: unsupported chunk type '%d'", c.Type)
// 	}
// }
