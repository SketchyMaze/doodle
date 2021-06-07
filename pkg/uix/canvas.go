package uix

import (
	"fmt"
	"runtime"
	"strings"

	"git.kirsle.net/apps/doodle/pkg/balance"
	"git.kirsle.net/apps/doodle/pkg/collision"
	"git.kirsle.net/apps/doodle/pkg/doodads"
	"git.kirsle.net/apps/doodle/pkg/drawtool"
	"git.kirsle.net/apps/doodle/pkg/filesystem"
	"git.kirsle.net/apps/doodle/pkg/level"
	"git.kirsle.net/apps/doodle/pkg/log"
	"git.kirsle.net/apps/doodle/pkg/scripting"
	"git.kirsle.net/apps/doodle/pkg/wallpaper"
	"git.kirsle.net/go/render"
	"git.kirsle.net/go/render/event"
	"git.kirsle.net/go/ui"
)

// Canvas is a custom ui.Widget that manages a single drawing.
type Canvas struct {
	ui.Frame
	Palette *level.Palette

	// Editable and Scrollable go hand in hand and, if you initialize a
	// NewCanvas() with editable=true, they are both enabled.
	Editable   bool // Clicking will edit pixels of this canvas.
	Scrollable bool // Cursor keys will scroll the viewport of this canvas.
	Zoom       int  // Zoom level on the canvas.

	// Custom label to place in the lower-right corner of the canvas.
	// Used for e.g. the quantity badge on Inventory items.
	CornerLabel string

	// Selected draw tool/mode, default Pencil, for editable canvases.
	Tool      drawtool.Tool
	BrushSize int // thickness of selected brush

	// MaskColor will force every pixel to render as this color regardless of
	// the palette index of that pixel. Otherwise pixels behave the same and
	// the palette does work as normal. Set to render.Invisible (zero value)
	// to remove the mask.
	MaskColor render.Color

	// Actor ID to follow the camera on automatically, i.e. the main player.
	FollowActor string

	// Debug tools
	// NoLimitScroll suppresses the scroll limit for bounded levels.
	NoLimitScroll bool

	// Underlying chunk data for the drawing.
	level    *level.Level
	chunks   *level.Chunker
	modified bool // set to True when the drawing has been modified, like in Editor Mode.

	// Actors to superimpose on top of the drawing.
	actor  *Actor   // if this canvas IS an actor
	actors []*Actor // if this canvas CONTAINS actors (i.e., is a level)

	// Collision memory for the actors.
	collidingActors map[string]string // mapping their IDs to each other

	// Doodad scripting engine supervisor.
	// NOTE: initialized and managed by the play_scene.
	scripting *scripting.Supervisor

	// Wallpaper settings.
	wallpaper *Wallpaper

	// When the Canvas wants to delete Actors, but ultimately it is upstream
	// that controls the actors. Upstream should delete them and then reinstall
	// the actor list from scratch.
	OnDeleteActors func([]*level.Actor)
	OnDragStart    func(*level.Actor)

	// -- WHEN Canvas.Tool is "Link" --
	// When the Canvas wants to link two actors together. Arguments are the IDs
	// of the two actors.
	OnLinkActors func(a, b *Actor)
	linkFirst    *Actor

	// Collision handlers for level geometry.
	OnLevelCollision func(*Actor, *collision.Collide)

	/********
	 * Editable canvas private variables.
	 ********/
	// The current stroke actively being drawn by the user, during a
	// mousedown-and-dragging event.
	currentStroke *drawtool.Stroke
	strokes       map[int]*drawtool.Stroke // active stroke mapped by ID
	lastPixel     *level.Pixel

	// We inherit the ui.Widget which manages the width and height.
	Scroll render.Point // Scroll offset for which parts of canvas are visible.
}

// NewCanvas initializes a Canvas widget.
//
// If editable is true, Scrollable is also set to true, which means the arrow
// keys will scroll the canvas viewport which is desirable in Edit Mode.
func NewCanvas(size int, editable bool) *Canvas {
	w := &Canvas{
		Editable:   editable,
		Scrollable: editable,
		Palette:    level.NewPalette(),
		BrushSize:  1,
		chunks:     level.NewChunker(size),
		actors:     make([]*Actor, 0),
		wallpaper:  &Wallpaper{},

		strokes: map[int]*drawtool.Stroke{},
	}
	w.setup()
	w.IDFunc(func() string {
		var attrs []string

		if w.Editable {
			attrs = append(attrs, "editable")
		} else {
			attrs = append(attrs, "read-only")
		}

		if w.Scrollable {
			attrs = append(attrs, "scrollable")
		}

		return fmt.Sprintf("Canvas<%d; %s>", size, strings.Join(attrs, "; "))
	})
	return w
}

// Load initializes the Canvas using an existing Palette and Grid.
func (w *Canvas) Load(p *level.Palette, g *level.Chunker) {
	w.Palette = p
	w.chunks = g
	w.modified = false

	if len(w.Palette.Swatches) > 0 {
		w.SetSwatch(w.Palette.Swatches[0])
	}
}

// LoadLevel initializes a Canvas from a Level object.
func (w *Canvas) LoadLevel(e render.Engine, level *level.Level) {
	w.level = level
	w.Load(level.Palette, level.Chunker)

	// TODO: wallpaper paths
	filename := "assets/wallpapers/" + level.Wallpaper
	if runtime.GOOS != "js" {
		// Check if the wallpaper wasn't found. Check bindata and file system.
		if _, err := filesystem.FindFileEmbedded(filename, level); err != nil {
			log.Error("LoadLevel: wallpaper %s did not appear to exist, default to notebook.png", filename)
			filename = "assets/wallpapers/notebook.png"
		}
	}

	wp, err := wallpaper.FromFile(e, filename, level)
	if err != nil {
		log.Error("wallpaper FromFile(%s): %s", filename, err)
	}

	w.wallpaper.maxWidth = level.MaxWidth
	w.wallpaper.maxHeight = level.MaxHeight
	err = w.wallpaper.Load(e, level.PageType, wp)
	if err != nil {
		log.Error("wallpaper Load: %s", err)
	}
}

