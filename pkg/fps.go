package doodle

import (
	"fmt"
	"runtime"
	"strings"

	"git.kirsle.net/SketchyMaze/doodle/pkg/balance"
	"git.kirsle.net/SketchyMaze/doodle/pkg/collision"
	"git.kirsle.net/SketchyMaze/doodle/pkg/drawtool"
	"git.kirsle.net/SketchyMaze/doodle/pkg/native"
	"git.kirsle.net/SketchyMaze/doodle/pkg/uix"
	"git.kirsle.net/go/render"
	"git.kirsle.net/go/ui"
)

// Frames to cache for FPS calculation.
const maxSamples = 100

// Debug mode options, these can be enabled in the dev console
// like: boolProp DebugOverlay true
var (
	DebugOverlay   = false
	DebugCollision = false

	DebugTextPadding = 8
	DebugTextSize    = 24
	DebugTextColor   = render.SkyBlue
	DebugTextStroke  = render.Grey
	DebugTextShadow  = render.Black
)

var (
	fpsCurrentTicks uint32 // current time we get sdl.GetTicks()
	fpsLastTime     uint32 // last time we printed the fpsCurrentTicks
	fpsCurrent      int
	fpsFrames       int
	fpsSkipped      uint32
	fpsInterval     uint32 = 1000
	fpsDoNotCap     bool   // remove the FPS delay cap in main loop

	// Memory usage metrics.
	memHeap      uint64 // Current heap memory size (shrinks on GC)
	memSys       uint64 // Current total memory usage
	numGoroutine int

	// Custom labels for individual Scenes to add debug info.
	customDebugLabels []debugLabel
)

type debugLabel struct {
	key      string
	variable *string
}

// DrawDebugOverlay draws the debug FPS text on the SDL canvas.
func (d *Doodle) DrawDebugOverlay() {
	if !DebugOverlay {
		return
	}

	var framesSkipped = fmt.Sprintf("(skip: %dms)", fpsSkipped)
	if fpsDoNotCap {
		framesSkipped = "uncapped"
	}

	// Get the size of cached SDL2 textures at the render engine level.
	var texCount = native.CountTextures(d.Engine)

	var (
		darken  = balance.DebugStrokeDarken
		Yoffset = 20 // leave room for the menu bar
		Xoffset = 20
		keys    = []string{
			"FPS:",
			"Scene:",
			"Mouse:",
			"Textures:",
			"Sys/Heap:",
			"Threads:",
		}
		values = []string{
			fmt.Sprintf("%d   %s", fpsCurrent, framesSkipped),
			d.Scene.Name(),
			fmt.Sprintf("%d,%d", d.event.CursorX, d.event.CursorY),
			texCount,
			fmt.Sprintf("%d MiB / %d MiB", memSys, memHeap),
			fmt.Sprintf("%d", numGoroutine),
		}
	)

	// Insert custom keys.
	for _, custom := range customDebugLabels {
		keys = append(keys, custom.key)
		if custom.variable == nil {
			values = append(values, "<nil>")
		} else if len(*custom.variable) == 0 {
			values = append(values, `""`)
		} else {
			values = append(values, *custom.variable)
		}
	}

	// Find the longest key to align the labels up.
	var longest int
	for _, key := range keys {
		if len(key) > longest {
			longest = len(key)
		}
	}

	// Space pad the keys for alignment.
	for i, key := range keys {
		if len(key) < longest {
			key = strings.Repeat(" ", longest-len(key)) + key
			keys[i] = key
		}
	}

	key := ui.NewLabel(ui.Label{
		Text: strings.Join(keys, "\n"),
		Font: render.Text{
			Size:         balance.DebugFontSize,
			FontFilename: balance.ShellFontFilename,
			Color:        balance.DebugLabelColor,
			Stroke:       balance.DebugLabelColor.Darken(darken),
		},
	})
	key.Compute(d.Engine)
	key.Present(d.Engine, render.NewPoint(
		DebugTextPadding+Xoffset,
		DebugTextPadding+Yoffset,
	))

	value := ui.NewLabel(ui.Label{
		Text: strings.Join(values, "\n"),
		Font: render.Text{
			Size:         balance.DebugFontSize,
			FontFilename: balance.DebugFontFilename,
			Color:        balance.DebugValueColor,
			Stroke:       balance.DebugValueColor.Darken(darken),
		},
	})
	value.Compute(d.Engine)
	value.Present(d.Engine, render.NewPoint(
		DebugTextPadding+Xoffset+key.Size().W+DebugTextPadding,
		DebugTextPadding+Yoffset, // padding to not overlay menu bar
	))
}

// DrawCollisionBox draws the collision box around a Doodad.
//
// The canvas will be the level Canvas, and the collision box is drawn in world
// space using the canvas.DrawStrokes function.
func (d *Doodle) DrawCollisionBox(canvas *uix.Canvas, actor *uix.Actor) {
	if !DebugCollision {
		return
	}

	var (
		rect   = collision.GetBoundingRect(actor)
		box    = collision.GetCollisionBox(rect)
		hitbox = actor.Hitbox()
	)

	// Adjust the actor's bounding rect by its stated Hitbox from its script.
	rect = collision.SizePlusHitbox(rect, hitbox)

	box = collision.GetCollisionBox(rect)

	// The stroke data for drawing the collision box "inside" the level Canvas,
	// so it scrolls and works in world units not screen units.
	var strokes = []struct {
		Color  render.Color
		PointA render.Point
		PointB render.Point
	}{
		{render.DarkGreen, box.Top[0], box.Top[1]},
		{render.DarkBlue, box.Bottom[0], box.Bottom[1]},
		{render.DarkYellow, box.Left[0], box.Left[1]},
		{render.Red, box.Right[0], box.Right[1]},
	}
	for _, cfg := range strokes {
		stroke := drawtool.NewStroke(drawtool.Line, cfg.Color)
		stroke.PointA = cfg.PointA
		stroke.PointB = cfg.PointB
		canvas.DrawStrokes(d.Engine, []*drawtool.Stroke{stroke})
	}
}

// TrackFPS shows the current FPS once per second.
//
// In debug mode, changes the window title to include the FPS counter.
func (d *Doodle) TrackFPS(skipped uint32) {
	fpsFrames++
	fpsCurrentTicks = d.Engine.GetTicks()

	// Skip the first second.
	if fpsCurrentTicks < fpsInterval {
		return
	}

	if fpsLastTime < fpsCurrentTicks-fpsInterval {
		fpsLastTime = fpsCurrentTicks
		fpsCurrent = fpsFrames
		fpsFrames = 0
		fpsSkipped = skipped
	}

	// FPS in the title bar.
	d.Engine.SetTitle(fmt.Sprintf("%s (%d FPS)", d.Title(), fpsCurrent))

	// Refresh memory usage.
	GetMemUsage()
}

// GetMemUsage collects statistics on memory utilization.
func GetMemUsage() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	memHeap = bToMb(m.Alloc)
	memSys = bToMb(m.Sys)

	numGoroutine = runtime.NumGoroutine()
}

func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}