// LoadDoodad initializes a Canvas from a Doodad object.
func (w *Canvas) LoadDoodad(d *doodads.Doodad) {
	// TODO more safe
	w.Load(d.Palette, d.Layers[0].Chunker)
}

// LoadDoodadToLayer initializes a Canvas from a Doodad object and picks
// a layer to load.
func (w *Canvas) LoadDoodadToLayer(d *doodads.Doodad, index int) {
	if index < 0 || index > len(d.Layers) {
		log.Error("LoadDoodadToLayer: index %d out of range", index)
		return
	}
	w.Load(d.Palette, d.Layers[index].Chunker)
}

// SetSwatch changes the currently selected swatch for editing.
func (w *Canvas) SetSwatch(s *level.Swatch) {
	w.Palette.ActiveSwatch = s
}

// setup common configs between both initializers of the canvas.
func (w *Canvas) setup() {
	// XXX: Debug code.
	if balance.DebugCanvasBorder != render.Invisible {
		w.Configure(ui.Config{
			BorderColor: balance.DebugCanvasBorder,
			BorderSize:  2,
			BorderStyle: ui.BorderSolid,
		})
	}
}

// Loop is called on the scene's event loop to handle mouse interaction with
// the canvas, i.e. to edit it.
func (w *Canvas) Loop(ev *event.State) error {
	// Process the arrow keys scrolling the level in Edit Mode.
	// canvas_scrolling.go
	w.loopEditorScroll(ev)
	if err := w.loopFollowActor(ev); err != nil {
		log.Error("Follow actor: %s", err) // not fatal but nice to know
	}
	_ = w.loopConstrainScroll()

	// Current time of this loop so we can advance animations.
	// now := time.Now()

	// Remove any actors that were destroyed the previous tick.
	var newActors []*Actor
	for _, a := range w.actors {
		if a.flagDestroy {
			continue
		}
		newActors = append(newActors, a)
	}
	if len(newActors) < len(w.actors) {
		w.actors = newActors
	}

	// Check collisions between actors.
	if w.scripting != nil {
		if err := w.loopActorCollision(); err != nil {
			log.Error("loopActorCollision: %s", err)
		}
	}

	// If the canvas is editable, only care if it's over our space.
	if w.Editable {
		cursor := render.NewPoint(ev.CursorX, ev.CursorY)
		if cursor.Inside(ui.AbsoluteRect(w)) {
			return w.loopEditable(ev)
		}
	}
	return nil
}

// Viewport returns a rect containing the viewable drawing coordinates in this
// canvas. The X,Y values are the scroll offset (top left) and the W,H values
// are the scroll offset plus the width/height of the Canvas widget.
//
// The Viewport rect are the Absolute World Coordinates of the drawing that are
// visible inside the Canvas. The X,Y is the top left World Coordinate and the
// W,H are the bottom right World Coordinate, making this rect an absolute
// slice of the world. For a normal rect with a relative width and height,
// use ViewportRelative().
//
// The rect X,Y are the negative Scroll Value.
// The rect W,H are the Canvas widget size minus the Scroll Value.
func (w *Canvas) Viewport() render.Rect {
	var S = w.Size()
	return render.Rect{
		X: -w.Scroll.X,
		Y: -w.Scroll.Y,
		W: S.W - w.Scroll.X,
		H: S.H - w.Scroll.Y,
	}
}

// ViewportRelative returns a relative viewport where the Width and Height
// values are zero-relative: so you can use it with point.Inside(viewport)
// to see if a World Index point should be visible on screen.
//
// The rect X,Y are the negative Scroll Value
// The rect W,H are the Canvas widget size.
func (w *Canvas) ViewportRelative() render.Rect {
	var S = w.Size()
	return render.Rect{
		X: -w.Scroll.X,
		Y: -w.Scroll.Y,
		W: S.W,
		H: S.H,
	}
}

// WorldIndexAt returns the World Index that corresponds to a Screen Pixel
// on the screen. If the screen pixel is the mouse coordinate (relative to
// the application window) this will return the World Index of the pixel below
// the mouse cursor.
func (w *Canvas) WorldIndexAt(screenPixel render.Point) render.Point {
	var P = ui.AbsolutePosition(w)
	world := render.Point{
		X: screenPixel.X - P.X - w.Scroll.X,
		Y: screenPixel.Y - P.Y - w.Scroll.Y,
	}

	// Handle Zoomies
	if w.Zoom != 0 {
		world.X = w.ZoomMultiply(world.X)
		world.Y = w.ZoomMultiply(world.Y)
	}
	return world
}

// Chunker returns the underlying Chunker object.
func (w *Canvas) Chunker() *level.Chunker {
	return w.chunks
}

// ScrollTo sets the viewport scroll position.
func (w *Canvas) ScrollTo(to render.Point) {
	w.Scroll.X = to.X
	w.Scroll.Y = to.Y
}

// ScrollBy adjusts the viewport scroll position.
func (w *Canvas) ScrollBy(by render.Point) {
	w.Scroll.Add(by)
}

// Compute the canvas.
func (w *Canvas) Compute(e render.Engine) {

}
